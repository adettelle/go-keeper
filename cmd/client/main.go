package main

import (
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
		CloudID     string `help:"CloudID." short:"c"`
	} `cmd:"" help:"File to add."`
}

// go run ./cmd/client login -l 111@121212 -p 1234
// go run ./cmd/client passwords
// go run ./cmd/client file -p './README.md' -t readme -d documentation -c 111
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
			// fmt.Printf("%s\t%s\t%s\n", pwd.Title, pwd.Description)
			t.AppendRow([]interface{}{pwd.Title, pwd.Description})
		}
		t.Render()
	}
}

type File struct {
	FileName    string `json:"fname"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CloudID     string `json:"cloudid"`
}

func addFile(fileName, title, description, cloudID string) {
	fileToAdd := File{
		FileName:    fileName,
		Title:       title,
		Description: description,
		CloudID:     cloudID,
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

	// var pwds []Password

	err = requests.
		URL("/api/user/addfile").
		Host("localhost:8080").
		Scheme("http").
		Header("Authorization", string(data)).
		BodyJSON(&fileToAdd).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/addfile: ", err)
	} else {
		fmt.Println("file add OK")

		// t := table.NewWriter()
		// t.SetOutputMirror(os.Stdout)
		// t.AppendHeader(table.Row{"Title", "Description"})

		// for _, pwd := range pwds {
		// 	// fmt.Printf("%s\t%s\t%s\n", pwd.Title, pwd.Description)
		// 	t.AppendRow([]interface{}{pwd.Title, pwd.Description})
		// }
		// t.Render()
	}
}
