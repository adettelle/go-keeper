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
	"database/sql"
	"log"

	"github.com/doug-martin/goqu/v9"
)

type PasswordRepo struct {
	DB *sql.DB
}

func NewPasswordRepo(db *sql.DB) *PasswordRepo {
	return &PasswordRepo{
		DB: db,
	}
}

func (pr *PasswordRepo) CreatePassword(
	ctx context.Context, password, title, description string, login string) error {

	// do not need to check if the user has a password with this title,
	// because there is a unique (title, customer_id) in table
	sqlSt := `insert into pass (pwd, title, description, customer_id) 
		values ($1, $2, $3, (select id from customer where login = $4));`

	_, err := pr.DB.ExecContext(ctx, sqlSt, password, title, description, login)
	if err != nil {
		log.Println("error in adding password:", err)
		return err
	}
	log.Println("Password is added.")
	return nil
}

// UpdatePassword updates the password and description by title.
// Fields not provided in json remain unchanged.
// Passing an empty string sets the field to empty.
func (pr *PasswordRepo) UpdatePassword(ctx context.Context,
	title string, password *string, description *string, userID int) error {

	type pwd struct {
		Password    *string `db:"pwd" goqu:"omitnil"`
		Description *string `db:"description" goqu:"omitnil"`
	}

	sqlSt, args, _ := goqu.Update("pass").Set(pwd{
		Password:    password,
		Description: description,
	}).Where(goqu.C("title").Eq(title)).Where(goqu.C("customer_id").Eq(userID)).ToSQL()

	_, err := pr.DB.ExecContext(ctx, sqlSt, args...)
	if err != nil {
		log.Println("error in changing password info:", err)
		return err
	}
	log.Println("Password info is changed.")
	return nil
}

// DeletePassword removes a password entry by title and by user login.
func (pr *PasswordRepo) DeletePassword(ctx context.Context, title string, login string) error {
	sqlSt := `delete from pass  
		where title = $1 and pass.customer_id = (select id from customer c where c.login = $2);`

	_, err := pr.DB.ExecContext(ctx, sqlSt, title, login)
	if err != nil {
		log.Println("error in deleting password:", err)
		return err
	}

	log.Println("Password is deleted.")
	return nil
}

func (pr *PasswordRepo) GetPasswordByTitle(
	ctx context.Context, title string, login string) (string, error) {

	sqlSt := `select pwd from pass 
		inner join customer c on c.id = pass.customer_id 
		where title = $1 and c.login = $2;`

	row := pr.DB.QueryRowContext(ctx, sqlSt, title, login)

	var pwd *string

	err := row.Scan(&pwd)
	if err != nil {
		log.Println("error in scan:", err)
		if err == sql.ErrNoRows { // считаем, что это не ошибка, просто не нашли пользователя
			return "", nil
		}
	}
	return *pwd, nil
}

type Password struct {
	Title       string
	Description string
}

// GetAllPasswords retrieves a list of all passwords info (title and description) for the specified user.
func (pr *PasswordRepo) GetAllPasswords(ctx context.Context, login string) ([]Password, error) {
	pwds := make([]Password, 0)

	sqlSt := `select title, description from pass 
		inner join customer c on c.id = pass.customer_id 
		where c.login = $1
		order by pass.title;`

	rows, err := pr.DB.QueryContext(ctx, sqlSt, login)
	if err != nil || rows.Err() != nil {
		log.Println("error in getting passwords:", err)
		return nil, err
	}
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var pwd Password
		err := rows.Scan(&pwd.Title, &pwd.Description)
		if err != nil {
			log.Println("error:", err)
			return nil, err
		}
		pwds = append(pwds, pwd)
	}

	return pwds, nil
}
