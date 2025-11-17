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

	"github.com/adettelle/go-keeper/internal/encryption"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/mocks"
	"github.com/carlmjohnson/requests"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------- Хендлер: PUT /api/user/card
func TestCardAdd(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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
	encryptedNum, err := encryption.AESEncrypt(card.Num, h.SignKey)
	require.NoError(t, err)
	encryptedExpire, err := encryption.AESEncrypt(card.Expire, h.SignKey)
	require.NoError(t, err)
	encryptedCvc, err := encryption.AESEncrypt(card.Cvc, h.SignKey)
	require.NoError(t, err)
	cardRepo.EXPECT().AddCard(gomock.Any(), encryptedNum, encryptedExpire, encryptedCvc,
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

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: PUT /api/user/card
func TestCardAddFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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
	encryptedNum, err := encryption.AESEncrypt(card.Num, h.SignKey)
	require.NoError(t, err)
	encryptedExpire, err := encryption.AESEncrypt(card.Expire, h.SignKey)
	require.NoError(t, err)
	encryptedCvc, err := encryption.AESEncrypt(card.Cvc, h.SignKey)
	require.NoError(t, err)

	cardRepo.EXPECT().AddCard(gomock.Any(), encryptedNum, encryptedExpire, encryptedCvc,
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

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: GET /api/user/cards
func TestAllCards(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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
	result := response.Result()
	require.Equal(t, wantHTTPStatus, response.Code)

	defer result.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	expectedBody, err := json.Marshal(NewCardListDTO(cards))
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}

// ------- Хендлер: GET /api/user/cards
func TestAllCardsFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: GET /api/user/card/{title}
func TestCardByTitle(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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

	cardRepo.EXPECT().GetCardByTitle(gomock.Any(), "title1", login).Return(&card, nil)

	encryptedNum, err := encryption.AESEncrypt(card.Num, h.SignKey)
	require.NoError(t, err)
	encryptedExpire, err := encryption.AESEncrypt(card.Expire, h.SignKey)
	require.NoError(t, err)
	encryptedCvc, err := encryption.AESEncrypt(card.Cvc, h.SignKey)
	require.NoError(t, err)
	card.Num = encryptedNum
	card.Expire = encryptedExpire
	card.Cvc = encryptedCvc

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
	result := response.Result()
	defer result.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, wantHTTPStatus, response.Code)
	expectedBody, err := json.Marshal(NewCardGetByTitleDTO(card))
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}

func TestCardByTitleFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
	}

	login := "Ane"
	userID := 123

	cardRepo.EXPECT().GetCardByTitle(gomock.Any(), "title1", login).Return(nil, fmt.Errorf("DB error"))

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

	require.Equal(t, wantHTTPStatus, response.Code)
}

func TestCardByTitleInvalidTitle(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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

	cardRepo.EXPECT().GetCardByTitle(gomock.Any(), "title1", login).Return(nil, nil)

	request, err := requests.
		URL("/api/user/card/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		BodyJSON(&card).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusNotFound

	response := httptest.NewRecorder()
	h.CardGetByTitle(response, request)

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Post /api/user/card/update/{title}
func TestCardUpdate(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	cardRepo := mocks.NewMockICardRepo(ctrl)

	// создаём объект-заглушку
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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
	encryptedNum, err := encryption.AESEncrypt(*card.Num, h.SignKey)
	require.NoError(t, err)
	encryptedExpire, err := encryption.AESEncrypt(*card.Expire, h.SignKey)
	require.NoError(t, err)
	encryptedCvc, err := encryption.AESEncrypt(*card.Cvc, h.SignKey)
	require.NoError(t, err)

	cardRepo.EXPECT().UpdateCard(gomock.Any(), title, &encryptedNum, &encryptedExpire,
		&encryptedCvc, card.Description, userID).Return(nil)

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
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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
	encryptedNum, err := encryption.AESEncrypt(*card.Num, h.SignKey)
	require.NoError(t, err)
	encryptedExpire, err := encryption.AESEncrypt(*card.Expire, h.SignKey)
	require.NoError(t, err)
	encryptedCvc, err := encryption.AESEncrypt(*card.Cvc, h.SignKey)
	require.NoError(t, err)

	cardRepo.EXPECT().UpdateCard(gomock.Any(), title, &encryptedNum, &encryptedExpire,
		&encryptedCvc, card.Description, userID).Return(fmt.Errorf("DB error"))

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
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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
	h := &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  []byte("my_super_secret_key"),
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
