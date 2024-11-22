package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adettelle/go-keeper/internal/jwt"
	"github.com/adettelle/go-keeper/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// type Customer struct {
// 	Name     string
// 	Login    string
// 	Password string
// }

// type MockCustomerHandlers struct {
// 	CustomerRepo ICustomerRepo
// 	JwtSignKey   []byte // убрать в config TODO []byte("my_secret_key")
// }

// type MockCusomerRepo struct {
// 	Customers map[string]*repo.Customer
// }

// func (r *MockCusomerRepo) AddCustomer(ctx context.Context, name, email, masterPassword string) error {
// 	return fmt.Errorf("err")
// }

// func (r *MockCusomerRepo) GetCustomerByLogin(ctx context.Context, login string) (*repo.Customer, error) {
// 	return nil, fmt.Errorf("err")
// }

/*
// ------- Хендлер: POST /api/user/register
func TestRegisterCustomer(t *testing.T) {

	h := &CustomerHandlers{
		CustomerRepo: &MockCusomerRepo{
			Customers: make(map[string]*repo.Customer),
		},
		JwtSignKey: []byte("my_key"),
	}

	cust1 := Customer{
		Name:     "sobakevich",
		Login:    "sobakevich@aaa.com",
		Password: "12f6c041f585bc63e7979af17d02a7e8c026f173dee8cea6e7b195290a879e46",
	}

	jsonData, err := json.Marshal(cust1)
	require.NoError(t, err)

	reqURL := "/api/user/register"
	reqBody := string(jsonData)
	wantHTTPStatus := 200

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()
	h.RegisterCustomer(response, request)
	result := response.Result()
	require.Equal(t, wantHTTPStatus, response.Code)

	token, err := jwt.GenerateJwtToken(h.JwtSignKey, cust1.Login)
	require.NoError(t, err)
	require.Equal(t, response.Header().Get("Authorization"), "Bearer "+token)
	defer result.Body.Close()

	// повторный запрос с тем же юзером
	request, err = http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response = httptest.NewRecorder()
	h.RegisterCustomer(response, request)
	result = response.Result()

	require.Equal(t, http.StatusConflict, response.Code)
	defer result.Body.Close()
}
*/

// ------- Хендлер: POST /api/user/register
func TestRegisterCustomer(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockICustomerRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CustomerRepo: repo,
		JwtSignKey:   []byte("my_key"),
	}

	cust1 := CustomerDTO{
		Name:           "sobakevich",
		Email:          "sobakevich@aaa.com",
		MasterPassword: "my_pass",
	}

	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(cust1.MasterPassword))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	repo.EXPECT().AddCustomer(gomock.Any(), cust1.Name, cust1.Email, hashStringPassword).Return(nil)

	jsonData, err := json.Marshal(cust1)
	require.NoError(t, err)

	reqURL := "/api/user/register"
	reqBody := string(jsonData)
	wantHTTPStatus := 200

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()
	h.RegisterCustomer(response, request)
	result := response.Result()
	require.Equal(t, wantHTTPStatus, response.Code)

	token, err := jwt.GenerateJwtToken(h.JwtSignKey, cust1.Email)
	require.NoError(t, err)
	require.Equal(t, response.Header().Get("Authorization"), "Bearer "+token)
	defer result.Body.Close()
}

// ------- Хендлер: POST /api/user/register
func TestRegisterCustomerAddingExistingUser(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockICustomerRepo(ctrl)

	// создаём объект-заглушку
	h := &CustomerHandlers{
		CustomerRepo: repo,
		JwtSignKey:   []byte("my_key"),
	}

	cust1 := CustomerDTO{
		Name:           "sobakevich",
		Email:          "sobakevich@aaa.com",
		MasterPassword: "my_pass",
	}

	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(cust1.MasterPassword))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	repo.EXPECT().AddCustomer(gomock.Any(), cust1.Name, cust1.Email, hashStringPassword).Return(fmt.Errorf("409"))

	jsonData, err := json.Marshal(cust1)
	require.NoError(t, err)

	reqURL := "/api/user/register"
	reqBody := string(jsonData)

	// повторный запрос с тем же юзером
	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()
	h.RegisterCustomer(response, request)

	require.Equal(t, http.StatusInternalServerError, response.Code)
}
