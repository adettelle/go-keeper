package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
	"github.com/carlmjohnson/requests"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Customer struct {
	Login    string `json:"login"`
	Password string `json:"pwd"`
}

type CLI struct {
	Login struct {
		Login     string `help:"User login." short:"l"`
		Passwword string `help:"User password." short:"p"`
	} `cmd:"" help:"Login."`

	Passwords struct {
	} `cmd:"" help:"List of passwords."`

	File struct {
		FileName    string `help:"File path." short:"p"`
		Title       string `help:"File title." short:"t"`
		Description string `help:"Description." short:"d"`
		CloudID     string `help:"CloudID." short:"c"` // TODO убрать
	} `cmd:"" help:"File to add."`

	GetFile struct {
		ID string `help:"File id." short:"i"`
	} `cmd:"" help:"File to get."`
}

// go run ./cmd/client login -l 111@121212 -p 1234
// go run ./cmd/client passwords
// go run ./cmd/client file -p './README.md' -t readme -d documentation -c 111
// go run ./cmd/client file -p './screenshot1.png' -t screenshot1 -d png -c 111
// go run ./cmd/client get-file -i 30 > newfile.sql
// TODO listfiles
func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "login":
		login(cli.Login.Login, cli.Login.Passwword)
	case "passwords":
		allpass()
	case "file":
		log.Println("cli.File", cli.File)
		addFile(cli.File.FileName, cli.File.Title, cli.File.Description, cli.File.CloudID)
	case "get-file":
		log.Println("cli.GetFile", cli.GetFile)
		getFile(cli.GetFile.ID)
	}
}

func login(login, password string) {
	customer := Customer{
		Login:    login,
		Password: password,
	}
	headers := http.Header{}
	fileName := "./headers.txt"

	err := requests.
		URL("/api/user/login").
		CopyHeaders(headers).
		CheckStatus(http.StatusOK).
		Host("localhost:8080").
		Scheme("http").
		BodyJSON(&customer).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/login: ", err)
	} else {
		fmt.Println("OK")
		log.Println(headers)

		// открываем файл для записи в конец
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Println("unable to open file: ", err)
			log.Fatal(err)
		}
		defer file.Close()

		headerAuth := headers.Get("Authorization")

		_, err = file.Write([]byte(headerAuth))
		if err != nil {
			log.Println("unable to write to file: ", err)
			log.Fatal(err)
		}
	}
}

type Password struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func allpass() {
	fileBearerName := "./headers.txt"
	file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	if err != nil {
		log.Println("unable to read file: ", err)
		log.Fatal(err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("data: ", string(data))

	var pwds []Password

	err = requests.
		URL("/api/user/passwords").
		Host("localhost:8080").
		Scheme("http").
		Header("Authorization", string(data)).
		ToJSON(&pwds).
		Method(http.MethodGet).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/login: ", err)
	} else {
		fmt.Println("pass OK")

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Title", "Description"})

		for _, pwd := range pwds {
			t.AppendRow([]interface{}{pwd.Title, pwd.Description})
		}
		t.Render()
	}
}

type File struct {
	FileName    string `json:"fname"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CloudID     string `json:"cloudid"` // TODO убрать
}

func addFile(fileName, title, description, cloudID string) {
	fileToAdd := File{
		FileName:    fileName,
		Title:       title,
		Description: description,
		CloudID:     cloudID, // TODO убрать
	}

	fileBearerName := "./headers.txt"
	fileToWriteBearer, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	if err != nil {
		log.Println("unable to read file: ", err)
		log.Fatal(err)
	}

	data, err := io.ReadAll(fileToWriteBearer)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("data: ", string(data))

	type presignURLResponse struct {
		URL string
	}

	var res presignURLResponse

	err = requests.
		URL("/api/user/addfile").
		Host("localhost:8080").
		Scheme("http").
		Header("Authorization", string(data)).
		BodyJSON(&fileToAdd).
		ToJSON(&res).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/addfile: ", err)
		return
	} else {
		fmt.Println("file add OK")
		fmt.Println(res.URL)

		// info, err := os.Stat(fileName)
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// } else {
		// 	log.Println(strconv.Itoa(int(info.Size())))
		// }

		err = uploadFile(fileName, res.URL)
		if err != nil {
			log.Println(err)
			return
		} else {
			log.Printf("File %s uploaded", fileName)
		}
	}

}

type FileToGet struct {
	ID string `json:"id"`
}

func getFile(id string) error {
	fileToGet := FileToGet{
		ID: id,
	}
	fileBearerName := "./headers.txt"
	fileToWriteBearer, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	if err != nil {
		log.Println("unable to open file: ", err)
		log.Fatal(err)
	}

	data, err := io.ReadAll(fileToWriteBearer)
	if err != nil {
		log.Println("unable to read file: ", err)
		log.Fatal(err)
	}

	err = requests.
		URL("/api/user/getfile/"+fileToGet.ID).
		Host("localhost:8080").
		Scheme("http").
		Header("Authorization", string(data)).
		Method(http.MethodGet).
		ToWriter(os.Stdout).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not get file:", err)
		return err
	}

	fmt.Println("got file")
	return nil
}

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

// TODO сделать config в сервере на все строки!!!!!!!!!!!!!!
// TODO сделать ручку GetFile по id, для этого сделать ListFiles - пока без этого можно

// TODO убрать из filename в таблицу полный путь, оставить только название
// TODO почему request upload не сработал?
//
