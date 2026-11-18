package server

import (
	"crypto/subtle"
	"encoding/json"
	"io/fs"
	"net/http"
)

// ManifestCookie carries the pairing key to the Home Screen install: iOS
// gives an installed web app its own storage, so the key Safari kept would
// be missing there. Set by the paired page itself (lib/session.ts).
const ManifestCookie = "ts4k"

// Manifest serves the web app manifest. For a phone that is already paired
// (it sends the key in the cookie), the start address carries the key, so
// the app added to the Home Screen opens paired; anyone else gets the plain
// manifest.
func Manifest(web fs.FS, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := fs.ReadFile(web, "manifest.json")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/manifest+json")
		c, err := r.Cookie(ManifestCookie)
		if err != nil || subtle.ConstantTimeCompare([]byte(c.Value), []byte(token)) != 1 {
			_, _ = w.Write(b)
			return
		}
		var m map[string]any
		if json.Unmarshal(b, &m) != nil {
			_, _ = w.Write(b)
			return
		}
		m["start_url"] = "./#k=" + token
		_ = json.NewEncoder(w).Encode(m)
	})
}
