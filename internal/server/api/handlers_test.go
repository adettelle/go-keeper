package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/adettelle/go-keeper/internal/jwt"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/mocks"
	"github.com/carlmjohnson/requests"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --------------------------------------- customer ---------------------------------------
// ------- Хендлер: POST /api/user/register
func TestRegisterCustomer(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	custRepo := mocks.NewMockICustomerRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CustomerRepo: custRepo,
		JwtSignKey:   []byte("my_key"),
	}

	cust1 := CustomerRegistrationRequestDTO{
		Name:           "sobakevich",
		Login:          "sobakevich@aaa.com",
		MasterPassword: "my_pass",
	}

	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(cust1.MasterPassword))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	custRepo.EXPECT().AddCustomer(gomock.Any(), cust1.Name, cust1.Login, hashStringPassword).Return(0, nil)

	request, err := requests.
		URL("/api/user/register").
		Method(http.MethodPost).
		BodyJSON(&cust1).Request(context.Background())

	require.NoError(t, err)

	wantHTTPStatus := 200

	response := httptest.NewRecorder()
	h.RegisterCustomer(response, request)
	result := response.Result()
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()
}

// ------- Хендлер: POST /api/user/register
func TestRegisterCustomerAddingExistingUser(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	custRepo := mocks.NewMockICustomerRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CustomerRepo: custRepo,
		JwtSignKey:   []byte("my_key"),
	}

	cust1 := CustomerRegistrationRequestDTO{
		Name:           "sobakevich",
		Login:          "sobakevich@aaa.com",
		MasterPassword: "my_pass",
	}

	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(cust1.MasterPassword))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	custRepo.EXPECT().AddCustomer(gomock.Any(), cust1.Name,
		cust1.Login, hashStringPassword).Return(0, fmt.Errorf("409"))

	// повторный запрос с тем же юзером
	request, err := requests.
		URL("/api/user/register").
		Method(http.MethodPost).
		BodyJSON(&cust1).Request(context.Background())

	require.NoError(t, err)

	response := httptest.NewRecorder()
	h.RegisterCustomer(response, request)

	require.Equal(t, http.StatusInternalServerError, response.Code)
}

// ------- Хендлер: POST /api/user/login
func TestLogin(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	custRepo := mocks.NewMockICustomerRepo(ctrl)
	jwtRepo := mocks.NewMockIJwtRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CustomerRepo: custRepo,
		JwtRepo:      jwtRepo,
		JwtSignKey:   []byte("my_key"),
	}

	auth := authRequestDTO{
		Login:    "user1",
		Password: "correct_pass",
	}
	custRepo.EXPECT().VerifyUser(gomock.Any(), auth.Login, auth.Password).Return(true, 1000)
	jwtRepo.EXPECT().AddJwtToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	request, err := requests.
		URL("/api/user/login").
		Method(http.MethodPost).
		BodyJSON(&auth).Request(context.Background())
	require.NoError(t, err)

	response := httptest.NewRecorder()
	h.Login(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	token := response.Header().Get("Authorization")
	require.NotEmpty(t, token)

	// проверяем доступы
	bearerToken := strings.Split(token, " ")
	require.Len(t, bearerToken, 2)
	require.Equal(t, bearerToken[0], "Bearer")

	cust, ok := jwt.VerifyToken(h.JwtSignKey, bearerToken[1])
	require.True(t, ok)
	require.Equal(t, cust.ID, 1000)
	require.Equal(t, cust.Login, auth.Login)
}

func TestLoginFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	custRepo := mocks.NewMockICustomerRepo(ctrl)
	jwtRepo := mocks.NewMockIJwtRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CustomerRepo: custRepo,
		JwtRepo:      jwtRepo,
		JwtSignKey:   []byte("my_key"),
	}

	auth := authRequestDTO{
		Login:    "user1",
		Password: "correct_pass",
	}
	custRepo.EXPECT().VerifyUser(gomock.Any(), auth.Login, auth.Password).Return(false, 0)
	// jwtRepo.EXPECT().AddJwtToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/login").
		Method(http.MethodPost).
		BodyJSON(&auth).Request(context.Background())
	require.NoError(t, err)

	response := httptest.NewRecorder()
	h.Login(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
	// token := response.Header().Get("Authorization")
	// require.NotEmpty(t, token)

	// проверяем доступы
	// bearerToken := strings.Split(token, " ")
	// require.Len(t, bearerToken, 2)
	// require.Equal(t, bearerToken[0], "Bearer")

	// cust, ok := jwt.VerifyToken(h.JwtSignKey, bearerToken[1])
	// require.True(t, ok)
	// require.Equal(t, cust.ID, 1000)
	// require.Equal(t, cust.Login, auth.Login)
}

