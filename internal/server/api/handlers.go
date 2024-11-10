package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adettelle/go-keeper/internal/repo"
)

type UserRepo struct {
	CustomerRepo ICustomerRepo
}

type ICustomerRepo interface {
	AddCustomer(ctx context.Context, name, email, masterPassword string) error
}

type CustomerDTO struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	MasterPassword string `json:"masterpassword"`
}

func (ur *UserRepo) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer
	var customer CustomerDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &customer); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ur.CustomerRepo.AddCustomer(
		context.Background(), customer.Name, customer.Email, customer.MasterPassword)
	if err != nil {
		if repo.IsCustomerExistsErr(err) { // !!!!!!!!!!!!!!!!!!!!!!!!!
			log.Printf("error %v in registering user %s", err, customer.Email)
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println("error in adding user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
