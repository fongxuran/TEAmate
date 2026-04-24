package rest

import "github.com/go-chi/chi/v5"

// Registrar is implemented by REST handlers that can mount their routes on a router.
//
// This keeps `cmd/serverd` wiring simple as new endpoints are added.
type Registrar interface {
	RegisterRoutes(r chi.Router)
}