// --------------------------------------- password ---------------------------------------
// ------- Хендлер: PUT /api/user/password
func TestPasswordCreate(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwd := PwdCreateRequestDTO{
		Password:    "password",
		Title:       "vk_pass",
		Description: "pass for vk",
	}
	pwdRepo.EXPECT().CreatePassword(gomock.Any(), pwd.Password,
		pwd.Title, pwd.Description, login).Return(nil)

	request, err := requests.
		URL("/api/user/password").
		Method(http.MethodPut).
		BodyJSON(&pwd).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusAccepted // 202

	response := httptest.NewRecorder()
	h.PasswordCreate(response, request)
	result := response.Result() // TODO HELP что ждем в result???
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()
}

func TestPasswordCreateFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwd := PwdCreateRequestDTO{
		Password:    "password",
		Title:       "vk_pass",
		Description: "pass for vk",
	}
	pwdRepo.EXPECT().CreatePassword(gomock.Any(), pwd.Password,
		pwd.Title, pwd.Description, login).Return(fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/password").
		Method(http.MethodPut).
		BodyJSON(&pwd).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.PasswordCreate(response, request)
	//result := response.Result()
	require.Equal(t, wantHTTPStatus, response.Code)

	//defer result.Body.Close()
}

// ------- Хендлер: GET /api/user/passwords
func TestAllPasswords(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwds := []repo.Password{
		{
			Title:       "title1",
			Description: "description1",
		},
		{
			Title:       "title2",
			Description: "description2",
		},
	}
	pwdRepo.EXPECT().GetAllPasswords(gomock.Any(), login).Return(pwds, nil)

	request, err := requests.
		URL("/api/user/passwords").
		Method(http.MethodGet).
		Header("x-user", login). // TOD HELP разве здесь не  должна быть структура, в которую падает get результат?
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.AllPasswords(response, request)
	result := response.Result() // TODO HELP что ждем в result???
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()
}

// ------- Хендлер: GET /api/user/passwords
func TestAllPasswordsFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwdRepo.EXPECT().GetAllPasswords(gomock.Any(), login).Return(nil, fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/passwords").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.AllPasswords(response, request)
	//result := response.Result()
	require.Equal(t, wantHTTPStatus, response.Code)

	//defer result.Body.Close()
}

