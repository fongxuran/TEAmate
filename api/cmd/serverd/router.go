package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"teammate/internal/handler/rest"
	"teammate/internal/realtime"
	"teammate/internal/utils/auth"
	"teammate/internal/utils/cors"
)

func newRouter(authUsers map[string]string, corsAllowedOrigins []string, hub *realtime.Hub, registrars ...rest.Registrar) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Middleware(corsAllowedOrigins))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Browser realtime endpoint (not under /api) so WebSocket connections don't need custom headers.
	if hub != nil {
		r.Get("/ws", hub.ServeWS)
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		if len(authUsers) > 0 {
			r.Use(auth.Middleware(authUsers))
		}
		for _, registrar := range registrars {
			if registrar == nil {
				continue
			}
			registrar.RegisterRoutes(r)
		}
	})

	return r
}
