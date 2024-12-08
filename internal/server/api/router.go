// Package api provides the routing and HTTP API endpoints for the application.
// It defines the routes and integrates middleware for authentication and request handling.
package api

import (
	"net/http"

	"github.com/adettelle/go-keeper/pkg/mware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handlers *CustomerHandlers, cardHandlers *CardHandlers, passHandlers *PassHandlers,
	fileHandlers *FileHandlers, jwtChecker mware.JwtChecker) chi.Router {

	r := chi.NewRouter()

	// withAuth wraps a given HTTP handler with authentication middleware.
	withAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return mware.AuthMwr(h, handlers.SignKey, jwtChecker)
	}

	// User authentication routes
	r.Post("/api/user/register", handlers.RegisterCustomer)
	r.Post("/api/user/login", handlers.Login)

	// Password management routes
	r.Put("/api/user/password", withAuth(passHandlers.PasswordCreate))
	r.Get("/api/user/passwords", withAuth(passHandlers.AllPasswords))
	r.Get("/api/user/password/{title}", withAuth(passHandlers.PasswordByTitle))
	r.Post("/api/user/password/update/{title}", withAuth(passHandlers.PasswordUpdate))
	r.Delete("/api/user/password/{title}", withAuth(passHandlers.PasswordDelete))

	// File management routes
	r.Put("/api/user/file", withAuth(fileHandlers.FileAdd))
	r.Get("/api/user/files", withAuth(fileHandlers.AllFiles))
	r.Get("/api/user/file/{title}", withAuth(fileHandlers.FileGetByTitle))
	r.Post("/api/user/file/update/{title}", withAuth(fileHandlers.FileUpdate))
	r.Delete("/api/user/file/{title}", withAuth(fileHandlers.FileDeleteByTitle))

	// Card management routes
	r.Put("/api/user/card", withAuth(cardHandlers.CardAdd))
	r.Get("/api/user/cards", withAuth(cardHandlers.AllCards))
	r.Get("/api/user/card/{title}", withAuth(cardHandlers.CardGetByTitle))
	r.Post("/api/user/card/update/{title}", withAuth(cardHandlers.CardUpdate))
	r.Delete("/api/user/card/{title}", withAuth(cardHandlers.CardDeleteByTitle))

	return r
}
