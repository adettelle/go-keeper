package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Customer struct {
	Name           string
	Email          string
	MasterPassword string
}

type CustomerRepo struct {
	DB *sql.DB // Customers []Customer
}

func NewCustomerRepo(db *sql.DB) *CustomerRepo {
	return &CustomerRepo{
		DB: db,
	}
}

type CustomerExistsErr struct {
	email string
}

func (ce *CustomerExistsErr) Error() string {
	return fmt.Sprintf("customer %s already exists", ce.email)
}

func NewCustomerExistsErr(email string) *CustomerExistsErr {
	return &CustomerExistsErr{
		email: email,
	}
}

func IsCustomerExistsErr(err error) bool {
	var customErr *CustomerExistsErr
	return errors.As(err, &customErr)
}

// регистрация пользователя
func (cr *CustomerRepo) AddCustomer(ctx context.Context, name, email, masterPassword string) error {
	sqlCustomer := `select count(*) > 0 from customer where email = $1 limit 1;`
	row := cr.DB.QueryRowContext(ctx, sqlCustomer, email)

	// переменная для чтения результата
	var customerExists bool

	err := row.Scan(&customerExists)
	if err != nil {
		return err
	}
	if customerExists {
		return NewCustomerExistsErr(email) // пользователь уже существует
	}

	sqlSt := `insert into customer (name, email, masterpassword) values ($1, $2, $3);`

	_, err = cr.DB.ExecContext(ctx, sqlSt, name, email, masterPassword)
	if err != nil {
		log.Println("error in registering customer:", err)
		return err
	}
	log.Println("Registered")
	return nil
}

// func (cr *CustomerRepo) GetCustomerByEmail(ctx context.Context, email string) (*Customer, error) {
//}

// type ICustomerRepo interface {
// }

// --------

// func CreateCustomer(name, email, masterPassword string) Customer {
// 	return Customer{Name: name, Email: email, MasterPassword: masterPassword}
// }

// func (c *CustomerRepo) AddCustomer(cust Customer) {
// 	c.Customers = append(c.Customers, cust)
// }
