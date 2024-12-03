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

// добавлять пароль может только аутентифицированный пользователь
func (pr *PasswordRepo) CreatePassword(
	ctx context.Context, password, title, description string, login string) error {

	// не проверяем наличие у пользователя пароля с таким title,
	// потому что етсь unique (title, customer_id)
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

// UpdatePassword меняет поля (pwd, description) по title пароля
// если в json не передать поле, то оно не измениться
// если передать пустую строку "" - то поле станет пустым
func (pr *PasswordRepo) UpdatePassword(ctx context.Context,
	title string, password *string, description *string, userID int) error { // id int,

	type pwd struct {
		// TODO эти поля в таблице не должны быть пустыми! not empty?
		// или сделать валидацию при записи в таблицу??????????
		// Title       *string `db:"title" goqu:"omitnil"`
		Password    *string `db:"pwd" goqu:"omitnil"`
		Description *string `db:"description" goqu:"omitnil"`
	}

	sqlSt, args, _ := goqu.Update("pass").Set(pwd{
		// Title:       title,
		Password:    password,
		Description: description,
	}).Where(goqu.C("title").Eq(title)).Where(goqu.C("customer_id").Eq(userID)).ToSQL()
	// было: Where(goqu.C("id").Eq(id))  - TODO CHECK

	// fmt.Println(sqlSt)

	_, err := pr.DB.ExecContext(ctx, sqlSt, args...)
	if err != nil {
		log.Println("error in changing password info:", err)
		return err
	}
	log.Println("Password info is changed.")
	return nil
}

// DeletePassword удаляет пароль с определенным названием (title) по id пользователя
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

// GetAllPasswords получает список паролей по логину пользователя
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
