// Client authenticate and authorize users on the remote server; gives access to private data on request.
package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/keyring"
	"github.com/adettelle/go-keeper/internal/client"
	"github.com/adettelle/go-keeper/internal/localstorage"
	"github.com/adettelle/go-keeper/pkg/cert"
	"github.com/alecthomas/kong"
)

type CLI struct {
	Register struct {
		Name           string `help:"User name." short:"n"`
		Login          string `help:"User login." short:"l"`
		MasterPassword string `help:"User masterpassword." short:"p"`
		SignIn         bool   `help:"Sign in now. If present user gets signed in instantly. Otherwise manual login is required (see login command)." short:"s" default:"false"`
	} `cmd:"" help:"Registration with optional authentication."`

	Login struct {
		Login    string `help:"User login." short:"l"`
		Password string `help:"User password." short:"p"`
	} `cmd:"" help:"Login."`

	// ------------ password ------------
	Passwords struct {
	} `cmd:"" help:"Shows list of passwords."`

	AddPassword struct {
		Password    string `help:"User password." short:"p"`
		Title       string `help:"Password uniqe title." short:"t"`
		Description string `help:"Password description." short:"d"`
	} `cmd:"" help:"Adds password."`

	GetPassword struct {
		Title string `help:"Password uniqe title." short:"t"`
	} `cmd:"" help:"Retrieves password by unique title. To retrieve it into file, use: command > filename."`

	// UpdatePassword changes values of password and description by title of password.
	// With no flag value would not change.
	// With empty string ("") the value becomes null.
	UpdatePassword struct {
		Title       string `help:"Password unique title." short:"t"`
		Password    string `help:"User password." short:"p"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Updates password and it's description by unique title. With no flag the value would not change."`

	DeletePassword struct {
		Title string `help:"Password title." short:"t"`
	} `cmd:"" help:"Deletes password by unique title."`

	// ------------ file ------------
	AddFile struct {
		FileName    string `help:"File path." short:"p"`
		Title       string `help:"File unique title." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Adds file."`

	GetFile struct {
		Title string `help:"File title." short:"t"`
	} `cmd:"" help:"Retrieves file by unique title. To retrieve it into file, use: command > filename."`

	Files struct {
	} `cmd:"" help:"Shows list of added files."`

	UpdateFile struct {
		Title       string `help:"File unique title." short:"t"`
		FileName    string `help:"File path." short:"p"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Updates file's name anf description by unique title. With no flag the value would not change."`

	DeleteFile struct {
		Title string `help:"File unique title." short:"t"`
	} `cmd:"" help:"Delete file by unique title."`

	// ------------ card ------------
	Cards struct {
	} `cmd:"" help:"Shows list of added cards."`

	AddCard struct {
		Num         string `help:"Card number, lehgth 16." short:"n"`
		Expire      string `help:"Date of expire, lehgth 4." short:"e"`
		Cvc         string `help:"Card cvc, lehgth 3." short:"c"`
		Title       string `help:"Card title, alphanumeric, min lehgth 4." short:"t"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Adds card."`

	GetCard struct {
		Title string `help:"Card title." short:"t"`
	} `cmd:"" help:"Retrieves card's details by unique title. To retrieve it into file, use command > filename"`

	UpdateCard struct {
		Title       string `help:"Card title." short:"t"`
		Num         string `help:"Card number, lehgth 16." short:"n"`
		Expire      string `help:"Date of expire, lehgth 4." short:"e"`
		Cvc         string `help:"Card cvc, lehgth 3." short:"c"`
		Description string `help:"Description." short:"d"`
	} `cmd:"" help:"Updates card's numbber, date of expire, cvc and description by unique title. With no flag the value would not change."`

	DeleteCard struct {
		Title string `help:"Card title." short:"t"`
	} `cmd:"" help:"Deletes card by unique title."`
}

const service = "gokeeper"

func fileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err == nil {
		// path/to/whatever exists
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return false, nil

	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return false, err

	}
}

type keysPath struct {
	cert       string
	privateKey string
}

func MustInitCerts() keysPath {
	existsDir, err := fileExists("./keys/")
	if err != nil {
		log.Fatal(err)
	}
	if !existsDir {
		err := os.Mkdir("./keys/", 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	pathToClientCert := "./keys/client_cert.pem"
	pathToClientPrivateKey := "./keys/client_privatekey.pem"

	existsCertFile, err := fileExists(pathToClientCert)
	if err != nil {
		log.Fatal(err)
	}
	existsClientPrivateKey, err := fileExists(pathToClientPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	if !existsCertFile || !existsClientPrivateKey {
		if existsCertFile {
			err := os.Remove(pathToClientCert)
			if err != nil {
				log.Fatal(err)
			}
		}
		if existsClientPrivateKey {
			err := os.Remove(pathToClientPrivateKey)
			if err != nil {
				log.Fatal(err)
			}
		}
		cert.MustGenCert("client")
	}
	return keysPath{cert: pathToClientCert, privateKey: pathToClientPrivateKey}
}

func main() {
	keyPaths := MustInitCerts()

	cert, err := tls.LoadX509KeyPair(keyPaths.cert, keyPaths.privateKey)
	if err != nil {
		fmt.Printf("error in loading key pair: %v", err)
		log.Fatal(err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: os.Getenv("DEBUG_RUN_INSECURE") == "true",
			//RootCAs:      caCertPool,
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
		log.Println("---", cli.Register.SignIn)
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
