package auth

import "net/http"

const defaultRealm = "api"

// Middleware enforces HTTP Basic Authentication with the provided users.
func Middleware(users map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok || users[username] != password {
				w.Header().Set("WWW-Authenticate", "Basic realm=\""+defaultRealm+"\"")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
