package api

import (
	"github.com/adettelle/go-keeper/pkg/mware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(userRepo *CustomerHandlers) chi.Router { // как переименовать userRepo???????
	r := chi.NewRouter()

	r.Post("/api/user/register", userRepo.RegisterCustomer)
	r.Post("/api/user/login", userRepo.Login) // TODO Check

	// TODO middleware на авторизацию, на проверку jwt, всё кроме логина
	r.Get("/api/user/passwords", mware.AuthMwr(userRepo.AllPasswords, userRepo.JwtSignKey))
	r.Get("/api/user/password/{title}", mware.AuthMwr(userRepo.PasswordByTitle, userRepo.JwtSignKey))
	r.Put("/api/user/password", mware.AuthMwr(userRepo.PasswordCreate, userRepo.JwtSignKey))
	r.Post("/api/user/password/update", mware.AuthMwr(userRepo.PasswordUpdate, userRepo.JwtSignKey))
	r.Delete("/api/user/delete/{title}", mware.AuthMwr(userRepo.PasswordDelete, userRepo.JwtSignKey))
	// r.Put("/api/user/update/{title}", PasswordUpdate)    // TODO by title

	return r
}
