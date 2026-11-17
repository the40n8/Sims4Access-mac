// Command ts4access-bridge connects The Sims 4 (TS4 Access mod) to phones.
//
// The mod starts it with pipes: speech and game state arrive on stdin,
// phone commands leave on stdout. It serves the phone app on the local
// network and exits when the game closes the pipe.
//
//	ts4access-bridge --speaker /path/ts4access-say [--rate 3 --volume 70 --voice Thomas --lang fr-FR]
//	ts4access-bridge --demo   # fake game, for developing the phone app
//	ts4access-bridge link     # print the phone link and its QR code
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	v1 "github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/demo"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/hub"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/pairing"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/pipe"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/server"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/speaker"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/webui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "link" {
		if err := printLink(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := run(); err != nil {
		slog.Error("bridge stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	home, _ := os.UserHomeDir()
	var (
		speakerPath = flag.String("speaker", "", "path to ts4access-say (default: next to this binary)")
		rate        = flag.Int("rate", 3, "voice rate -10..10")
		volume      = flag.Int("volume", 70, "voice volume 0..100")
		voiceName   = flag.String("voice", "", "part of a macOS voice name")
		lang        = flag.String("lang", "", "voice language, e.g. fr-FR")
		dataDir     = flag.String("data", filepath.Join(home, "Library", "Application Support", "TS4AccessCompanion"), "pairing and log folder")
		webDir      = flag.String("web", "", "serve the phone app from this folder instead of the embedded build (development)")
		port        = flag.Int("port", 0, "listen port (default: saved port, 47800)")
		demoMode    = flag.Bool("demo", false, "run a fake game instead of reading the mod's pipe")
	)
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0o700); err != nil {
		return err
	}
	// stdout is the pipe to the game: logs go to a file (and stderr in demo).
	logFile, err := os.OpenFile(filepath.Join(*dataDir, "bridge.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	var logOut = logFile
	if *demoMode {
		logOut = os.Stderr
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(logOut, nil)))

	cfg, err := pairing.Load(*dataDir)
	if err != nil {
		return err
	}
	if *port != 0 {
		cfg.Port = *port
	}
	linkFile, err := cfg.WriteLinkFile(*dataDir)
	if err != nil {
		return err
	}
	slog.Info("phone link written", "file", linkFile)
	if *demoMode {
		for _, l := range cfg.Links() {
			fmt.Fprintln(os.Stderr, "phone link:", l)
		}
	}

	var web fs.FS = webui.FS()
	if *webDir != "" {
		web = os.DirFS(*webDir)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	h := hub.New(nil)
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))
	if err != nil {
		// Another game session or bridge holds the port: keep speech working.
		slog.Error("phone server unavailable", "port", cfg.Port, "err", err)
	} else {
		srv := &http.Server{
			Handler:           server.Handler(h, cfg.Token, web),
			ReadHeaderTimeout: 10 * time.Second,
		}
		go func() {
			if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("phone server stopped", "err", err)
			}
		}()
		defer func() {
			sctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = srv.Shutdown(sctx)
		}()
		slog.Info("phone server listening", "port", cfg.Port)
	}

	if *demoMode {
		demo.Run(ctx, h)
		return nil
	}

	if *speakerPath == "" {
		exe, _ := os.Executable()
		*speakerPath = filepath.Join(filepath.Dir(exe), "ts4access-say")
	}
	spk := speaker.New(*speakerPath, &v1.Voice{Rate: int32(*rate), Volume: int32(*volume), Name: *voiceName, Lang: *lang})
	defer spk.Close()

	p := pipe.New(os.Stdin, os.Stdout)
	h.SetSender(p.Send)
	h.SetConnected(true)
	defer h.SetConnected(false)

	done := make(chan error, 1)
	go func() {
		done <- p.Run(pipe.Handlers{
			Speech: func(line string) {
				spk.Line(line)
				if text := pipe.SpeechText(line); text != "" {
					h.AddEvent(&v1.Event{Kind: v1.EventKind_EVENT_KIND_SPOKEN, Subject: text})
				}
			},
			Hello:  func(m *v1.Hello) { slog.Info("game connected", "mod", m.GetModVersion(), "game", m.GetGameVersion()) },
			State:  h.SetState,
			Event:  h.AddEvent,
			Result: h.Resolve,
			Voice:  spk.SetVoice,
		})
	}()
	select {
	case err := <-done:
		slog.Info("game closed the pipe", "err", err)
		return nil
	case <-ctx.Done():
		return nil
	}
}

// printLink shows the phone link as text and as a QR code to scan.
func printLink(args []string) error {
	home, _ := os.UserHomeDir()
	fs := flag.NewFlagSet("link", flag.ExitOnError)
	dataDir := fs.String("data", filepath.Join(home, "Library", "Application Support", "TS4AccessCompanion"), "pairing folder")
	all := fs.Bool("all", false, "list every address")
	invert := fs.Bool("invert", false, "QR for terminals with a light background")
	port := fs.Int("port", 0, "link to a different port (e.g. 5173 for the live-reload dev server)")
	_ = fs.Parse(args)

	cfg, err := pairing.Load(*dataDir)
	if err != nil {
		return err
	}
	if *port != 0 {
		cfg.Port = *port
	}
	links := cfg.Links()
	if len(links) == 0 {
		return errors.New("no network address: is the Mac on Wi-Fi?")
	}
	code, err := pairing.TerminalQR(links[0])
	if err != nil {
		return err
	}
	if *invert {
		code = strings.NewReplacer("█", " ", " ", "█", "▀", "▄", "▄", "▀").Replace(code)
	}
	fmt.Print(code)
	fmt.Println(links[0])
	if *all {
		fmt.Println("\nOther addresses, if that one does not open:")
		for _, l := range links[1:] {
			fmt.Println(l)
		}
	}
	return nil
}
