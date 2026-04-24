package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	messageshandler "teammate/internal/handler/rest/messages"
	"teammate/internal/utils/auth"
	"teammate/internal/utils/cors"
)

func newRouter(handler messageshandler.Handler, authUsers map[string]string, corsAllowedOrigins []string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Middleware(corsAllowedOrigins))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		if len(authUsers) > 0 {
			r.Use(auth.Middleware(authUsers))
		}
		handler.RegisterRoutes(r)
	})

	return r
}
