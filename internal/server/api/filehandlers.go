package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/config"
	"github.com/google/uuid"
)

type FileHandlers struct {
	FileRepo     IFileRepo
	MinioService IMinioService
	SignKey      []byte
	Config       *config.Config
}

func NewFileHandlers(fileRepo IFileRepo, minioService IMinioService,
	signKey []byte, cfg *config.Config) *FileHandlers {

	return &FileHandlers{
		FileRepo:     fileRepo,
		MinioService: minioService,
		SignKey:      signKey,
		Config:       cfg,
	}
}

type IFileRepo interface {
	AddFile(ctx context.Context, fileName, title, description, cloudID string, login string) error
	GetFileCoudIDByTitle(ctx context.Context, fileID, login string) (string, error)
	GetAllFiles(ctx context.Context, login string) ([]repo.FileToGet, error)
	UpdateFile(ctx context.Context, title string, fileName *string, description *string, userID int) error
	DeleteFileByTitle(ctx context.Context, title string, login string) error
	FileExists(ctx context.Context, title string, custID string) (bool, error)
}

type fileCreateRequestDTO struct {
	FileName    string `json:"fname" validate:"required,filepath"`
	Title       string `json:"title" validate:"required,alphanumunicode,min=1"`
	Description string `json:"description"`
}

func (fh *FileHandlers) FileAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")
	userID := r.Header.Get("x-user-id")
	fileName := r.Header.Get("x-file-name")
	fileTitle := r.Header.Get("x-file-title")
	fileDescription := r.Header.Get("x-file-description")
	// log.Println("==============", r.Header.Get("Content-Type"), r.Header.Get("Content-Length"))

	file := fileCreateRequestDTO{
		FileName:    fileName,
		Title:       fileTitle,
		Description: fileDescription,
	}

	err := validate.Struct(file)
	if err != nil {
		log.Println("error in validating:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cloudID := uuid.NewString() // генерируем случайную строку опр-го формата

	// убрем из filename полный путь, оставив только название файла
	fileNameWithoutPath := filepath.Base(file.FileName)

	fileExists, err := fh.FileRepo.FileExists(context.Background(), file.Title, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if fileExists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	err = fh.MinioService.Upload(cloudID, r.Body)
	if err != nil {
		log.Println("error in uploading file:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = fh.FileRepo.AddFile(
		context.Background(), fileNameWithoutPath, file.Title, file.Description, cloudID, userLogin)
	if err != nil {
		log.Println("error in adding file:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (fh *FileHandlers) FileGetByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")
	title := r.PathValue("title")

	fileCLoudID, err := fh.FileRepo.GetFileCoudIDByTitle(context.Background(), title, userLogin)
	if err != nil {
		log.Println("error in getting file by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if fileCLoudID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	obj, err := fh.MinioService.GetObject(fileCLoudID)
	if err != nil {
		log.Println("error in getting minio:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/binary")
	w.WriteHeader(http.StatusOK)

	// в http.ResponseWriter копируем (начинаем писать) содержимое файла, который вернул minio.GetObject
	_, err = io.Copy(w, obj)
	if err != nil {
		log.Println("error in copy response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type fileGetRequestDTO struct {
	FileName    string `json:"file_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewFileDTO(file repo.FileToGet) *fileGetRequestDTO {
	return &fileGetRequestDTO{
		FileName:    file.FileName,
		Title:       file.Title,
		Description: file.Description,
	}
}

func NewFileListDTO(files []repo.FileToGet) []*fileGetRequestDTO {
	res := []*fileGetRequestDTO{}
	for _, file := range files {
		res = append(res, NewFileDTO(file))
	}
	return res
}

func (fh *FileHandlers) AllFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userLogin := r.Header.Get("x-user")

	files, err := fh.FileRepo.GetAllFiles(context.Background(), userLogin)
	if err != nil {
		log.Println("error in getting files: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(NewFileListDTO(files))
	if err != nil {
		log.Println("error in marshalling json:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println("error in writing resp:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type fileUpdateRequestDTO struct {
	FileName    *string `json:"fname,omitempty"` // для ссылки значение по умолчанию - nil
	Description *string `json:"description,omitempty"`
}

func (fh *FileHandlers) FileUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("x-user-id")
	custID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	var file fileUpdateRequestDTO
	title := r.PathValue("title")

	// читаем тело запроса
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &file); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = fh.FileRepo.UpdateFile(context.Background(), title, file.FileName, file.Description, custID)
	if err != nil {
		log.Println("error in updating password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted) // новый файл принят
}

func (fh *FileHandlers) FileDeleteByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	title := r.PathValue("title")
	userLogin := r.Header.Get("x-user")

	err := fh.FileRepo.DeleteFileByTitle(context.Background(), title, userLogin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
