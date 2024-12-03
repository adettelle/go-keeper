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

	//is_valid := true

	sqlSt := `insert into jwttoken (customer_id, token, is_valid)
		values ($1, $2, true);` // is_valid сначала всегда true

	_, err = jr.DB.ExecContext(ctx, sqlSt, custID, token) //), is_valid)
	if err != nil {
		log.Println("error in adding jwt token:", err)
		return err
	}
	log.Println("Token is added.")
	return nil
}

// инвалидирует старые токены перед добавлением нового
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
		if err == sql.ErrNoRows { // считаем, что это не ошибка, просто не нашли токен
			return isValid, nil
		}
	}

	return isValid, nil
}
