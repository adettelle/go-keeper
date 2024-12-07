package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/mocks"
	"github.com/carlmjohnson/requests"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --------------------------------------- password ---------------------------------------
// ------- Хендлер: PUT /api/user/password
func TestPasswordCreate(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &PassHandlers{
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

	require.Equal(t, wantHTTPStatus, response.Code)
}

func TestPasswordCreateFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	// создаём объект-заглушку
	h := &PassHandlers{
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

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: GET /api/user/passwords
func TestAllPasswords(t *testing.T) {
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
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
	result := response.Result()
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	expectedBody, err := json.Marshal(NewPwdListResponseDTO(pwds))
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}

// ------- Хендлер: GET /api/user/passwords
func TestAllPasswordsFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
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

	require.Equal(t, wantHTTPStatus, response.Code)

}

// ------- Хендлер: GET /api/user/password/{title}
func TestPasswordByTitle(t *testing.T) {
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
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

func TestPasswordByTitleFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
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

func TestPasswordByTitleInvalidTitle(t *testing.T) {
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
		PwdRepo:    pwdRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	pwdRepo.EXPECT().GetPasswordByTitle(gomock.Any(), "title1", login).Return("", nil)

	request, err := requests.
		URL("/api/user/password/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusNotFound

	response := httptest.NewRecorder()
	h.PasswordByTitle(response, request)
	// result := response.Result()
	// defer result.Body.Close()

	// passwordReturned, err := io.ReadAll(result.Body)
	// require.NoError(t, err)
	require.Equal(t, wantHTTPStatus, response.Code)
	// require.Equal(t, "correct_password", string(passwordReturned))
}

// ------- Хендлер: Post /api/user/password/update
func TestPasswordUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
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
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
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
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
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
	ctrl := gomock.NewController(t)

	pwdRepo := mocks.NewMockIPwdRepo(ctrl)

	h := &PassHandlers{
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
