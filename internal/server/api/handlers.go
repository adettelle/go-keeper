package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/adettelle/go-keeper/internal/jwt"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/config"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ICustomerRepo interface {
	AddCustomer(ctx context.Context, name, login, masterPassword string) (int, error)
	GetCustomerByLogin(ctx context.Context, login string) (*repo.CustomerGetByLogin, error) // добвать в результат id
	VerifyUser(ctx context.Context, login string, pass string) (bool, int)
}

type IPwdRepo interface {
	GetAllPasswords(ctx context.Context, name string) ([]repo.Password, error)
	CreatePassword(ctx context.Context, password, title, description string, login string) error
	UpdatePassword(ctx context.Context, title string, password *string, description *string, userID int) error
	DeletePassword(ctx context.Context, title string, login string) error
	GetPasswordByTitle(ctx context.Context, title string, login string) (string, error)
}

type IFileRepo interface {
	AddFile(ctx context.Context, fileName, title, description, cloudID string, login string) error
	// AddFile2(ctx context.Context, fileName, title, description, cloudID string, login string) error
	GetFileCoudIDByTitle(ctx context.Context, fileID, login string) (string, error)
	GetAllFiles(ctx context.Context, login string) ([]repo.FileToGet, error)
	UpdateFile(ctx context.Context, title string, fileName *string, description *string, userID int) error
	DeleteFileByTitle(ctx context.Context, title string, login string) error
	FileExists(ctx context.Context, title string, custID string) (bool, error)
}

type ICardRepo interface {
	AddCard(ctx context.Context, cardNum, expire, cvc, title, description string, login string) error
	// GetCardByID(ctx context.Context, cardID, login string) (repo.CardGetByID, error)
	GetAllCards(ctx context.Context, login string) ([]repo.CardToGet, error)
	GetCardByTitle(ctx context.Context, cardTitle, login string) (repo.CardGetByTitle, error)
	UpdateCard(ctx context.Context, title string, cardNum *string, expire *string,
		cvc *string, description *string, userID int) error
	DeleteCardByTitle(ctx context.Context, cardTitle, login string) error
}

type IJwtRepo interface {
	AddJwtToken(ctx context.Context, custID int, token string) error // isvalid сначала всега true
}

