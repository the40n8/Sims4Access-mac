// Package speaker runs the ts4access-say companion (macOS voices) and feeds
// it speech lines in the TS4 Access fallback-voice format.
package speaker

import (
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"sync"

	v1 "github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1"
)

type Speaker struct {
	path  string
	mu    sync.Mutex
	voice *v1.Voice
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func New(path string, voice *v1.Voice) *Speaker {
	return &Speaker{path: path, voice: voice}
}

func (s *Speaker) args() []string {
	v := s.voice
	if v == nil {
		return nil
	}
	args := []string{"--rate", strconv.Itoa(int(v.GetRate())), "--volume", strconv.Itoa(int(v.GetVolume()))}
	if v.GetName() != "" {
		args = append(args, "--voice", v.GetName())
	}
	if v.GetLang() != "" {
		args = append(args, "--lang", v.GetLang())
	}
	return args
}

// startLocked launches the speaker if it is not running. Caller holds s.mu.
func (s *Speaker) startLocked() error {
	if s.cmd != nil {
		return nil
	}
	cmd := exec.Command(s.path, s.args()...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.cmd, s.stdin = cmd, stdin
	go func() {
		_ = cmd.Wait()
		s.mu.Lock()
		if s.cmd == cmd {
			s.cmd, s.stdin = nil, nil
		}
		s.mu.Unlock()
	}()
	slog.Info("speaker started", "args", s.args())
	return nil
}

func (s *Speaker) stopLocked() {
	if s.stdin != nil {
		_ = s.stdin.Close() // the speaker exits on EOF
	}
	s.cmd, s.stdin = nil, nil
}

// Line forwards one speech line ("I<text>", "Q<text>" or "I").
func (s *Speaker) Line(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.startLocked(); err != nil {
		slog.Error("speaker start failed", "err", err)
		return
	}
	if _, err := io.WriteString(s.stdin, line+"\n"); err != nil {
		slog.Warn("speaker write failed, restarting next line", "err", err)
		s.stopLocked()
	}
}

// SetVoice applies new voice settings; the speaker restarts on next line.
func (s *Speaker) SetVoice(v *v1.Voice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.voice = v
	s.stopLocked()
}

func (s *Speaker) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()
}
