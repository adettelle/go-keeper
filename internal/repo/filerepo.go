package repo

import (
	"context"
	"database/sql"
	"log"
)

type FileRepo struct {
	DB *sql.DB
}

func NewFileRepo(db *sql.DB) *FileRepo {
	return &FileRepo{
		DB: db,
	}
}

// добавлять файл может только аутентифицированный пользователь
func (pr *FileRepo) AddFile(
	ctx context.Context, fileName, title, description, cloudID string, login string) error {

	sqlSt := `insert into bfile (file_name, title, description, cloud_id, customer_id) 
		values ($1, $2, $3, $4, (select id from customer where login = $5));`

	_, err := pr.DB.ExecContext(ctx, sqlSt, fileName, title, description, cloudID, login)
	if err != nil {
		log.Println("error in adding file:", err)
		return err
	}
	log.Println("File added")
	return nil
}
