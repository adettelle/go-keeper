package repo

import (
	"context"
	"database/sql"
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

// регистрация пользователя, надо ли проверить, что такого пользователя нет???????????????????
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

	log.Println("Registered")
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
