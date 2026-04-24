package cors

import (
	"net/http"
	"strings"
)

// Middleware applies a minimal CORS policy suitable for local development.
//
// - allowedOrigins supports exact-match origins (e.g. "http://localhost:3000")
// - include "*" to allow any origin
// - Authorization + Content-Type headers are allowed for BasicAuth and JSON APIs
func Middleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowedAll := false
	allowedSet := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		o := strings.TrimSpace(origin)
		if o == "" {
			continue
		}
		if o == "*" {
			allowedAll = true
			continue
		}
		allowedSet[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if allowedAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					if _, ok := allowedSet[origin]; ok {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Add("Vary", "Origin")
					}
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "600")
			}

			// Short-circuit preflight.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
