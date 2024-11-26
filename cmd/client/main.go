package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
	"github.com/carlmjohnson/requests"
	"github.com/jedib0t/go-pretty/v6/table"
)

type CustomerToReg struct {
	Name           string `json:"name"`
	Login          string `json:"login"`
	MasterPassword string `json:"masterpassword"`
}

type CustomerToLogin struct {
	Login    string `json:"login"`
	Password string `json:"pwd"`
}

type CLI struct {
	Register struct {
		Name           string `help:"User name." short:"n"`
		Login          string `help:"User login." short:"l"`
		MasterPassword string `help:"User masterpassword." short:"p"`
	} `cmd:"" help:"Register."`

	Login struct {
		Login     string `help:"User login." short:"l"`
		Passwword string `help:"User password." short:"p"`
	} `cmd:"" help:"Login."`

	Passwords struct {
	} `cmd:"" help:"List of passwords."`

	AddPassword struct {
		Passwword   string `help:"User password." short:"p"`
		Title       string `help:"Password title." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Password to add."`

	AddFile struct {
		FileName    string `help:"File path." short:"p"`
		Title       string `help:"File title." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"File to add."`

	GetFile struct {
		ID string `help:"File id." short:"i"`
	} `cmd:"" help:"File to get."`

	Files struct {
	} `cmd:"" help:"List of files added."`
}

func main() {
	caCert, err := os.ReadFile("./keys/server_cert.pem") //  config.ServerCert
	if err != nil {
		fmt.Printf("error in reading certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// "./keys/client_cert.pem", "./keys/client_privatekey.pem"
	cert, err := tls.LoadX509KeyPair("./keys/client_cert.pem", "./keys/client_privatekey.pem") // config.ClientCert, config.CryptoKey
	if err != nil {
		fmt.Printf("error in loading key pair: %v", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{cert},
		},
	}

	// x := NewHTTPSender(client, fmt.Sprintf("https://%s/api/user/files", "localhost:8080"), "my_secret_key") // config.Address, config.Key

	var cli CLI
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "register":
		register(cli.Register.Name, cli.Register.Login, cli.Register.MasterPassword, transport)
	case "login":
		login(cli.Login.Login, cli.Login.Passwword, transport)
	case "passwords":
		allpass(transport)
	case "add-password":
		addPassword(cli.AddPassword.Passwword, cli.AddPassword.Title, cli.AddPassword.Description, transport)
	case "add-file":
		log.Println("cli.File", cli.AddFile)
		addFile(cli.AddFile.FileName, cli.AddFile.Title, cli.AddFile.Description, transport)
	case "get-file":
		log.Println("cli.GetFile", cli.GetFile)
		getFile(cli.GetFile.ID, transport)
	case "files":
		allFiles(transport)
	}

}

func register(name, login, masterPassword string, transport *http.Transport) {
	customer := CustomerToReg{
		Name:           name,
		Login:          login,
		MasterPassword: masterPassword,
	}

	err := requests.
		URL("/api/user/register").
		CheckStatus(http.StatusOK).
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Method(http.MethodPost).
		BodyJSON(&customer).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not register: ", err)
	}
	log.Println("You've been registered.")
}

func login(login, password string, transport *http.Transport) {
	customer := CustomerToLogin{
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
		Scheme("https").
		Transport(transport).
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

func allpass(transport *http.Transport) {
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
		Scheme("https").
		Transport(transport).
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

type FileToGetAll struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func allFiles(transport *http.Transport) {
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

	var files []FileToGetAll

	err = requests.
		URL("/api/user/files").
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(data)).
		ToJSON(&files).
		Method(http.MethodGet).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not get files: ", err)
	} else {
		fmt.Println("pass OK")

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"ID", "File Name", "Title", "Description"})

		for _, file := range files {
			t.AppendRow([]interface{}{file.ID, file.FileName, file.Title, file.Description})
		}
		t.Render()
	}
}

type Pwd struct {
	Password    string `json:"fname"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func addPassword(password, title, description string, transport *http.Transport) {
	pwdToAdd := Pwd{
		Password:    password,
		Title:       title,
		Description: description,
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

	// type presignURLResponse struct {
	// 	URL string
	// }

	// var res presignURLResponse

	err = requests.
		URL("/api/user/password").
		Host("localhost:8080").
		Scheme("https").
		Header("Authorization", string(data)).
		Transport(transport).
		BodyJSON(&pwdToAdd).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/password: ", err)
		return
	} else {
		fmt.Println("password add OK")

		// info, err := os.Stat(fileName)
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// } else {
		// 	log.Println(strconv.Itoa(int(info.Size())))
		// }

		// err = uploadFile(fileName, res.URL)
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// } else {
		// 	log.Printf("File %s uploaded", fileName)
		// }
	}

}

type File struct {
	FileName    string `json:"fname"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func addFile(fileName, title, description string, transport *http.Transport) {
	fileToAdd := File{
		FileName:    fileName,
		Title:       title,
		Description: description,
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
		Scheme("https").
		Header("Authorization", string(data)).
		Transport(transport).
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

func getFile(id string, transport *http.Transport) error {
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
		Scheme("https").
		Header("Authorization", string(data)).
		Transport(transport).
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
