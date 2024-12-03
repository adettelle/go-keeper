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

type FileToGetAll struct {
	// ID          string `json:"id"`
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
		// fmt.Println("Could not get files:", err)
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
		// fmt.Println("Could not connect to localhost:8080/api/user/delete/file/"+cloudID, err)
		return err
	} else {
		log.Println("File is deleted.")
	}
	return nil
}

type File struct {
	FileName    string `json:"fname"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

/*
func (fs *FileService) AddFile(fileName, title, description string) error {
	fileToAdd := File{
		FileName:    fileName,
		Title:       title,
		Description: description,
	}

	jwtToken, err := fs.keyStore.Get()
	if err != nil {
		log.Fatal(err)
	}

	type presignURLResponse struct {
		URL string
	}

	var res presignURLResponse

	err = requests.
		URL("/api/user/file").
		Host(settings.ServerURL).
		Scheme("https").
		Header("Authorization", string(jwtToken)).
		Transport(fs.transport).
		BodyJSON(&fileToAdd).
		ToJSON(&res).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		// fmt.Println("Could not connect to localhost:8080/api/user/addfile: ", err)
		return err
	} else {
		err = uploadFile(fileName, res.URL)
		if err != nil {
			return err
		} else {
			log.Printf("File %s is uploaded.", fileName)
		}
	}
	return nil
}
*/

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
	// fileToGet := FileToGet{
	// 	Title: title,
	// }

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
		// fmt.Println("Could not get file:", err)
		return err
	}

	return nil
}

type FileToUpdate struct {
	FileName    string `json:"fname,omitempty"`
	Description string `json:"description,omitempty"`
}

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
		// fmt.Println("could not connect to localhost:8080/api/user/file/update", err)
		return err
	} else {
		log.Println("File info is updated.")
	}
	return nil
}

/*
func uploadFile(filePath string, url string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPut, url, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "multipart/form-data")
	client := &http.Client{}
	_, err = client.Do(request)
	return err
}
*/
