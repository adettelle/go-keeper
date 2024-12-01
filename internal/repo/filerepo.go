package repo

import (
	"context"
	"database/sql"
	"log"
)

type FileToGet struct {
	ID          string
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
func (fr *FileRepo) AddFile(
	ctx context.Context, fileName, title, description, cloudID string, login string) error {

	sqlSt := `insert into bfile (file_name, title, description, cloud_id, customer_id) 
		values ($1, $2, $3, $4, (select id from customer where login = $5));`

	_, err := fr.DB.ExecContext(ctx, sqlSt, fileName, title, description, cloudID, login)
	if err != nil {
		log.Println("error in adding file:", err)
		return err
	}
	log.Println("File added")
	return nil
}

// получать файл может только аутентифицированный пользователь
func (fr *FileRepo) GetFileCoudIDByID(ctx context.Context, fileID, login string) (string, error) {

	sqlSt := `select cloud_id from bfile 
		where id = $1 and customer_id = (select id from customer c where login = $2);`

	row := fr.DB.QueryRowContext(ctx, sqlSt, fileID, login)

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

// GetAllFiles получает список файлов по имени (name) пользователя
func (fr *FileRepo) GetAllFiles(ctx context.Context, name string) ([]FileToGet, error) {
	files := make([]FileToGet, 0)

	sqlSt := `select id, file_name, title, description from bfile  
		where customer_id = (select id from customer c where name = $1);`

	rows, err := fr.DB.QueryContext(ctx, sqlSt, name)
	if err != nil || rows.Err() != nil {
		log.Println("error in getting files:", err)
		return nil, err
	}
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var file FileToGet
		err := rows.Scan(&file.ID, &file.FileName, &file.Title, &file.Description)
		if err != nil {
			log.Println("error: ", err)
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (fr *FileRepo) DeleteFileByCloudID(ctx context.Context, cloudID string, login string) error {
	sqlSt := `delete from bfile 
		where cloud_id = $1 and customer_id = (select id from customer where login = $2);`

	_, err := fr.DB.ExecContext(ctx, sqlSt, cloudID, login)
	if err != nil {
		log.Println("error in deleting file:", err)
		return err
	}

	log.Println("File deleted")
	return nil
}
