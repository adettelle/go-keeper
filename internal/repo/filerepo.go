package repo

import (
	"context"
	"database/sql"
	"log"
)

type File struct {
	FileName    string
	Title       string
	Description string
}

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

// получать файл может только аутентифицированный пользователь
func (pr *FileRepo) GetFileCoudIDByID(ctx context.Context, fileID, login string) (string, error) {

	sqlSt := `select cloud_id from bfile 
		where id = $1 and customer_id = (select id from customer c where login = $2);`

	row := pr.DB.QueryRowContext(ctx, sqlSt, fileID, login)

	var fileCloudID *string

	err := row.Scan(&fileCloudID)
	if err != nil {
		log.Println("error in scan: ", err)
		if err == sql.ErrNoRows { // считаем, что это не ошибка, просто не нашли пользователя
			return "", nil
		}
	}

	log.Printf("File cloudID is: %s\n", *fileCloudID)
	return *fileCloudID, nil
}

// GetAllPasswords получает список паролей по имени (name) пользователя
func (pr *FileRepo) GetAllFiles(ctx context.Context, name string) ([]File, error) {
	files := make([]File, 0)

	sqlSt := `select file_name, title, description from bfile  
		where customer_id = (select id from customer c where name = $1);`

	rows, err := pr.DB.QueryContext(ctx, sqlSt, name)
	if err != nil || rows.Err() != nil {
		log.Println("error in getting files:", err)
		return nil, err
	}
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var file File
		err := rows.Scan(&file.FileName, &file.Title, &file.Description)
		if err != nil {
			log.Println("error: ", err)
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}
