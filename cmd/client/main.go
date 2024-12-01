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

	"github.com/adettelle/go-keeper/internal/localstorage"
	"github.com/alecthomas/kong"
	"github.com/carlmjohnson/requests"
	"github.com/jedib0t/go-pretty/v6/table"
)

type CustomerToLogin struct {
	Login    string `json:"login"`
	Password string `json:"pwd"`
}

type CLI struct {
	Register struct {
		Name           string `help:"User name." short:"n"`
		Login          string `help:"User login." short:"l"`
		MasterPassword string `help:"User masterpassword." short:"p"`
		SignIn         bool   `help:"If you want to sign in, flag should be true." short:"s"`
	} `cmd:"" help:"Registration with auntication needs flag s (true)."`

	Login struct {
		Login    string `help:"User login." short:"l"`
		Password string `help:"User password." short:"p"`
	} `cmd:"" help:"Login."`

	Passwords struct {
	} `cmd:"" help:"List of passwords."`

	AddPassword struct {
		Password    string `help:"User password." short:"p"`
		Title       string `help:"Password title." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Password to add."`

	GetPassword struct { // HELP
		// ID string `help:"Password id." short:"i"`
		// Password   string `help:"User password." short:"p"`
		Title string `help:"Password title." short:"t"`
		// Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Password to get by title."`

	// UpdatePassword меняет поля (pwd, title, description) по id пароля
	// если в json не передать поле, то оно не измениться
	// если передать пустую строку "" - то поле станет пустым
	UpdatePassword struct { // HELP TODO
		ID          int    `help:"Password id, by id the password is deleted." short:"i"`
		Password    string `help:"User password." short:"p"`
		Title       string `help:"Password title." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Password to update by id."`

	DeletePassword struct {
		Title string `help:"Password title." short:"t"`
	} `cmd:"" help:"Password to delete by title."`

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

	DeleteFile struct {
		CloudID string `help:"File cloudID." short:"c"`
	} `cmd:"" help:"File to delete by cloudID."`

	Cards struct {
	} `cmd:"" help:"List of cards."`

	AddCard struct {
		Num         string `help:"Card number, lehgth 16." short:"n"`
		Expire      string `help:"Date of expire, lehgth 4." short:"e"`
		Cvc         string `help:"Card cvc, lehgth 3." short:"c"`
		Title       string `help:"Card title, alphanumeric, min lehgth 4." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Card to add."`

	DeleteCard struct {
		ID string `help:"Card id." short:"i"`
	} `cmd:"" help:"Card to delete by id."`
}

// TODO Update card.Description, card.Title
// TODO ПОчему попадают номера карт < 16 цифр
func main() {
	caCert, err := os.ReadFile("./keys/server_cert.pem") //  config.ServerCert
	if err != nil {
		fmt.Printf("error in reading certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

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

	var cli CLI
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "register":
		register(cli.Register.Name, cli.Register.Login,
			cli.Register.MasterPassword, cli.Register.SignIn, transport)
	case "login":
		logIn(cli.Login.Login, cli.Login.Password, transport)
	case "passwords":
		allpass(transport)
	case "add-password":
		addPassword(cli.AddPassword.Password, cli.AddPassword.Title, cli.AddPassword.Description, transport)
	case "get-password":
		getPasswordByTitle(cli.GetPassword.Title, transport)
	case "update-password":
		updatePassword(transport, cli.UpdatePassword.ID, cli.UpdatePassword.Password,
			cli.UpdatePassword.Title, cli.UpdatePassword.Description)
	case "delete-password":
		deletePasswordByTitle(cli.DeletePassword.Title, transport)
	case "add-file":
		log.Println("cli.File", cli.AddFile)
		addFile(cli.AddFile.FileName, cli.AddFile.Title, cli.AddFile.Description, transport)
	case "get-file":
		log.Println("cli.GetFile", cli.GetFile)
		getFile(cli.GetFile.ID, transport)
	case "files":
		allFiles(transport)
	case "delete-file":
		deleteFileByCloudID(cli.DeleteFile.CloudID, transport)
	case "cards":
		allcards(transport)
	case "add-card":
		addCard(cli.AddCard.Num, cli.AddCard.Expire, cli.AddCard.Cvc, cli.AddCard.Title, cli.AddCard.Description, transport)
		// case "get-card":
		// 	getPasswordByTitle(cli.GetCard.ID, transport)
	case "delete-card":
		deleteCardByID(cli.DeleteCard.ID, transport)
	}
}

type CustomerToReg struct {
	Name           string `json:"name"`
	Login          string `json:"login"`
	MasterPassword string `json:"masterpassword"`
	Authentication bool   `json:"signin"`
}

func register(name, login, masterPassword string, signIn bool, transport *http.Transport) {
	customer := CustomerToReg{
		Name:           name,
		Login:          login,
		MasterPassword: masterPassword,
		Authentication: signIn,
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
		return
	}

	if customer.Authentication {
		logIn(customer.Login, customer.MasterPassword, transport)
	}
	// log.Println("You've been registered.")
}

func logIn(login, password string, transport *http.Transport) {
	customer := CustomerToLogin{
		Login:    login,
		Password: password,
	}

	headers := http.Header{}
	// fileName := "./headers.txt"

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
	} // else {
	// fmt.Println("OK")
	// log.Println(headers)

	// открываем файл для записи в конец
	// file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	// if err != nil {
	// 	log.Println("unable to open file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	jwtToken := headers.Get("Authorization")
	err = localstorage.Set(jwtToken)
	if err != nil {
		log.Fatal(err)
	}
	// _, err = file.Write([]byte(headerAuth))
	// if err != nil {
	// 	log.Println("unable to write to file: ", err)
	// 	log.Fatal(err)
	// }
	// }
}

type PasswordToGet struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func getPasswordByTitle(title string, transport *http.Transport) {
	// service := "gokeeper" // TODO
	// user := ""
	// // password := "secret"
	// jwtToken, err := keyring.Get(service, user)
	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	// fileBearerName := "./headers.txt"
	// file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	var pwd string // PasswordToGet
	// var buf bytes.Buffer

	err = requests.
		URL("/api/user/password/"+title).
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(jwtToken)). // data
		Method(http.MethodGet).
		//ToBytesBuffer(&buf).
		// ToString(&pwd). // BodyReader(strings.NewReader(pwd)).
		// ToWriter().
		ToJSON(&pwd).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/password/"+title, err)
	} else {
		log.Println("pass OK")
		log.Println("Password is:", pwd)
		// t := table.NewWriter()
		// t.SetOutputMirror(os.Stdout)
		// t.AppendHeader(table.Row{"ID", "Title", "Description"})

		// t.AppendRow([]interface{}{pwd.ID, pwd.Title, pwd.Description})
		// t.Render()
	}
}

func allpass(transport *http.Transport) {
	// fileBearerName := "./headers.txt"
	// file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("data: ", string(data))
	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	var pwds []PasswordToGet

	err = requests.
		URL("/api/user/passwords").
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(jwtToken)). // data
		ToJSON(&pwds).
		Method(http.MethodGet).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/passwords", err)
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

type PasswordToUpdate struct {
	ID          int    `json:"id"`
	Password    string `json:"pwd,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func updatePassword(transport *http.Transport, id int, args ...string) {
	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	// fileBearerName := "./headers.txt"
	// file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("data: ", string(data))

	// var pwd PasswordToUpdate
	pwd := PasswordToUpdate{
		ID:          id,
		Password:    args[0],
		Title:       args[1],
		Description: args[2],
	}

	err = requests.
		URL("/api/user/password/update").
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(jwtToken)). // data
		BodyJSON(&pwd).
		Method(http.MethodPost).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/password/update", err)
	} else {
		fmt.Println("pass OK")

		// t := table.NewWriter()
		// t.SetOutputMirror(os.Stdout)
		// t.AppendHeader(table.Row{"ID", "Title", "Description"})

		// for _, pwd := range pwds {
		// 	t.AppendRow([]interface{}{pwd.ID, pwd.Title, pwd.Description})
		// }
		// t.Render()
	}
}

// TODO HELP чтение bearer из файла!!!!!!!!!!!!!
// надо сделать таблицу token

func deletePasswordByTitle(title string, transport *http.Transport) {
	// fileBearerName := "./headers.txt"
	// file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	err = requests.
		URL("/api/user/delete/"+title).
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(jwtToken)). // data
		Method(http.MethodDelete).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/delete/"+title, err)
	} else {
		log.Println("pass deleted")
	}
}

type Pwd struct {
	Password    string `json:"pwd"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func addPassword(password, title, description string, transport *http.Transport) {
	pwdToAdd := Pwd{
		Password:    password,
		Title:       title,
		Description: description,
	}

	// fileBearerName := "./headers.txt"
	// fileToWriteBearer, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer fileToWriteBearer.Close()

	// data, err := io.ReadAll(fileToWriteBearer)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("data: ", string(data))

	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	// type presignURLResponse struct {
	// 	URL string
	// }

	// var res presignURLResponse

	err = requests.
		URL("/api/user/password").
		Host("localhost:8080").
		Scheme("https").
		Header("Authorization", string(jwtToken)). // data
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

// -------------------- files --------------------

type FileToGetAll struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func allFiles(transport *http.Transport) {
	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	// fileBearerName := "./headers.txt"
	// file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("data: ", string(data))

	var files []FileToGetAll

	err = requests.
		URL("/api/user/files").
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(jwtToken)). // data
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

func deleteFileByCloudID(cloudID string, transport *http.Transport) {
	// fileBearerName := "./headers.txt"
	// file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	err = requests.
		URL("/api/user/delete/file/"+cloudID).
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(jwtToken)). // data
		Method(http.MethodDelete).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/delete/file/"+cloudID, err)
	} else {
		log.Println("file deleted")
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

	// fileBearerName := "./headers.txt"
	// fileToWriteBearer, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer fileToWriteBearer.Close()

	// data, err := io.ReadAll(fileToWriteBearer)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("data: ", string(data))

	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	type presignURLResponse struct {
		URL string
	}

	var res presignURLResponse

	err = requests.
		URL("/api/user/addfile").
		Host("localhost:8080").
		Scheme("https").
		Header("Authorization", string(jwtToken)). // data
		Transport(transport).
		BodyJSON(&fileToAdd).
		ToJSON(&res).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/addfile: ", err)
		return
	} else {
		fmt.Printf("File %s added.", fileName)
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
			log.Printf("File %s uploaded.", fileName)
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
	// fileBearerName := "./headers.txt"
	// fileToWriteBearer, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to open file: ", err)
	// 	log.Fatal(err)
	// }
	// defer fileToWriteBearer.Close()

	// data, err := io.ReadAll(fileToWriteBearer)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }

	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	err = requests.
		URL("/api/user/getfile/"+fileToGet.ID).
		Host("localhost:8080").
		Scheme("https").
		Header("Authorization", string(jwtToken)). // data
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

// -------------------- cards --------------------

type CardToGet struct {
	Num         string `json:"num"` // validate:"required,credit_card"
	Title       string `json:"title"`
	Description string `json:"description"`
}

// could not connect to localhost:8080/api/user/cards ErrHandler: unexpected end of JSON input
func allcards(transport *http.Transport) {
	// fileBearerName := "./headers.txt"
	// file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("data: ", string(data))

	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	var cards []CardToGet

	err = requests.
		URL("/api/user/cards").
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(jwtToken)). // data
		ToJSON(&cards).
		Method(http.MethodGet).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/cards", err)
	} else {
		fmt.Println("pass OK")

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Num", "Title", "Description"})

		for _, card := range cards { // HELP TODO номер карты показывать не надо????
			// cvc, exp показывать не надо, Num - последние 4 цифры
			t.AppendRow([]interface{}{card.Num, card.Title, card.Description})
		}
		t.Render()
	}
}

type CardToAdd struct {
	Num         string `json:"num"`
	Expire      string `json:"expires_at"`
	Cvc         string `json:"cvc"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func addCard(num, expire, cvc, title, description string, transport *http.Transport) {
	cardToAdd := CardToAdd{
		Num:         num,
		Expire:      expire,
		Cvc:         cvc,
		Title:       title,
		Description: description,
	}

	// fileBearerName := "./headers.txt"
	// fileToWriteBearer, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer fileToWriteBearer.Close()

	// data, err := io.ReadAll(fileToWriteBearer)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	err = requests.
		URL("/api/user/addcard").
		Host("localhost:8080").
		Scheme("https").
		Header("Authorization", string(jwtToken)). // data
		Transport(transport).
		BodyJSON(&cardToAdd).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/addcard: ", err)
		return
	} else {
		fmt.Println("card added OK")
	}
}

func deleteCardByID(cardID string, transport *http.Transport) {
	// fileBearerName := "./headers.txt"
	// file, err := os.OpenFile(fileBearerName, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("unable to read file: ", err)
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	jwtToken, err := localstorage.Get()
	if err != nil {
		log.Fatal(err)
	}

	err = requests.
		URL("/api/user/delete/card/"+cardID).
		Host("localhost:8080").
		Scheme("https").
		Transport(transport).
		Header("Authorization", string(jwtToken)). // data
		Method(http.MethodDelete).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to localhost:8080/api/user/delete/"+cardID, err)
	} else {
		log.Println("card deleted")
	}
}
