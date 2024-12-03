package repo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
)

type CustomerRepo struct {
	DB *sql.DB // Customers []Customer
}

func NewCustomerRepo(db *sql.DB) *CustomerRepo {
	return &CustomerRepo{
		DB: db,
	}
}

type CustomerExistsErr struct {
	login string
}

func (ce *CustomerExistsErr) Error() string {
	return fmt.Sprintf("customer %s already exists", ce.login)
}

func NewCustomerExistsErr(login string) *CustomerExistsErr {
	return &CustomerExistsErr{
		login: login,
	}
}

func IsCustomerExistsErr(err error) bool {
	var customErr *CustomerExistsErr
	return errors.As(err, &customErr)
}

func (cr *CustomerRepo) AddCustomer(ctx context.Context, name, login, masterPassword string) (int, error) {
	sqlCustomer := `select count(*) > 0 from customer where login = $1 limit 1;`
	row := cr.DB.QueryRowContext(ctx, sqlCustomer, login)

	// переменная для чтения результата
	var customerExists bool

	err := row.Scan(&customerExists)
	if err != nil {
		return 0, err
	}
	if customerExists {
		return 0, NewCustomerExistsErr(login) // пользователь уже существует
	}

	sqlSt := `insert into customer (name, login, masterpassword) values ($1, $2, $3) returning id;`

	var custID int
	row = cr.DB.QueryRowContext(ctx, sqlSt, name, login, masterPassword)
	err = row.Scan(&custID)
	if err != nil {
		return 0, err
	}

	log.Println("Customer is registered.")
	return custID, nil
}

type CustomerGetByLogin struct {
	ID             int
	Name           string
	Login          string
	MasterPassword string
}

func (cr *CustomerRepo) GetCustomerByLogin(ctx context.Context, login string) (*CustomerGetByLogin, error) {
	sqlSt := `select id, name, login, masterpassword from customer where login = $1;`

	row := cr.DB.QueryRowContext(ctx, sqlSt, login)

	var customer CustomerGetByLogin

	err := row.Scan(&customer.ID, &customer.Name, &customer.Login, &customer.MasterPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // считаем, что это не ошибка, просто не нашли пользователя
		}
		return nil, err
	}
	return &customer, nil
}

// VerifyUser — функция, которая выполняет аутентификацию и авторизацию пользователя
// login — это email пользователя, pass — это masterpassword, permission — необходимая привилегия.
// если пользователь ввел правильные данные, и у него есть необходимая привилегия — возвращаем true, иначе — false
// VerifyUser возвращает userID и bool (существует ли пользователь)
func (cr *CustomerRepo) VerifyUser(ctx context.Context, login string, pass string) (bool, int) {
	if login == "" || pass == "" {
		return false, 0
	}
	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(pass))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	// проверяем введенные данные
	cust, err := cr.GetCustomerByLogin(ctx, login)
	log.Println("++++++", cust)
	if err != nil {
		log.Printf("Error in authorization %s", cust.Login)
		return false, 0
	}
	if cust == nil {
		log.Printf("Error in authorization %s, user not found", login)
		return false, 0
	}

	return cust.MasterPassword == hashStringPassword, cust.ID
}
