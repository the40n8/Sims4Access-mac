// Package speaker runs the ts4access-say companion (macOS voices) and feeds
// it speech lines in the TS4 Access fallback-voice format.
package speaker

import (
	"bufio"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1"
)

type Speaker struct {
	path   string
	onKey  func(name string)
	mu     sync.Mutex
	voice  *v1.Voice
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	closed bool
	// Bare keys to take while the game is in front ("G vk,vk+shift"),
	// sent again whenever the speaker restarts.
	grab string
}

func New(path string, voice *v1.Voice) *Speaker {
	return &Speaker{path: path, voice: voice}
}

// Listen starts the speaker now and keeps it running, with its keyboard
// shortcut (Option+Command+S) reported to onKey("voice"). The shortcut must
// work while speech is off, so the speaker does not wait for a first line.
func (s *Speaker) Listen(onKey func(name string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onKey = onKey
	if err := s.startLocked(); err != nil {
		slog.Error("speaker start failed", "err", err)
	}
}

func (s *Speaker) args() []string {
	v := s.voice
	if v == nil {
		if s.onKey != nil {
			return []string{"--hotkey"}
		}
		return nil
	}
	args := []string{"--rate", strconv.Itoa(int(v.GetRate())), "--volume", strconv.Itoa(int(v.GetVolume()))}
	if s.onKey != nil {
		args = append(args, "--hotkey")
	}
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
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.cmd, s.stdin = cmd, stdin
	if s.grab != "" {
		_, _ = io.WriteString(stdin, s.grab+"\n")
	}
	onKey := s.onKey
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			if name, ok := strings.CutPrefix(sc.Text(), "K "); ok && onKey != nil {
				onKey(name)
			}
		}
		_ = cmd.Wait()
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.cmd != cmd {
			return // replaced or stopped on purpose
		}
		s.cmd, s.stdin = nil, nil
		if s.onKey != nil && !s.closed {
			// Crashed: bring the shortcut back.
			slog.Warn("speaker exited, restarting")
			time.AfterFunc(2*time.Second, func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				if !s.closed {
					_ = s.startLocked()
				}
			})
		}
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

// Grab tells the speaker which bare keys the mod owns now.
func (s *Speaker) Grab(keys []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grab = "G " + strings.Join(keys, ",")
	if s.stdin != nil {
		if _, err := io.WriteString(s.stdin, s.grab+"\n"); err != nil {
			slog.Warn("speaker grab failed", "err", err)
		}
	}
}

// SetVoice applies new voice settings by restarting the speaker.
func (s *Speaker) SetVoice(v *v1.Voice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.voice = v
	s.stopLocked()
	if s.onKey != nil {
		_ = s.startLocked()
	}
}

func (s *Speaker) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.stopLocked()
}
