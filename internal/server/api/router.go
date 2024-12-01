package api

import (
	"github.com/adettelle/go-keeper/pkg/mware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(userRepo *CustomerHandlers, jwtChecker mware.JwtChecker) chi.Router {
	r := chi.NewRouter()

	r.Post("/api/user/register", userRepo.RegisterCustomer)
	r.Post("/api/user/login", userRepo.Login)

	r.Put("/api/user/password", mware.AuthMwr(userRepo.PasswordCreate, userRepo.JwtSignKey, jwtChecker))
	r.Get("/api/user/passwords", mware.AuthMwr(userRepo.AllPasswords, userRepo.JwtSignKey, jwtChecker))
	r.Get("/api/user/password/{title}", mware.AuthMwr(userRepo.PasswordByTitle, userRepo.JwtSignKey, jwtChecker))
	r.Post("/api/user/password/update", mware.AuthMwr(userRepo.PasswordUpdate, userRepo.JwtSignKey, jwtChecker))
	r.Delete("/api/user/delete/{title}", mware.AuthMwr(userRepo.PasswordDelete, userRepo.JwtSignKey, jwtChecker))

	r.Put("/api/user/addfile", mware.AuthMwr(userRepo.FileAdd, userRepo.JwtSignKey, jwtChecker))
	r.Get("/api/user/getfile/{id}", mware.AuthMwr(userRepo.FileGetByID, userRepo.JwtSignKey, jwtChecker))
	r.Get("/api/user/files", mware.AuthMwr(userRepo.AllFiles, userRepo.JwtSignKey, jwtChecker))
	r.Delete("/api/user/delete/file/{cloudid}", mware.AuthMwr(userRepo.FileDeleteByCloudID, userRepo.JwtSignKey, jwtChecker))

	r.Put("/api/user/addcard", mware.AuthMwr(userRepo.CardAdd, userRepo.JwtSignKey, jwtChecker))
	r.Get("/api/user/cards", mware.AuthMwr(userRepo.AllCards, userRepo.JwtSignKey, jwtChecker))
	r.Get("/api/user/card/{id}", mware.AuthMwr(userRepo.CardGetByID, userRepo.JwtSignKey, jwtChecker))
	r.Delete("/api/user/delete/card/{id}", mware.AuthMwr(userRepo.CardDeleteByID, userRepo.JwtSignKey, jwtChecker))

	return r
}
