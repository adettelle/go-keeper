package api

import "github.com/go-chi/chi/v5"

func NewRouter(userRepo *UserRepo) chi.Router {
	r := chi.NewRouter()

	r.Post("/api/user/register", userRepo.RegisterCustomer)

	return r
}
