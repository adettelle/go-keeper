package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/keyring"
	"github.com/adettelle/go-keeper/internal/client"
	"github.com/adettelle/go-keeper/internal/localstorage"
	"github.com/alecthomas/kong"
)

type CLI struct {
	Register struct {
		Name           string `help:"User name." short:"n"`
		Login          string `help:"User login." short:"l"`
		MasterPassword string `help:"User masterpassword." short:"p"`
		SignIn         bool   `help:"Sign in now, if present user gets signed in instantly. Otherwise manual login is required (see login command)." short:"s"`
	} `cmd:"" help:"Registration with auntication needs flag -s (empty flag means true)."`

	Login struct {
		Login    string `help:"User login." short:"l"`
		Password string `help:"User password." short:"p"`
	} `cmd:"" help:"Login."`

	// ------------ password ------------
	Passwords struct {
	} `cmd:"" help:"List of passwords."`

	AddPassword struct {
		Password    string `help:"User password." short:"p"`
		Title       string `help:"Password uniqe title." short:"t"`
		Description string `help:"Password description." short:"d"`
	} `cmd:"" help:"Password to add."`

	GetPassword struct {
		Title string `help:"Password uniqe title." short:"t"`
	} `cmd:"" help:"Password to get by uniqe title."`

	// UpdatePassword меняет поля (pwd, description) по title пароля
	// если в json не передать поле, то оно не измениться
	// если передать пустую строку "" - то поле станет пустым
	// TODO а как поменять сам title????/
	UpdatePassword struct { // HELP TODO CHECK
		// ID          int    `help:"Password id, by id the password is deleted." short:"i"` // TODO не нужен?
		Title       string `help:"Password uniqe title." short:"t"`
		Password    string `help:"User password." short:"p"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Password to update by uniqe title."`

	DeletePassword struct {
		Title string `help:"Password title." short:"t"`
	} `cmd:"" help:"Password to delete by title."`

	// ------------ file ------------
	AddFile struct {
		FileName    string `help:"File path." short:"p"`
		Title       string `help:"File uniqe title." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"File to add."`

	GetFile struct {
		Title string `help:"File title." short:"t"`
	} `cmd:"" help:"File to get by uniqe title."`
	// TODO как написать, что сохранить под новым именем можно через >

	Files struct {
	} `cmd:"" help:"List of files added."`

	UpdateFile struct {
		Title       string `help:"File uniqe title." short:"t"`
		FileName    string `help:"File path." short:"p"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"File to update."`

	DeleteFile struct {
		Title string `help:"File uniqe title." short:"t"`
	} `cmd:"" help:"File to delete by uniqe title."`

	// ------------ card ------------
	Cards struct {
	} `cmd:"" help:"List of cards."`

	AddCard struct {
		Num         string `help:"Card number, lehgth 16." short:"n"`
		Expire      string `help:"Date of expire, lehgth 4." short:"e"`
		Cvc         string `help:"Card cvc, lehgth 3." short:"c"`
		Title       string `help:"Card title, alphanumeric, min lehgth 4." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Card to add."`

	GetCard struct {
		Title string `help:"Card title." short:"t"`
	} `cmd:"" help:"Card to get by uniqe title."`

	UpdateCard struct {
		Title       string `help:"Card title." short:"t"`
		Num         string `help:"Card number, lehgth 16." short:"n"`
		Expire      string `help:"Date of expire, lehgth 4." short:"e"`
		Cvc         string `help:"Card cvc, lehgth 3." short:"c"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Card to update."`

	DeleteCard struct {
		Title string `help:"Card title." short:"t"`
	} `cmd:"" help:"Card to delete by uniqe title."`
}

const service = "gokeeper"

func main() {
	caCert, err := os.ReadFile("./keys/server_cert.pem") //  config.ServerCert TODO
	if err != nil {
		fmt.Printf("error in reading certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// config.ClientCert, config.CryptoKey TODO
	cert, err := tls.LoadX509KeyPair("./keys/client_cert.pem", "./keys/client_privatekey.pem")
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

	keyStore := localstorage.NewKeyStore(&keyring.Config{
		ServiceName:      service,
		AllowedBackends:  nil, // []keyring.BackendType{keyring.FileBackend}
		FilePasswordFunc: keyring.TerminalPrompt,
		// AllowedBackends: []keyring.BackendType{keyring.KWalletBackend},
		FileDir: "~/",
	})

	cardService := client.NewCardService(transport, keyStore)
	passwordService := client.NewPasswordService(transport, keyStore)
	fileService := client.NewFileService(transport, keyStore)
	userService := client.NewUserService(transport, keyStore)

	switch ctx.Command() {
	case "register":
		AssertNoError(userService.Register(cli.Register.Name, cli.Register.Login,
			cli.Register.MasterPassword, cli.Register.SignIn))
	case "login":
		AssertNoError(userService.LogIn(cli.Login.Login, cli.Login.Password))

	case "passwords":
		AssertNoError(passwordService.AllPass())
	case "add-password":
		AssertNoError(passwordService.AddPassword(cli.AddPassword.Password,
			cli.AddPassword.Title, cli.AddPassword.Description))
	case "get-password":
		AssertNoError(passwordService.GetPasswordByTitle(cli.GetPassword.Title))
	case "update-password":
		AssertNoError(passwordService.UpdatePassword(cli.UpdatePassword.Title,
			cli.UpdatePassword.Password, cli.UpdatePassword.Description))
	case "delete-password":
		AssertNoError(passwordService.DeletePasswordByTitle(cli.DeletePassword.Title))

	case "add-file":
		AssertNoError(fileService.AddFile(cli.AddFile.FileName,
			cli.AddFile.Title, cli.AddFile.Description))
	case "get-file":
		AssertNoError(fileService.GetFile(cli.GetFile.Title))
	case "files":
		AssertNoError(fileService.AllFiles())
	case "update-file":
		AssertNoError(fileService.UpdateFile(cli.UpdateFile.Title, cli.UpdateFile.FileName,
			cli.UpdateFile.Description, cli.UpdatePassword.Description))
	case "delete-file":
		AssertNoError(fileService.DeleteFileByTitle(cli.DeleteFile.Title))

	case "cards":
		AssertNoError(cardService.AllCards())
	case "add-card":
		AssertNoError(cardService.AddCard(cli.AddCard.Num, cli.AddCard.Expire,
			cli.AddCard.Cvc, cli.AddCard.Title, cli.AddCard.Description))
	case "get-card":
		AssertNoError(cardService.GetCardByTitle(cli.GetCard.Title))
	case "update-card":
		AssertNoError(cardService.UpdateCard(cli.UpdateCard.Title, cli.UpdateCard.Num, cli.UpdateCard.Expire,
			cli.UpdateCard.Cvc, cli.UpdateCard.Description))
	case "delete-card":
		AssertNoError(cardService.DeleteCardByTitle(cli.DeleteCard.Title))
	}
}

func AssertNoError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
