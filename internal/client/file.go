// Package client provides functionality for managing card, file, password and user information through
// HTTP requests, including operations to add, retrieve, update, and delete.
package client

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/adettelle/go-keeper/cmd/settings"
	"github.com/adettelle/go-keeper/internal/localstorage"
	"github.com/carlmjohnson/requests"
	"github.com/jedib0t/go-pretty/v6/table"
)

// FileService is a service for managing file-related operations.
type FileService struct {
	transport *http.Transport
	keyStore  localstorage.IKeyStorage
}

func NewFileService(transport *http.Transport, keyStore localstorage.IKeyStorage) *FileService {
	return &FileService{
		transport: transport,
		keyStore:  keyStore,
	}
}

// FileToGetAll retrieves all files associated with the user.
// It fetches file data from the server and displays it in a tabular format.
type FileToGetAll struct {
	FileName    string `json:"file_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (fs *FileService) AllFiles() error {
	jwtToken, err := fs.keyStore.Get()
	if err != nil {
		return err
	}

	var files []FileToGetAll

	err = requests.
		URL("/api/user/files").
		Host(settings.ServerURL).
		Scheme("https").
		Transport(fs.transport).
		Header("Authorization", string(jwtToken)).
		ToJSON(&files).
		Method(http.MethodGet).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Title", "File Name", "Description"})

		for _, file := range files {
			t.AppendRow([]interface{}{file.Title, file.FileName, file.Description})
		}
		t.Render()
	}
	return nil
}

func (fs *FileService) DeleteFileByTitle(title string) error {
	jwtToken, err := fs.keyStore.Get()
	if err != nil {
		log.Fatal(err)
	}

	err = requests.
		URL("/api/user/file/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(fs.transport).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodDelete).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		log.Println("File is deleted.")
	}
	return nil
}

func (fs *FileService) AddFile(fileName, title, description string) error {
	jwtToken, err := fs.keyStore.Get()
	if err != nil {
		log.Fatal(err)
	}

	fileStat, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	// log.Println("------------", fileStat.Size())
	// mtype, err := mimetype.DetectFile("/path/to/file")
	// fmt.Println(mtype.String(), mtype.Extension())

	err = requests.
		URL("/api/user/file").
		Host(settings.ServerURL).
		Scheme("https").
		Header("Authorization", string(jwtToken)).
		Header("x-file-name", fileName).
		Header("x-file-title", title).
		Header("x-file-description", description).
		Header("Content-Length", strconv.Itoa(int(fileStat.Size()))).
		Transport(fs.transport).
		BodyFile(fileName).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		return err
	}

	return nil
}

type FileToGetByTitle struct {
	Title       string `json:"title"`
	FileName    string `json:"fname"`
	Description string `json:"description"`
}

func (fs *FileService) GetFile(title string) error {
	jwtToken, err := fs.keyStore.Get()
	if err != nil {
		log.Fatal(err)
	}

	err = requests.
		URL("/api/user/file/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Header("Authorization", string(jwtToken)).
		Transport(fs.transport).
		Method(http.MethodGet).
		ToWriter(os.Stdout).
		Fetch(context.Background())

	if err != nil {
		return err
	}
	return nil
}

type FileToUpdate struct {
	FileName    string `json:"fname,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateFile updates file's name and description by unique title.
// It updates only arguments which are provided.
func (ps *FileService) UpdateFile(title string, args ...string) error {
	jwtToken, err := ps.keyStore.Get()
	if err != nil {
		return err
	}

	file := FileToUpdate{
		FileName:    args[0],
		Description: args[1],
	}

	err = requests.
		URL("/api/user/file/update/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(ps.transport).
		Header("Authorization", string(jwtToken)).
		BodyJSON(&file).
		Method(http.MethodPost).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		log.Println("File info is updated.")
	}
	return nil
}
