package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/adettelle/go-keeper/internal/encryption"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/config"
)

type PassHandlers struct {
	PwdRepo IPwdRepo
	SignKey []byte
	Config  *config.Config
}

func NewPassHandlers(pwdRepo IPwdRepo, signKey []byte, cfg *config.Config) *PassHandlers {
	return &PassHandlers{
		PwdRepo: pwdRepo,
		SignKey: signKey,
		Config:  cfg,
	}
}

type IPwdRepo interface {
	GetAllPasswords(ctx context.Context, name string) ([]repo.Password, error)
	CreatePassword(ctx context.Context, password, title, description string, login string) error
	UpdatePassword(ctx context.Context, title string, password *string, description *string, userID int) error
	DeletePassword(ctx context.Context, title string, login string) error
	GetPasswordByTitle(ctx context.Context, title string, login string) (string, error)
}

type PasswordResponseDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewPwdResponseDTO(pwd repo.Password) *PasswordResponseDTO {
	return &PasswordResponseDTO{
		Title:       pwd.Title,
		Description: pwd.Description,
	}
}

func NewPwdListResponseDTO(pwds []repo.Password) []*PasswordResponseDTO {
	res := []*PasswordResponseDTO{}
	for _, pwd := range pwds {
		res = append(res, NewPwdResponseDTO(pwd))
	}
	return res
}

// Далее все хендлеры доступны только авторизованному пользователю
func (ph *PassHandlers) AllPasswords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userLogin := r.Header.Get("x-user")

	pwds, err := ph.PwdRepo.GetAllPasswords(context.Background(), userLogin)
	if err != nil {
		log.Println("error in getting passwords: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(NewPwdListResponseDTO(pwds))
	if err != nil {
		log.Println("error in marshalling json:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println("error in writing resp:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type PwdCreateRequestDTO struct {
	Password    string `json:"pwd" validate:"required,min=1"`
	Title       string `json:"title" validate:"required,min=1"`
	Description string `json:"description"`
}

func (ph *PassHandlers) PasswordCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")

	var buf bytes.Buffer
	var pwd PwdCreateRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &pwd); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validate.Struct(pwd)
	if err != nil {
		log.Println("error in validating:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	encryptedPass, err := encryption.AESEncrypt(pwd.Password, ph.SignKey)
	if err != nil {
		log.Println("error in encrypting password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = ph.PwdRepo.CreatePassword(
		context.Background(), encryptedPass, pwd.Title, pwd.Description, userLogin)
	if err != nil {
		log.Println("error in adding password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted) // новый пароль принят
}

type pwdUpdateRequestDTO struct {
	Password    *string `json:"pwd,omitempty"`
	Description *string `json:"description,omitempty"` // для ссылки значение по умолчанию - nil
}

func (ph *PassHandlers) PasswordUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("x-user-id")
	custID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	var pwd pwdUpdateRequestDTO
	title := r.PathValue("title")

	// читаем тело запроса
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &pwd); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	encryptedPass, err := encryption.AESEncrypt(*pwd.Password, ph.SignKey)
	if err != nil {
		log.Println("error in encrypting password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = ph.PwdRepo.UpdatePassword(context.Background(), title, &encryptedPass, pwd.Description, custID)
	if err != nil {
		log.Println("error in updating password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted) // новый пароль принят
}

func (ph *PassHandlers) PasswordDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pwdTitle := r.PathValue("title")
	userLogin := r.Header.Get("x-user")

	err := ph.PwdRepo.DeletePassword(context.Background(), pwdTitle, userLogin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type PwdRespDTO struct {
	Password string `json:"pwd"`
}

func (ph *PassHandlers) PasswordByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")
	pwdTitle := r.PathValue("title")

	pwd, err := ph.PwdRepo.GetPasswordByTitle(context.Background(), pwdTitle, userLogin)
	if err != nil {
		log.Println("error in getting password by title:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if pwd == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decryptedPass, err := encryption.AESDecrypt(pwd, ph.SignKey)
	if err != nil {
		log.Println("error in decrypting password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(decryptedPass))
	if err != nil {
		log.Println("error in writing resp:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
