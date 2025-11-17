package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adettelle/go-keeper/internal/jwt"
	"github.com/adettelle/go-keeper/mocks"
	"github.com/carlmjohnson/requests"
	"github.com/golang/mock/gomock"
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
		SignKey:      []byte("my_key"),
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
	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: POST /api/user/register
func TestRegisterCustomerAddingExistingUser(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	custRepo := mocks.NewMockICustomerRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CustomerRepo: custRepo,
		SignKey:      []byte("my_key"),
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
		SignKey:      []byte("my_key"),
	}

	auth := authRequestDTO{
		Login:    "user1@user1.com",
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

	cust, ok := jwt.VerifyToken(h.SignKey, bearerToken[1])
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
		SignKey:      []byte("my_key"),
	}

	auth := authRequestDTO{
		Login:    "user1@user1.com",
		Password: "correct_pass",
	}
	custRepo.EXPECT().VerifyUser(gomock.Any(), auth.Login, auth.Password).Return(false, 0)

	request, err := requests.
		URL("/api/user/login").
		Method(http.MethodPost).
		BodyJSON(&auth).Request(context.Background())
	require.NoError(t, err)

	response := httptest.NewRecorder()
	h.Login(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
}
