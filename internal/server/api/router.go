package api

import (
	"github.com/adettelle/go-keeper/pkg/mware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(userRepo *CustomerHandlers) chi.Router {
	r := chi.NewRouter()

	r.Post("/api/user/register", userRepo.RegisterCustomer)
	r.Post("/api/user/login", userRepo.Login)

	r.Get("/api/user/passwords", mware.AuthMwr(userRepo.AllPasswords, userRepo.JwtSignKey))
	r.Get("/api/user/password/{title}", mware.AuthMwr(userRepo.PasswordByTitle, userRepo.JwtSignKey))
	r.Put("/api/user/password", mware.AuthMwr(userRepo.PasswordCreate, userRepo.JwtSignKey))
	r.Post("/api/user/password/update", mware.AuthMwr(userRepo.PasswordUpdate, userRepo.JwtSignKey))
	r.Delete("/api/user/delete/{title}", mware.AuthMwr(userRepo.PasswordDelete, userRepo.JwtSignKey))

	r.Put("/api/user/addfile", mware.AuthMwr(userRepo.FileAdd, userRepo.JwtSignKey))
	r.Get("/api/user/getfile/{id}", mware.AuthMwr(userRepo.FileGetByID, userRepo.JwtSignKey))
	r.Get("/api/user/files", mware.AuthMwr(userRepo.AllFiles, userRepo.JwtSignKey))

	return r
}
