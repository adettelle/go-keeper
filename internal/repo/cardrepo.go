package repo

import (
	"context"
	"database/sql"
	"log"
)

type CardToGet struct {
	ID          string
	Num         string
	Expire      string
	Description string
	Cvc         string
}

type CardGetByID struct {
	Num         string
	Expire      string
	Description string
	Cvc         string
}

type CardRepo struct {
	DB *sql.DB
}

func NewCardRepo(db *sql.DB) *CardRepo {
	return &CardRepo{
		DB: db,
	}
}

// добавлять файл может только аутентифицированный пользователь
func (cr *CardRepo) AddCard(
	ctx context.Context, cardNum, expire, description, cvc string, login string) error {

	sqlSt := `insert into card (num, expire, description, cvc, customer_id) 
		values ($1, $2, $3, $4, (select id from customer where login = $5));`

	_, err := cr.DB.ExecContext(ctx, sqlSt, cardNum, expire, description, cvc, login)
	if err != nil {
		log.Println("error in adding card:", err)
		return err
	}
	log.Println("Card added")
	return nil
}

// получать файл может только аутентифицированный пользователь
func (cr *CardRepo) GetCardByID(ctx context.Context, cardID, login string) (CardGetByID, error) {

	sqlSt := `select num, expire, description, cvc from card
		where id = $1 and customer_id = (select id from customer c where login = $2);`

	row := cr.DB.QueryRowContext(ctx, sqlSt, cardID, login)

	var card CardGetByID

	err := row.Scan(&card.Num, &card.Expire, &card.Description, &card.Cvc)
	if err != nil {
		log.Println("error in scan: ", err)
		if err == sql.ErrNoRows { // считаем, что это не ошибка, просто не нашли пользователя
			return card, nil
		}
	}

	return card, nil
}

// GetAllCards получает список паролей по логину пользователя
func (cr *CardRepo) GetAllCards(ctx context.Context, login string) ([]CardToGet, error) {
	log.Println("---------------")
	log.Println("login1:", login)
	cards := make([]CardToGet, 0)

	sqlSt := `select id, num, expire, description, cvc from card  
		where customer_id = (select id from customer where login = $1);`

	rows, err := cr.DB.QueryContext(ctx, sqlSt, login)
	if err != nil || rows.Err() != nil {
		log.Println("error in getting cards:", err)
		return nil, err
	}
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var card CardToGet
		err := rows.Scan(&card.ID, &card.Num, &card.Expire, &card.Description, &card.Cvc)
		if err != nil {
			log.Println("error: ", err)
			return nil, err
		}
		cards = append(cards, card)
	}
	log.Println("cards:", cards)
	return cards, nil
}
