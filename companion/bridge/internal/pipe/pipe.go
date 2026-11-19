// Package pipe speaks the game <-> bridge protocol on stdin/stdout: one
// message per line, protobuf JSON, plus legacy speech lines ("I..."/"Q...").
package pipe

import (
	"bufio"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"sync"

	v1 "github.com/the40n8/simsight-companion/companion/bridge/gen/ts4access/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// maxLine bounds one message (a full household state is a few tens of KB).
const maxLine = 4 << 20

var (
	unmarshal = protojson.UnmarshalOptions{DiscardUnknown: true}
	marshal   = protojson.MarshalOptions{}
)

// Handlers receive what the game sends. They are called from Run's goroutine.
type Handlers struct {
	Speech func(line string)
	Hello  func(*v1.Hello)
	State  func(*v1.GameState)
	Event  func(*v1.Event)
	Result func(*v1.CommandResult)
	Voice  func(*v1.Voice)
	Grab   func(*v1.GrabKeys)
}

type Pipe struct {
	in  io.Reader
	mu  sync.Mutex
	out io.Writer
}

func New(in io.Reader, out io.Writer) *Pipe {
	return &Pipe{in: in, out: out}
}

// Send writes one command line to the game.
func (p *Pipe) Send(cmd *v1.ToGame) error {
	b, err := marshal.Marshal(cmd)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	_, err = p.out.Write(append(b, '\n'))
	return err
}

// Run reads until the game closes the pipe (game quit or crashed).
func (p *Pipe) Run(h Handlers) error {
	sc := bufio.NewScanner(p.in)
	sc.Buffer(make([]byte, 64<<10), maxLine)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if line == "" {
			continue
		}
		if line[0] != '{' {
			if h.Speech != nil {
				h.Speech(line)
			}
			continue
		}
		var msg v1.ToBridge
		if err := unmarshal.Unmarshal([]byte(line), &msg); err != nil {
			slog.Warn("bad line from game", "err", err, "head", head(line))
			continue
		}
		switch m := msg.Payload.(type) {
		case *v1.ToBridge_Hello:
			call(h.Hello, m.Hello)
		case *v1.ToBridge_State:
			call(h.State, m.State)
		case *v1.ToBridge_Event:
			call(h.Event, m.Event)
		case *v1.ToBridge_Result:
			call(h.Result, m.Result)
		case *v1.ToBridge_Voice:
			call(h.Voice, m.Voice)
		case *v1.ToBridge_Grab:
			call(h.Grab, m.Grab)
		}
	}
	return sc.Err()
}

func call[T any](fn func(T), v T) {
	if fn != nil {
		fn(v)
	}
}

func head(s string) string {
	if len(s) > 80 {
		return s[:80]
	}
	return s
}

// SpeechText decodes a speech line's text ("" for a bare stop line).
func SpeechText(line string) string {
	if len(line) < 2 {
		return ""
	}
	text, err := url.PathUnescape(line[1:])
	if err != nil {
		return line[1:]
	}
	return text
}