// ------- Хендлер: GET /api/user/password/{title}
func TestPasswordByTitle(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwdRepo.EXPECT().GetPasswordByTitle(gomock.Any(), "title1", login).Return("correct_password", nil)

	request, err := requests.
		URL("/api/user/password/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.PasswordByTitle(response, request)
	result := response.Result()
	defer result.Body.Close()

	passwordReturned, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Equal(t, wantHTTPStatus, response.Code)
	require.Equal(t, "correct_password", string(passwordReturned))
}

// ------- Хендлер: GET /api/user/password/{title}
func TestPasswordByTitleFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwdRepo.EXPECT().GetPasswordByTitle(gomock.Any(), "title1", login).Return("", fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/password/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.PasswordByTitle(response, request)

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Post /api/user/password/update
func TestPasswordUpdate(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	password := "password"
	title := "title1"
	description := "pass for vk"

	pwd := pwdUpdateRequestDTO{
		Password:    &password,
		Description: &description,
	}

	pwdRepo.EXPECT().UpdatePassword(gomock.Any(), title, pwd.Password, pwd.Description, userID).Return(nil)

	request, err := requests.
		URL("/api/user/password/update/"+title).
		Method(http.MethodPost).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		BodyJSON(&pwd).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusAccepted

	response := httptest.NewRecorder()
	h.PasswordUpdate(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Post /api/user/password/update
func TestPasswordUpdateFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	password := "password"
	description := "pass for vk"
	pwd := pwdUpdateRequestDTO{
		Password:    &password,
		Description: &description,
	}

	pwdRepo.EXPECT().UpdatePassword(gomock.Any(), "title1", pwd.Password,
		pwd.Description, userID).Return(fmt.Errorf("Update repo fail"))

	request, err := requests.
		URL("/api/user/password/update/title1").
		Method(http.MethodPost).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		BodyJSON(&pwd).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.PasswordUpdate(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Post /api/user/password/{title}
func TestPasswordDelete(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwdRepo.EXPECT().DeletePassword(gomock.Any(), "title1", login).Return(nil)

	request, err := requests.
		URL("/api/user/password/title1").
		Method(http.MethodDelete).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.PasswordDelete(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Post /api/user/password/{title}
func TestPasswordDeleteFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwdRepo.EXPECT().DeletePassword(gomock.Any(), "title1", login).Return(fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/password/title1").
		Method(http.MethodDelete).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.PasswordDelete(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// --------------------------------------- file ---------------------------------------
/*
// ------- Хендлер: PUT /api/user/file
func TestFileAdd(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)
	minioService := mocks.NewMockIMinioService(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		FileRepo:     fileRepo,
		MinioService: minioService,
		JwtSignKey:   []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	file := fileCreateRequestDTO{
		FileName:    "./file1.png",
		Title:       "png",
		Description: "png file",
	}
	fileNameWithoutPath := filepath.Base(file.FileName)

	fileRepo.EXPECT().FileExists(gomock.Any(), file.Title, userID).Return(false, nil)

	minioService.EXPECT().Upload(gomock.Any(), io.Reader).Return(nil) // cloudID ????? TODO

	fileRepo.EXPECT().AddFile(gomock.Any(), fileNameWithoutPath, file.Title,
		file.Description, gomock.Any(), login).Return(nil)

	// var url string = "http://aaa.bbb.ccc"
	// minioService.EXPECT().PresignedPutObject(gomock.Any()).Return(url, nil)

	request, err := requests.
		URL("/api/user/file").
		Method(http.MethodPut).
		BodyJSON(&file).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.FileAdd(response, request)

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, wantHTTPStatus, response.Code)
	expectedBody, err := json.Marshal(addFileResponseDTO{
		URL: url,
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}
*/
/*
// ------- Хендлер: PUT /api/user/file
func TestFileAddFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)
	minioService := mocks.NewMockIMinioService(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		FileRepo:     fileRepo,
		MinioService: minioService,
		JwtSignKey:   []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	file := fileCreateRequestDTO{
		FileName:    "./file1.png",
		Title:       "png",
		Description: "png file",
	}
	fileNameWithoutPath := filepath.Base(file.FileName)

	fileRepo.EXPECT().AddFile(gomock.Any(), fileNameWithoutPath, file.Title,
		file.Description, gomock.Any(), login).Return(fmt.Errorf("DB error"))

	// TODO MinioClient
	var url string = "http://aaa.bbb.ccc"
	minioService.EXPECT().PresignedPutObject(gomock.Any()).Return(url, nil)

	request, err := requests.
		URL("/api/user/file").
		Method(http.MethodPut).
		BodyJSON(&file).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.FileAdd(response, request)

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, wantHTTPStatus, response.Code)
	expectedBody, err := json.Marshal(addFileResponseDTO{
		URL: url,
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}
*/
// ------- Хендлер: GET /api/user/files
func TestAllFiles(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	files := []repo.FileToGet{
		{
			FileName:    "./file1.png",
			Title:       "png1",
			Description: "description1",
		},
		{
			FileName:    "./file2.png",
			Title:       "png2",
			Description: "description2",
		},
	}
	fileRepo.EXPECT().GetAllFiles(gomock.Any(), login).Return(files, nil)

	request, err := requests.
		URL("/api/user/files").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.AllFiles(response, request)

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, wantHTTPStatus, response.Code)

	expectedBody, err := json.Marshal([]fileGetRequestDTO{
		{
			FileName:    "./file1.png",
			Title:       "png1",
			Description: "description1",
		},
		{
			FileName:    "./file2.png",
			Title:       "png2",
			Description: "description2",
		},
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}

// ------- Хендлер: GET /api/user/files
func TestAllFilesFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	fileRepo.EXPECT().GetAllFiles(gomock.Any(), login).Return(nil, fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/files").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.AllFiles(response, request)

	require.Equal(t, wantHTTPStatus, response.Code)
}

/*
// ------- Хендлер: GET /api/user/file/{title}
func TestFileByTitle(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)
	minioService := mocks.NewMockIMinioService(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		FileRepo:     fileRepo,
		MinioService: minioService,
		JwtSignKey:   []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	// card := repo.CardGetByTitle{
	// 	Num:         "6011000990139424",
	// 	Expire:      "0130",
	// 	Cvc:         "123",
	// 	Title:       "bank1",
	// 	Description: "card for bank1",
	// }

	fileRepo.EXPECT().GetFileCoudIDByTitle(gomock.Any(), "title1", login).Return(gomock.Any(), nil)
	minioService.EXPECT().GetObject(gomock.Any()).Return(gomock.Any(), nil) // io.Reader ??????

	request, err := requests.
		URL("/api/user/file/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		ToWriter(os.Stdout). // ????????????????
		// BodyJSON(&card).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.CardGetByTitle(response, request)
	result := response.Result()
	defer result.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, wantHTTPStatus, response.Code)
	expectedBody, err := json.Marshal(NewCardGetByTitleDTO(card))
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}
*/

// ------- Хендлер: Delete /api/user/file/{title}
func TestFIleDelete(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	fileRepo.EXPECT().DeleteFileByTitle(gomock.Any(), "title1", login).Return(nil)

	request, err := requests.
		URL("/api/user/file/title1").
		Method(http.MethodDelete).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.FileDeleteByTitle(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Post /api/user/file/update/{title}
func TestFileUpdate(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	title := "title1"
	fname := "file1.png"
	description := "png file"

	file := fileUpdateRequestDTO{
		FileName:    &fname,
		Description: &description,
	}

	fileRepo.EXPECT().UpdateFile(gomock.Any(), title, file.FileName, file.Description, userID).Return(nil)

	request, err := requests.
		URL("/api/user/file/update/"+title).
		Method(http.MethodPost).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		BodyJSON(&file).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusAccepted

	response := httptest.NewRecorder()
	h.FileUpdate(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

func TestFileUpdateFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	title := "title1"
	fname := "file1.png"
	description := "png file"

	file := fileUpdateRequestDTO{
		FileName:    &fname,
		Description: &description,
	}

	fileRepo.EXPECT().UpdateFile(gomock.Any(), title, file.FileName, file.Description,
		userID).Return(fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/file/update/"+title).
		Method(http.MethodPost).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		BodyJSON(&file).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.FileUpdate(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// --------------------------------------- card ---------------------------------------
// ------- Хендлер: PUT /api/user/card
func TestCardAdd(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	card := cardCreateRequestDTO{
		Num:         "6011000990139424",
		Expire:      "0130",
		Cvc:         "123",
		Title:       "bank1",
		Description: "card for bank1",
	}
	cardRepo.EXPECT().AddCard(gomock.Any(), card.Num, card.Expire, card.Cvc,
		card.Title, card.Description, login).Return(nil)

	request, err := requests.
		URL("/api/user/card").
		Method(http.MethodPut).
		BodyJSON(&card).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusAccepted // 202

	response := httptest.NewRecorder()
	h.CardAdd(response, request)
	result := response.Result() // TODO HELP что ждем в result???
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()
}

// ------- Хендлер: PUT /api/user/card
func TestCardAddFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	card := cardCreateRequestDTO{
		Num:         "6011000990139424",
		Expire:      "0130",
		Cvc:         "123",
		Title:       "bank1",
		Description: "card for bank1",
	}
	cardRepo.EXPECT().AddCard(gomock.Any(), card.Num, card.Expire, card.Cvc,
		card.Title, card.Description, login).Return(fmt.Errorf("error in DB"))

	request, err := requests.
		URL("/api/user/card").
		Method(http.MethodPut).
		BodyJSON(&card).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.CardAdd(response, request)
	result := response.Result() // TODO HELP что ждем в result???
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()
}

// ------- Хендлер: GET /api/user/cards
func TestAllCards(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	cards := []repo.CardToGet{
		{
			Num:         "6011000990139424",
			Title:       "bank1",
			Description: "description1",
		},
		{
			Num:         "3530111333300000",
			Title:       "banc2",
			Description: "description2",
		},
	}
	cardRepo.EXPECT().GetAllCards(gomock.Any(), login).Return(cards, nil)

	request, err := requests.
		URL("/api/user/cards").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.AllCards(response, request)
	result := response.Result() // TODO HELP что ждем в result???
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()
}

// ------- Хендлер: GET /api/user/cards
func TestAllCardsFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	cardRepo.EXPECT().GetAllCards(gomock.Any(), login).Return(nil, fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/cards").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.AllCards(response, request)
	result := response.Result() // TODO HELP что ждем в result???
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()
}

// ------- Хендлер: GET /api/user/card/{title}
func TestCardByTitle(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	card := repo.CardGetByTitle{
		Num:         "6011000990139424",
		Expire:      "0130",
		Cvc:         "123",
		Title:       "bank1",
		Description: "card for bank1",
	}

	cardRepo.EXPECT().GetCardByTitle(gomock.Any(), "title1", login).Return(card, nil)

	request, err := requests.
		URL("/api/user/card/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		BodyJSON(&card).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.CardGetByTitle(response, request)
	result := response.Result() // TODO HELP что ждем в result???
	defer result.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, wantHTTPStatus, response.Code)
	expectedBody, err := json.Marshal(NewCardGetByTitleDTO(card))
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}

// ------- Хендлер: GET /api/user/card/{title}
func TestCardByTitleFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	card := repo.CardGetByTitle{}

	cardRepo.EXPECT().GetCardByTitle(gomock.Any(), "title1", login).Return(card, fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/card/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.CardGetByTitle(response, request)
	result := response.Result() // TODO HELP что ждем в result???
	defer result.Body.Close()

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Post /api/user/card/update/{title}
func TestCardUpdate(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	title := "title1"
	num := "6331101999990016"
	expire := "0141"
	cvc := "222"
	description := "pass for vk"

	card := cardUpdateRequestDTO{
		Num:         &num,
		Expire:      &expire,
		Cvc:         &cvc,
		Description: &description,
	}

	cardRepo.EXPECT().UpdateCard(gomock.Any(), title, card.Num, card.Expire,
		card.Cvc, card.Description, userID).Return(nil)

	request, err := requests.
		URL("/api/user/card/update/"+title).
		Method(http.MethodPost).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		BodyJSON(&card).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusAccepted

	response := httptest.NewRecorder()
	h.CardUpdate(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

func TestCardUpdateFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	title := "title1"
	num := "6331101999990016"
	expire := "0141"
	cvc := "222"
	description := "pass for vk"

	card := cardUpdateRequestDTO{
		Num:         &num,
		Expire:      &expire,
		Cvc:         &cvc,
		Description: &description,
	}

	cardRepo.EXPECT().UpdateCard(gomock.Any(), title, card.Num, card.Expire,
		card.Cvc, card.Description, userID).Return(fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/card/update/"+title).
		Method(http.MethodPost).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		BodyJSON(&card).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.CardUpdate(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Delete /api/user/card/{title}
func TestCardDelete(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	cardRepo.EXPECT().DeleteCardByTitle(gomock.Any(), "title1", login).Return(nil)

	request, err := requests.
		URL("/api/user/card/title1").
		Method(http.MethodDelete).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.CardDeleteByTitle(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Delete /api/user/card/{title}
func TestCardDeleteFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CardRepo:   cardRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	cardRepo.EXPECT().DeleteCardByTitle(gomock.Any(), "title1", login).Return(fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/card/title1").
		Method(http.MethodDelete).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.CardDeleteByTitle(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}
