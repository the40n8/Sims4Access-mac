// Package server exposes the companion API (Connect) and the web app to
// phones on the local network.
package server

import (
	"context"
	"crypto/subtle"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	v1 "github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1/ts4accessv1connect"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/hub"
)

type service struct {
	hub *hub.Hub
}

var _ ts4accessv1connect.CompanionServiceHandler = (*service)(nil)

func (s *service) Watch(ctx context.Context, _ *connect.Request[v1.WatchRequest], stream *connect.ServerStream[v1.WatchResponse]) error {
	for msg := range s.hub.Subscribe(ctx) {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) SetSpeed(ctx context.Context, req *connect.Request[v1.SetSpeedRequest]) (*connect.Response[v1.SetSpeedResponse], error) {
	switch req.Msg.GetSpeed() {
	case v1.Speed_SPEED_PAUSED, v1.Speed_SPEED_NORMAL, v1.Speed_SPEED_FAST, v1.Speed_SPEED_ULTRA:
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("speed must be paused, normal, fast or ultra"))
	}
	out := s.hub.Command(ctx, &v1.ToGame{Command: &v1.ToGame_SetSpeed{SetSpeed: req.Msg}})
	return connect.NewResponse(&v1.SetSpeedResponse{Outcome: out}), nil
}

func (s *service) SelectSim(ctx context.Context, req *connect.Request[v1.SelectSimRequest]) (*connect.Response[v1.SelectSimResponse], error) {
	out := s.hub.Command(ctx, &v1.ToGame{Command: &v1.ToGame_SelectSim{SelectSim: req.Msg}})
	return connect.NewResponse(&v1.SelectSimResponse{Outcome: out}), nil
}

func (s *service) CancelAction(ctx context.Context, req *connect.Request[v1.CancelActionRequest]) (*connect.Response[v1.CancelActionResponse], error) {
	out := s.hub.Command(ctx, &v1.ToGame{Command: &v1.ToGame_CancelAction{CancelAction: req.Msg}})
	return connect.NewResponse(&v1.CancelActionResponse{Outcome: out}), nil
}

func (s *service) SetNarration(ctx context.Context, req *connect.Request[v1.SetNarrationRequest]) (*connect.Response[v1.SetNarrationResponse], error) {
	if req.Msg.GetNarration() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("narration is required"))
	}
	out := s.hub.Command(ctx, &v1.ToGame{Command: &v1.ToGame_SetNarration{SetNarration: req.Msg}})
	return connect.NewResponse(&v1.SetNarrationResponse{Outcome: out}), nil
}

// authInterceptor accepts only requests carrying the pairing token.
func authInterceptor(token string) connect.Interceptor {
	check := func(h http.Header) error {
		got := strings.TrimPrefix(h.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			return connect.NewError(connect.CodeUnauthenticated, errors.New("pair this phone again with the link from the Mac"))
		}
		return nil
	}
	return &auth{check: check}
}

type auth struct{ check func(http.Header) error }

func (a *auth) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if err := a.check(req.Header()); err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func (a *auth) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (a *auth) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if err := a.check(conn.RequestHeader()); err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

// Handler serves the API and the web app (an SPA: unknown paths get index.html).
func Handler(h *hub.Hub, token string, web fs.FS) http.Handler {
	mux := http.NewServeMux()
	path, api := ts4accessv1connect.NewCompanionServiceHandler(&service{hub: h},
		connect.WithInterceptors(authInterceptor(token)))
	mux.Handle(path, api)

	files := http.FileServerFS(web)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "no-referrer")
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if _, err := fs.Stat(web, name); err == nil {
				if strings.HasPrefix(name, "_app/immutable/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, web, "index.html")
	})
	return mux
}
