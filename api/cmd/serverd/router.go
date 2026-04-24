package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	messageshandler "teammate/internal/handler/rest/messages"
	"teammate/internal/utils/auth"
)

func newRouter(handler messageshandler.Handler, authUsers map[string]string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Use(auth.Middleware(authUsers))
		handler.RegisterRoutes(r)
	})

	return r
}
