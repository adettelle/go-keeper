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
)

type JwtRepo struct {
	DB *sql.DB
}

func NewJwtRepo(db *sql.DB) *JwtRepo {
	return &JwtRepo{
		DB: db,
	}
}

func (jr *JwtRepo) AddJwtToken(ctx context.Context, custID int, token string) error {
	err := jr.invalidateTokens(custID)
	if err != nil {
		return err
	}

	sqlSt := `insert into jwttoken (customer_id, token, is_valid)
		values ($1, $2, true);` // is_valid is always true initially

	_, err = jr.DB.ExecContext(ctx, sqlSt, custID, token)
	if err != nil {
		log.Println("error in adding jwt token:", err)
		return err
	}
	log.Println("Token is added.")
	return nil
}

// invalidateTokens sets all existing tokens for a specific customer as invalid in the database.
// This function is called internally before adding a new token.
func (jr *JwtRepo) invalidateTokens(custID int) error {
	sqlSt := `update jwttoken set is_valid = false where customer_id = $1;`

	_, err := jr.DB.ExecContext(context.Background(), sqlSt, custID)
	if err != nil {
		log.Println("error in invalidating tokens:", err)
		return err
	}
	log.Println("Tokens are invalidated.")
	return nil
}

func (jr *JwtRepo) TokenIsValid(ctx context.Context, token string) (bool, error) {
	sqlSt := `select is_valid from jwttoken where token = $1;`

	row := jr.DB.QueryRowContext(ctx, sqlSt, token)

	var isValid bool

	err := row.Scan(&isValid)
	if err != nil {
		log.Println("error in scan:", err)
		if err == sql.ErrNoRows { // This is not considered an error; the token was simply not found.
			return isValid, nil
		}
	}

	return isValid, nil
}
