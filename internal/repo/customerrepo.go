// Package repo provides functionality for interacting
// with the card, file and password database repositories.
// It allows adding, retrieving, updating, and deleting related data
// while ensuring proper access controls.
// Package provides functionality for managing customer data in the database.
// It includes operations for adding customers, retrieving customer details,
// and verifying user credentials.
// Also it provides functionality for managing JWT tokens in a database,
// including adding, invalidating, and checking the validity of tokens.
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
	DB *sql.DB
}

func NewCustomerRepo(db *sql.DB) *CustomerRepo {
	return &CustomerRepo{
		DB: db,
	}
}

// CustomerExistsErr represents an error indicating that a customer with the given login already exists.
type CustomerExistsErr struct {
	login string // The login of the existing customer
}

func (ce *CustomerExistsErr) Error() string {
	return fmt.Sprintf("customer %s already exists", ce.login)
}

func NewCustomerExistsErr(login string) *CustomerExistsErr {
	return &CustomerExistsErr{
		login: login,
	}
}

// IsCustomerExistsErr checks if an error is of type CustomerExistsErr.
func IsCustomerExistsErr(err error) bool {
	var customErr *CustomerExistsErr
	return errors.As(err, &customErr)
}

// AddCustomer adds a new customer to the database.
// Returns the ID of the newly added customer and an error if the operation fails.
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

// VerifyUser verifies the credentials of a user and checks their authorization.
// VerifyUser — функция, которая выполняет аутентификацию и авторизацию пользователя
//
// Parameters:
//   - ctx: The context for managing request lifetimes.
//   - login: The login (email) of the user.
//   - pass: The master password of the user.
//
// Returns:
//   - A boolean indicating if the user is authenticated and their ID.
func (cr *CustomerRepo) VerifyUser(ctx context.Context, login string, pass string) (bool, int) {
	if login == "" || pass == "" {
		return false, 0
	}
	// Generate a hash of the provided password
	hashedPassword := sha256.Sum256([]byte(pass))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	// Retrieve the customer by login
	cust, err := cr.GetCustomerByLogin(ctx, login)
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
