package repo

import (
	"context"
	"database/sql"
	"log"
)

type CardGetByID struct {
	Num         string
	Expire      string
	Description string
	Cvc         string
	Title       string
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
	ctx context.Context, cardNum, expire, cvc, title, description string, login string) error {

	sqlSt := `insert into card (num, expires_at, cvc, title, description, customer_id) 
		values ($1, $2, $3, $4, $5, (select id from customer where login = $6));` // TODO

	_, err := cr.DB.ExecContext(ctx, sqlSt, cardNum, expire, cvc, title, description, login)
	if err != nil {
		log.Println("error in adding card:", err)
		return err
	}
	log.Println("Card added")
	return nil
}

// получать файл может только аутентифицированный пользователь
func (cr *CardRepo) GetCardByID(ctx context.Context, cardID, login string) (CardGetByID, error) {

	sqlSt := `select num, expires_at, cvc, title, description from card
		inner join customer c on c.id = card.customer_id 
		where card.id = $1 and c.login = $2;`
	// TODO expire rename expires_at

	row := cr.DB.QueryRowContext(ctx, sqlSt, cardID, login)

	var card CardGetByID

	err := row.Scan(&card.Num, &card.Expire, &card.Cvc, &card.Title, &card.Description)
	if err != nil {
		log.Println("error in scan: ", err)
		if err == sql.ErrNoRows { // считаем, что это не ошибка, просто не нашли пользователя
			return card, nil
		}
	}

	return card, nil
}

type CardToGet struct {
	Num         string
	Title       string
	Description string
}

// TODO показать 4 последних цифры из 16
// GetAllCards получает список карт по логину пользователя
func (cr *CardRepo) GetAllCards(ctx context.Context, login string) ([]CardToGet, error) {
	log.Println("login1:", login)
	cards := make([]CardToGet, 0)

	sqlSt := `select num, title, description from card  
		inner join customer c on c.id = card.customer_id
		where c.login = $1;`

	rows, err := cr.DB.QueryContext(ctx, sqlSt, login)
	if err != nil || rows.Err() != nil {
		log.Println("error in getting cards:", err)
		return nil, err
	}
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var card CardToGet
		err := rows.Scan(&card.Num, &card.Title, &card.Description)
		if err != nil {
			log.Println("error: ", err)
			return nil, err
		}
		modifiedNum := "************" + card.Num[12:]
		card.Num = modifiedNum
		cards = append(cards, card)
	}

	log.Println("cards:", cards)
	return cards, nil
}

func (cr *CardRepo) DeleteCardByID(ctx context.Context, cardID, login string) error {
	sqlSt := `delete from card 
		where card.id = $1 and customer_id = (select id from customer where login = $2);`

	_, err := cr.DB.ExecContext(ctx, sqlSt, cardID, login)
	if err != nil {
		log.Println("error in deleting password:", err)
		return err
	}

	log.Println("Password deleted")
	return nil
}

type CardGetByDescription struct {
	Num         string
	Expire      string
	Cvc         string
	Title       string
	Description string
}

// получать файл может только аутентифицированный пользователь
func (cr *CardRepo) GetCardByTitle(ctx context.Context,
	cardTitle, login string) (CardGetByDescription, error) {

	sqlSt := `select num, expires_at, cvc, title, description from card
		inner join customer c on c.id = card.customer_id 
		where card.title = $1 and c.login = $2;`
	// TODO expire rename expires_at

	row := cr.DB.QueryRowContext(ctx, sqlSt, cardTitle, login)

	var card CardGetByDescription

	err := row.Scan(&card.Num, &card.Expire, &card.Cvc, &card.Title, &card.Description)
	if err != nil {
		log.Println("error in scan: ", err)
		if err == sql.ErrNoRows { // считаем, что это не ошибка, просто не нашли пользователя
			return card, nil
		}
	}

	return card, nil
}
