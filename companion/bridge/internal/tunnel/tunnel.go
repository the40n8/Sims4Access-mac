// Package tunnel runs a Cloudflare quick tunnel (cloudflared, no account)
// to the bridge, so the phone can reach it from any network. The address is
// random and changes each time the tunnel starts.
package tunnel

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/the40n8/simsight-companion/companion/bridge/internal/osutil"
)

var urlPattern = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

// Find returns cloudflared: next to the bridge first (bundled), then Homebrew.
func Find() string {
	for _, p := range []string{
		osutil.NextToMe(osutil.Exe("cloudflared")),
		"/opt/homebrew/bin/cloudflared",
		"/usr/local/bin/cloudflared",
	} {
		if osutil.Runnable(p) {
			return p
		}
	}
	return ""
}

// Tunnel keeps a quick tunnel to a local port running until ctx ends, or a
// named tunnel (fixed address) when created with NewNamed.
type Tunnel struct {
	bin, target string
	onReady     func(url string)
	// Named tunnel: token from the Cloudflare dashboard and its public URL.
	token, fixedURL string

	mu     sync.Mutex
	url    string
	status string // "starting", "ready", "unavailable"
}

func New(bin string, port int, onReady func(url string)) *Tunnel {
	return &Tunnel{bin: bin, target: "http://127.0.0.1:" + strconv.Itoa(port), onReady: onReady, status: "starting"}
}

// NewNamed runs a Cloudflare named tunnel: a fixed address set up once in the
// Cloudflare dashboard (its public hostname points at this Mac's bridge).
func NewNamed(bin, token, publicURL string, port int, onReady func(url string)) *Tunnel {
	return &Tunnel{bin: bin, token: token, fixedURL: strings.TrimRight(publicURL, "/"),
		target: "http://127.0.0.1:" + strconv.Itoa(port), onReady: onReady, status: "starting"}
}

// URL is the public address once it works ("" before).
func (t *Tunnel) URL() (url, status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.url, t.status
}

func (t *Tunnel) set(url, status string) {
	t.mu.Lock()
	t.url, t.status = url, status
	t.mu.Unlock()
}

// Run starts cloudflared and restarts it (with backoff) if it exits.
func (t *Tunnel) Run(ctx context.Context) {
	delay := 5 * time.Second
	for ctx.Err() == nil {
		started := time.Now()
		t.once(ctx)
		if ctx.Err() != nil {
			return
		}
		t.set("", "starting")
		if time.Since(started) > 5*time.Minute {
			delay = 5 * time.Second
		}
		slog.Warn("tunnel exited, restarting", "in", delay)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = min(delay*2, 5*time.Minute)
	}
}

func (t *Tunnel) once(ctx context.Context) {
	cmd := exec.CommandContext(ctx, t.bin, "tunnel", "--no-autoupdate", "--url", t.target)
	if t.token != "" {
		// The token goes in the environment, not the command line (visible in ps).
		// --url sends everything to the bridge, so the tunnel needs no ingress
		// rules of its own (a dashboard public hostname works too).
		cmd = exec.CommandContext(ctx, t.bin, "tunnel", "--no-autoupdate", "--url", t.target, "run")
		cmd.Env = append(os.Environ(), "TUNNEL_TOKEN="+t.token)
	}
	osutil.HideWindow(cmd)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.set("", "unavailable")
		return
	}
	if err := cmd.Start(); err != nil {
		slog.Error("tunnel start failed", "err", err)
		t.set("", "unavailable")
		return
	}
	slog.Info("tunnel starting", "cloudflared", t.bin)
	found := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(stderr)
		sent := false
		for sc.Scan() {
			line := sc.Text()
			m := urlPattern.FindString(line)
			if t.fixedURL != "" && strings.Contains(line, "Registered tunnel connection") {
				m = t.fixedURL
			}
			if m != "" && !sent {
				found <- m
				sent = true
			}
		}
		_, _ = io.Copy(io.Discard, stderr)
	}()
	select {
	case url := <-found:
		go t.publish(ctx, url)
	case <-time.After(2 * time.Minute):
		slog.Error("tunnel gave no address")
		_ = cmd.Process.Kill()
	case <-ctx.Done():
	}
	_ = cmd.Wait()
}

// publish waits until the new name resolves on public DNS and answers
// through Cloudflare: a phone that looks it up earlier caches "not found"
// for minutes.
func (t *Tunnel) publish(ctx context.Context, url string) {
	host := url[len("https://"):]
	resolver := &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, network, "1.1.1.1:53")
	}}
	for i := 0; i < 90 && ctx.Err() == nil; i++ {
		if ips, err := resolver.LookupIPAddr(ctx, host); err == nil && len(ips) > 0 && reachable(ctx, url, host, ips[0].IP.String()) {
			time.Sleep(3 * time.Second)
			t.set(url, "ready")
			slog.Info("tunnel ready", "host", host)
			if t.onReady != nil {
				t.onReady(url)
			}
			return
		}
		time.Sleep(2 * time.Second)
	}
	slog.Warn("tunnel name not published yet; using it anyway", "host", host)
	t.set(url, "ready")
	if t.onReady != nil {
		t.onReady(url)
	}
}

func reachable(ctx context.Context, url, host, ip string) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ServerName: host},
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, net.JoinHostPort(ip, "443"))
		},
	}
	defer tr.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/", nil)
	if err != nil {
		return false
	}
	res, err := (&http.Client{Transport: tr, Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return false
	}
	res.Body.Close()
	return res.StatusCode == http.StatusOK
}
