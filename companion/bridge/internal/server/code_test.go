package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCodePairsOnceAndExpires(t *testing.T) {
	c := NewCodes("secret-token-0123456789")
	now := time.Unix(1000, 0)
	c.now = func() time.Time { return now }
	var said []string
	c.SetSay(func(code string) { said = append(said, code) })

	code := c.Show(true)
	if len(code) != 6 || len(said) != 1 {
		t.Fatalf("code %q said %v", code, said)
	}
	if c.Show(false) != code || len(said) != 1 {
		t.Fatal("a refresh must keep the code and stay quiet")
	}
	c.Show(true)
	if len(said) != 2 {
		t.Fatal("opening the page again says the code again")
	}

	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/pair-code", strings.NewReader(body)))
		return w
	}
	if w := post(`{"code":"` + SpokenCode(code) + `"}`); w.Code != 200 || !strings.Contains(w.Body.String(), "secret-token") {
		t.Fatalf("right code (with spaces): %d %s", w.Code, w.Body)
	}
	if w := post(`{"code":"` + code + `"}`); w.Code != http.StatusForbidden {
		t.Fatalf("code reused: %d", w.Code)
	}

	code = c.Show(false)
	now = now.Add(CodeTTL)
	if ok, _ := c.Check(code); ok {
		t.Fatal("expired code accepted")
	}
}

func TestCodeLocksAfterWrongTries(t *testing.T) {
	c := NewCodes("secret-token-0123456789")
	code := c.Show(false)
	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}
	for i := 0; i < CodeMaxTries; i++ {
		c.Check(wrong)
	}
	if ok, tooMany := c.Check(code); ok || !tooMany {
		t.Fatalf("after %d wrong tries: ok=%v tooMany=%v", CodeMaxTries, ok, tooMany)
	}
	if c.Show(false) == "" {
		t.Fatal("a new code is made when the page opens again")
	}
}
