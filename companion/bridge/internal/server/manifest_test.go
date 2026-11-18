package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestManifestCarriesKeyOnlyForPairedPhone(t *testing.T) {
	web := fstest.MapFS{"manifest.json": {Data: []byte(`{"name":"Sims","start_url":"./"}`)}}
	h := Manifest(web, "secret-token-0123456789")
	get := func(cookie string) string {
		r := httptest.NewRequest(http.MethodGet, "/manifest.json", nil)
		if cookie != "" {
			r.AddCookie(&http.Cookie{Name: ManifestCookie, Value: cookie})
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Body.String()
	}
	if strings.Contains(get(""), "secret") || strings.Contains(get("wrong"), "secret") {
		t.Fatal("key leaked to an unpaired request")
	}
	if !strings.Contains(get("secret-token-0123456789"), `"start_url":"./#k=secret-token-0123456789"`) {
		t.Fatalf("paired phone: got %s", get("secret-token-0123456789"))
	}
}
