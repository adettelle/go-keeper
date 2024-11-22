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
}

// go run ./cmd/client login -l 111@121212 -p 1234
// go run ./cmd/client passwords
func main() {
	var cli CLI
	// для 111@121212
	// var bearer = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6IjExMUAxMjEyMTIifQ.PJ6yEXvLeqeWgXM6AgsOCju6ZOJio0xdbsj1-yuQxJ0"
	// "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4"
	ctx := kong.Parse(&cli)
	// kong.Name("go-keeper"),
	// kong.Description("A shell-like app."),
	// kong.UsageOnError(),
	// kong.ConfigureHelp(kong.HelpOptions{
	// 	Compact: true,
	// 	Summary: true,
	// }))

	// if ctx.Command() == "login" {
	// 	fmt.Println(cli.Login.Login, cli.Login.Passwword)
	// 	login()
	// }
	switch ctx.Command() {
	case "login":
		login(cli.Login.Login, cli.Login.Passwword)
	case "passwords":
		allpass()
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
	fileName := "./headers.txt"
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
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
