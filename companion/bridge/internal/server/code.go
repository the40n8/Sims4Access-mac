package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Pairing by code: when the phone link is shown on the Mac, the Mac also says
// a 6-digit code; typed in the phone app, it gives the app the pairing key.
// No camera, no copy and paste. A code is made on this Mac only, works once,
// for a few minutes, and only for a few wrong tries.
const (
	CodeTTL      = 5 * time.Minute
	CodeMaxTries = 5
)

type Codes struct {
	token string
	now   func() time.Time

	mu      sync.Mutex
	code    string
	expires time.Time
	tries   int
	say     func(code string)
}

func NewCodes(token string) *Codes {
	return &Codes{token: token, now: time.Now}
}

// SetSay sets how a code is announced on the Mac (the Mac voice).
func (c *Codes) SetSay(say func(code string)) {
	c.mu.Lock()
	c.say = say
	c.mu.Unlock()
}

// Show returns the current code, making a new one when there is none; it is
// announced when new, or when announce is true (the page opened again).
func (c *Codes) Show(announce bool) string {
	c.mu.Lock()
	fresh := false
	if c.code == "" || !c.now().Before(c.expires) || c.tries >= CodeMaxTries {
		n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
		if err != nil {
			c.mu.Unlock()
			return ""
		}
		c.code = fmt.Sprintf("%06d", n.Int64())
		c.expires = c.now().Add(CodeTTL)
		c.tries = 0
		fresh = true
	}
	code, say := c.code, c.say
	c.mu.Unlock()
	if say != nil && (fresh || announce) {
		say(code)
	}
	return code
}

// Check uses up the code when it matches; wrong tries count against it.
// ok is false with tooMany true once the code is used up by wrong tries.
func (c *Codes) Check(code string) (ok, tooMany bool) {
	code = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, code)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.code == "" || !c.now().Before(c.expires) {
		return false, false
	}
	if c.tries >= CodeMaxTries {
		return false, true
	}
	if len(code) == len(c.code) && subtle.ConstantTimeCompare([]byte(code), []byte(c.code)) == 1 {
		c.code = "" // once only
		return true, false
	}
	c.tries++
	return false, c.tries >= CodeMaxTries
}

// Handler answers POST {"code":"123456"} with {"token":"..."}; 403 for a wrong
// or expired code, 429 once there were too many wrong tries.
func (c *Codes) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256)).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ok, tooMany := c.Check(body.Code)
		switch {
		case ok:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": c.token})
		case tooMany:
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			time.Sleep(300 * time.Millisecond) // slows guessing a little more
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

// Spoken digit by digit, so the Mac voice does not read "four hundred…".
func SpokenCode(code string) string {
	return strings.Join(strings.Split(code, ""), " ")
}