type IMinioService interface {
	// PresignedPutObject(fileCLoudID string) (string, error)
	GetObject(fileCLoudID string) (io.Reader, error)
	Upload(fileCloudID string, reader io.Reader) error
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

type CustomerHandlers struct {
	CustomerRepo ICustomerRepo
	PwdRepo      IPwdRepo
	FileRepo     IFileRepo
	CardRepo     ICardRepo
	JwtRepo      IJwtRepo
	MinioService IMinioService
	JwtSignKey   []byte
	Config       *config.Config
}

func NewCustomerHandlers(
	customerRepo ICustomerRepo,
	pwdRepo IPwdRepo,
	fileRepo IFileRepo,
	cardRepo ICardRepo,
	jwtRepo IJwtRepo,
	minioService IMinioService,
	jwtSignKey []byte, cfg *config.Config) *CustomerHandlers {
	return &CustomerHandlers{
		CustomerRepo: customerRepo,
		PwdRepo:      pwdRepo,
		FileRepo:     fileRepo,
		CardRepo:     cardRepo,
		JwtRepo:      jwtRepo,
		MinioService: minioService,
		JwtSignKey:   jwtSignKey,
		Config:       cfg,
	}
}

type authRequestDTO struct {
	Login    string `json:"login"`
	Password string `json:"pwd"`
}

type PasswordResponseDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewPwdResponseDTO(pwd repo.Password) *PasswordResponseDTO {
	return &PasswordResponseDTO{
		Title:       pwd.Title,
		Description: pwd.Description,
	}
}

func NewPwdListResponseDTO(pwds []repo.Password) []*PasswordResponseDTO {
	res := []*PasswordResponseDTO{}
	for _, pwd := range pwds {
		res = append(res, NewPwdResponseDTO(pwd))
	}
	return res
}

type fileCreateRequestDTO struct {
	FileName    string `json:"fname"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type fileGetRequestDTO struct {
	//ID          string `json:"id"`
	FileName    string `json:"file_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewFileDTO(file repo.FileToGet) *fileGetRequestDTO {
	return &fileGetRequestDTO{
		//ID:          file.ID,
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

// Login происходит по логину и паролю (masterPassword)
func (ch *CustomerHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	var auth authRequestDTO

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &auth); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok, custID := ch.CustomerRepo.VerifyUser(context.Background(), auth.Login, auth.Password)
	// log.Println("!!!!!!", ok, custID)
	if !ok {
		// log.Println("authlogin, authPass: ", auth.Login, auth.Password)
		w.WriteHeader(http.StatusUnauthorized) // неверная пара логин/пароль
		return
	}

	token, err := jwt.GenerateJwtToken(ch.JwtSignKey, auth.Login, custID)
	if err != nil {
		log.Println("error in generating token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)

	err = ch.JwtRepo.AddJwtToken(context.Background(), custID, token)
	if err != nil {
		log.Println("error in adding jwt token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// service := "gokeeper"
	// user := auth.Login
	// jwtToken := "Bearer " + token

	// err = localstorage.Set(jwtToken)
	// // err = keyring.Set(service, user, jwtToken)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	w.WriteHeader(http.StatusOK)
}

type CustomerRegistrationRequestDTO struct {
	Name           string `json:"name"`
	Login          string `json:"login" validate:"required,email"`
	MasterPassword string `json:"masterpassword"`
}

func (ch *CustomerHandlers) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer
	var customer CustomerRegistrationRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &customer); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(customer.MasterPassword))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	_, err = ch.CustomerRepo.AddCustomer(
		context.Background(), customer.Name, customer.Login, hashStringPassword)
	if err != nil {
		if repo.IsCustomerExistsErr(err) { // TODO HELP!!!!!!!!!!!!!!!!!!!!!!!!!
			log.Printf("error %v in registering user %s", err, customer.Login)
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println("error in adding user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("customer: ", customer)

	w.WriteHeader(http.StatusOK)
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) AllPasswords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userLogin := r.Header.Get("x-user")

	pwds, err := ch.PwdRepo.GetAllPasswords(context.Background(), userLogin) // customer.Login
	if err != nil {
		log.Println("error in getting passwords: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(NewPwdListResponseDTO(pwds))
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

type PwdCreateRequestDTO struct {
	Password    string `json:"pwd"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) PasswordCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")

	var buf bytes.Buffer
	var pwd PwdCreateRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &pwd); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ch.PwdRepo.CreatePassword(
		context.Background(), pwd.Password, pwd.Title, pwd.Description, userLogin)
	if err != nil {
		log.Println("error in adding password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted) // новый пароль принят
}

type pwdUpdateRequestDTO struct {
	// ID          int     `json:"id"`
	// Title       *string `json:"title"`
	Password    *string `json:"pwd"`
	Description *string `json:"description"` // для ссылки значение по умолчанию - nil
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) PasswordUpdate(w http.ResponseWriter, r *http.Request) {
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
	var pwd pwdUpdateRequestDTO
	title := r.PathValue("title")

	// читаем тело запроса
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &pwd); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ch.PwdRepo.UpdatePassword(context.Background(), title, pwd.Password, pwd.Description, custID)
	if err != nil {
		log.Println("error in updating password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted) // новый пароль принят
}

// PasswordDelete удаляет пароль по title
func (ch *CustomerHandlers) PasswordDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pwdTitle := r.PathValue("title")
	userLogin := r.Header.Get("x-user")

	err := ch.PwdRepo.DeletePassword(context.Background(), pwdTitle, userLogin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type PwdRespDTO struct {
	Password string `json:"pwd"`
}

func (ch *CustomerHandlers) PasswordByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")
	pwdTitle := r.PathValue("title")

	pwd, err := ch.PwdRepo.GetPasswordByTitle(context.Background(), pwdTitle, userLogin)
	if err != nil {
		log.Println("error in getting password by title:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(pwd))
	if err != nil {
		log.Println("error in writing resp:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ---------------- file ----------------
/*
type addFileResponseDTO struct {
	URL string
}


// Хендлер доступен только авторизованному пользователю

	func (ch *CustomerHandlers) FileAdd(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		userLogin := r.Header.Get("x-user")

		var buf bytes.Buffer
		var file fileCreateRequestDTO

		// читаем тело запроса
		_, err := buf.ReadFrom(r.Body)
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

		cloudID := uuid.NewString() // генерируем случайную строку опр-го формата

		// убрем из filename полный путь, оставив только название файла
		fileNameWithoutPath := filepath.Base(file.FileName)

		err = ch.FileRepo.AddFile(
			context.Background(), fileNameWithoutPath, file.Title, file.Description, cloudID, userLogin)
		if err != nil {
			log.Println("error in adding file:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url, err := ch.MinioService.PresignedPutObject(cloudID)

		if err != nil {
			log.Println("error in generating upload url:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := addFileResponseDTO{
			URL: url,
		}
		resBody, err := json.Marshal(res)
		if err != nil {
			log.Println("error in marshalling json:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resBody)

		return
	}
*/
func (ch *CustomerHandlers) FileAdd(w http.ResponseWriter, r *http.Request) {
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

	cloudID := uuid.NewString() // генерируем случайную строку опр-го формата

	// убрем из filename полный путь, оставив только название файла
	fileNameWithoutPath := filepath.Base(file.FileName)

	fileExists, err := ch.FileRepo.FileExists(context.Background(), file.Title, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if fileExists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	err = ch.MinioService.Upload(cloudID, r.Body)
	if err != nil {
		log.Println("error in uploading file:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = ch.FileRepo.AddFile(
		context.Background(), fileNameWithoutPath, file.Title, file.Description, cloudID, userLogin)
	if err != nil {
		log.Println("error in adding file:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ch *CustomerHandlers) FileGetByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")
	title := r.PathValue("title")

	fileCLoudID, err := ch.FileRepo.GetFileCoudIDByTitle(context.Background(), title, userLogin)
	if err != nil {
		log.Println("error in getting file by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	obj, err := ch.MinioService.GetObject(fileCLoudID)
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

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) AllFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userLogin := r.Header.Get("x-user")

	// TODO CHECK лишнее?????????????
	/*
		customer, err := ch.CustomerRepo.GetCustomerByLogin(context.Background(), userLogin)
		log.Println("user from get customer by login:", *customer)
		if err != nil {
			log.Println("error in getting user by login:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if customer == nil {
			log.Println("customer == nil")
			w.WriteHeader(http.StatusNotFound) // это значит, нет такого пользователя
			return
		}
	*/

	files, err := ch.FileRepo.GetAllFiles(context.Background(), userLogin) // customer.Login
	if err != nil {
		log.Println("error in getting files: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// log.Println("files: ", files)

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
	FileName    *string `json:"fname"` // для ссылки значение по умолчанию - nil
	Description *string `json:"description"`
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) FileUpdate(w http.ResponseWriter, r *http.Request) {
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

	err = ch.FileRepo.UpdateFile(context.Background(), title, file.FileName, file.Description, custID)
	if err != nil {
		log.Println("error in updating password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted) // новый пароль принят
}

func (ch *CustomerHandlers) FileDeleteByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	title := r.PathValue("title")
	userLogin := r.Header.Get("x-user")

	err := ch.FileRepo.DeleteFileByTitle(context.Background(), title, userLogin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// ---------------- card ----------------

type cardCreateRequestDTO struct {
	Num         string `json:"num" validate:"required,credit_card,len=16"`
	Expire      string `json:"expires_at" validate:"required,numeric,len=4"`
	Cvc         string `json:"cvc" validate:"required,numeric,len=3"`
	Title       string `json:"title" validate:"required,alphanumunicode,min=1"` // len дб > 4
	Description string `json:"description"`
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) CardAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")

	var buf bytes.Buffer
	var card cardCreateRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &card); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validate.Struct(card)
	if err != nil {
		log.Println("error in validating:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ch.CardRepo.AddCard(
		context.Background(), card.Num, card.Expire, card.Cvc, card.Title, card.Description, userLogin)
	if err != nil {
		log.Println("error in adding card:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

type cardGetRequestDTO struct {
	Num         string `json:"num"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewCardDTO(card repo.CardToGet) *cardGetRequestDTO {
	return &cardGetRequestDTO{
		Num:         card.Num,
		Title:       card.Title,
		Description: card.Description,
	}
}

func NewCardListDTO(cards []repo.CardToGet) []*cardGetRequestDTO {
	res := []*cardGetRequestDTO{}
	for _, card := range cards {
		res = append(res, NewCardDTO(card))
	}
	return res
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) AllCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userLogin := r.Header.Get("x-user")

	cards, err := ch.CardRepo.GetAllCards(context.Background(), userLogin)
	if err != nil {
		log.Println("error in getting card by login:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(NewCardListDTO(cards))
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

type cardGetByTitleRequestDTO struct {
	Num         string `json:"num"`
	Expire      string `json:"expires_at"`
	Cvc         string `json:"cvc"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewCardGetByTitleDTO(card repo.CardGetByTitle) *cardGetByTitleRequestDTO {
	return &cardGetByTitleRequestDTO{
		Num:         card.Num,
		Expire:      card.Expire,
		Cvc:         card.Cvc,
		Title:       card.Title,
		Description: card.Description,
	}
}

// Хендлер доступен только авторизованному пользователю
// CardGetByTitle shows num, expireAt, cvc and description by cadd title
func (ch *CustomerHandlers) CardGetByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userLogin := r.Header.Get("x-user")
	title := r.PathValue("title")

	card, err := ch.CardRepo.GetCardByTitle(context.Background(), title, userLogin)
	if err != nil {
		log.Println("error in getting card:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(NewCardGetByTitleDTO(card))
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

// TODO Надо ли здесь тоже ставить ограничение ommitempty или только в клиенте????????????
type cardUpdateRequestDTO struct {
	Num         *string `json:"num"`
	Expire      *string `json:"expires_at"`
	Cvc         *string `json:"cvc"`
	Description *string `json:"description"`
}

func (ch *CustomerHandlers) CardUpdate(w http.ResponseWriter, r *http.Request) {
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
	var card cardUpdateRequestDTO
	title := r.PathValue("title")

	// читаем тело запроса
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &card); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ch.CardRepo.UpdateCard(context.Background(), title, card.Num, card.Expire,
		card.Cvc, card.Description, custID)
	if err != nil {
		log.Println("error in updating card:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (ch *CustomerHandlers) CardDeleteByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cardTitle := r.PathValue("title")
	userLogin := r.Header.Get("x-user")

	err := ch.CardRepo.DeleteCardByTitle(context.Background(), cardTitle, userLogin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
