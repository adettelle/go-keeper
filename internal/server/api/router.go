package api

import (
	"net/http"

	"github.com/adettelle/go-keeper/pkg/mware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handlers *CustomerHandlers, jwtChecker mware.JwtChecker) chi.Router {
	r := chi.NewRouter()

	withAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return mware.AuthMwr(h, handlers.JwtSignKey, jwtChecker)
	}

	r.Post("/api/user/register", handlers.RegisterCustomer)
	r.Post("/api/user/login", handlers.Login)

	r.Put("/api/user/password", withAuth(handlers.PasswordCreate))
	r.Get("/api/user/passwords", withAuth(handlers.AllPasswords))
	r.Get("/api/user/password/{title}", withAuth(handlers.PasswordByTitle))
	r.Post("/api/user/password/update/{title}", withAuth(handlers.PasswordUpdate))
	r.Delete("/api/user/password/{title}", withAuth(handlers.PasswordDelete))

	r.Put("/api/user/file", withAuth(handlers.FileAdd))
	r.Get("/api/user/files", withAuth(handlers.AllFiles))
	r.Get("/api/user/file/{title}", withAuth(handlers.FileGetByTitle))
	r.Post("/api/user/file/update/{title}", withAuth(handlers.FileUpdate))
	r.Delete("/api/user/file/{title}", withAuth(handlers.FileDeleteByTitle))

	r.Put("/api/user/card", withAuth(handlers.CardAdd))
	r.Get("/api/user/cards", withAuth(handlers.AllCards))
	r.Get("/api/user/card/{title}", withAuth(handlers.CardGetByTitle))
	r.Post("/api/user/card/update/{title}", withAuth(handlers.CardUpdate))
	r.Delete("/api/user/card/{title}", withAuth(handlers.CardDeleteByTitle))

	return r
}
