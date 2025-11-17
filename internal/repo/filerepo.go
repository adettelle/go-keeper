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

type FileRepo struct {
	DB *sql.DB
}

func NewFileRepo(db *sql.DB) *FileRepo {
	return &FileRepo{
		DB: db,
	}
}

func (fr *FileRepo) FileExists(ctx context.Context, title string, custID string) (bool, error) {
	sqlSt := `select exists (select 1 from bfile 
		where title = $1 and customer_id = $2);`

	row := fr.DB.QueryRowContext(ctx, sqlSt, title, custID)

	var fileExists *bool

	err := row.Scan(&fileExists)
	if err != nil {
		return false, err
	}

	return *fileExists, nil
}

// добавлять файл может только аутентифицированный пользователь
func (fr *FileRepo) AddFile(
	ctx context.Context, fileName, title, description, cloudID string, login string) error {

	sqlSt := `insert into bfile (file_name, title, description, cloud_id, customer_id) 
		values ($1, $2, $3, $4, (select id from customer where login = $5));`

	_, err := fr.DB.ExecContext(ctx, sqlSt, fileName, title, description, cloudID, login)
	if err != nil {
		log.Println("error in adding file:", err)
		return err
	}
	log.Println("File is added.")
	return nil
}

// получать файл может только аутентифицированный пользователь
func (fr *FileRepo) GetFileCoudIDByTitle(ctx context.Context, title, login string) (string, error) {

	sqlSt := `select cloud_id from bfile 
		inner join customer c on c.id = bfile.customer_id 
		where title = $1 and c.login = $2;`

	row := fr.DB.QueryRowContext(ctx, sqlSt, title, login)

	var fileCloudID *string

	err := row.Scan(&fileCloudID)
	if err != nil {
		log.Println("error in scan: ", err)
		if err == sql.ErrNoRows { // считаем, что это не ошибка, просто не нашли пользователя
			return "", nil
		}
	}

	return *fileCloudID, nil
}

type FileToGet struct {
	ID          string
	FileName    string
	Title       string
	Description string
}

// GetAllFiles получает список файлов по имени (name) пользователя
func (fr *FileRepo) GetAllFiles(ctx context.Context, login string) ([]FileToGet, error) {
	files := make([]FileToGet, 0)

	sqlSt := `select title, file_name,  description from bfile  
		inner join customer c on c.id = bfile.customer_id 
		where c.login = $1
		order by title;`

	rows, err := fr.DB.QueryContext(ctx, sqlSt, login)
	if err != nil || rows.Err() != nil {
		log.Println("error in getting files:", err)
		return nil, err
	}
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var file FileToGet
		err := rows.Scan(&file.Title, &file.FileName, &file.Description)
		if err != nil {
			log.Println("error: ", err)
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (fr *FileRepo) DeleteFileByTitle(ctx context.Context, title string, login string) error {
	sqlSt := `delete from bfile 
		where title = $1 and customer_id = (select id from customer c where c.login = $2);`

	_, err := fr.DB.ExecContext(ctx, sqlSt, title, login)
	if err != nil {
		log.Println("error in deleting file:", err)
		return err
	}

	log.Println("File is deleted.")
	return nil
}

// UpdateFile updates file values (file_name, description) by card title.
// если в json не передать поле, то оно не измениться
// если передать пустую строку "" - то поле станет пустым
func (pr *FileRepo) UpdateFile(ctx context.Context,
	title string, fileName *string, description *string, userID int) error {

	type file struct {
		FileName    *string `db:"fname" goqu:"omitnil"`
		Description *string `db:"description" goqu:"omitnil"`
	}

	sqlSt, args, _ := goqu.Update("bfile").Set(file{
		FileName:    fileName,
		Description: description,
	}).Where(goqu.C("title").Eq(title)).Where(goqu.C("customer_id").Eq(userID)).ToSQL()

	_, err := pr.DB.ExecContext(ctx, sqlSt, args...)
	if err != nil {
		log.Println("error in changing file info:", err)
		return err
	}
	log.Println("File info is changed.")
	return nil
}
