package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/adettelle/go-keeper/internal/encryption"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/config"
	"github.com/go-playground/validator/v10"
)

type CardHandlers struct {
	CardRepo ICardRepo
	SignKey  []byte
	Config   *config.Config
}

func NewCardHandlers(cardRepo ICardRepo, signKey []byte, cfg *config.Config) *CardHandlers {
	return &CardHandlers{
		CardRepo: cardRepo,
		SignKey:  signKey,
		Config:   cfg,
	}
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

type ICardRepo interface {
	AddCard(ctx context.Context, cardNum, expire, cvc, title, description string, login string) error
	GetAllCards(ctx context.Context, login string) ([]repo.CardToGet, error)
	GetCardByTitle(ctx context.Context, cardTitle, login string) (*repo.CardGetByTitle, error)
	UpdateCard(ctx context.Context, title string, cardNum *string, expire *string,
		cvc *string, description *string, userID int) error
	DeleteCardByTitle(ctx context.Context, cardTitle, login string) error
}

type cardCreateRequestDTO struct {
	Num         string `json:"num" validate:"required,credit_card,len=16"`
	Expire      string `json:"expires_at" validate:"required,numeric,len=4"`
	Cvc         string `json:"cvc" validate:"required,numeric,len=3"`
	Title       string `json:"title" validate:"required,min=1"`
	Description string `json:"description"`
}

func (ch *CardHandlers) CardAdd(w http.ResponseWriter, r *http.Request) {
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

	encryptedNum, err := encryption.AESEncrypt(card.Num, ch.SignKey)
	if err != nil {
		log.Println("error in encrypting card number:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	encryptedExpire, err := encryption.AESEncrypt(card.Expire, ch.SignKey)
	if err != nil {
		log.Println("error in encrypting card expire date:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	encryptedCvc, err := encryption.AESEncrypt(card.Cvc, ch.SignKey)
	if err != nil {
		log.Println("error in encrypting card cvc:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = ch.CardRepo.AddCard(context.Background(), encryptedNum, encryptedExpire, encryptedCvc,
		card.Title, card.Description, userLogin)

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

func (ch *CardHandlers) AllCards(w http.ResponseWriter, r *http.Request) {
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

// CardGetByTitle shows num, expireAt, cvc and description by card title
func (ch *CardHandlers) CardGetByTitle(w http.ResponseWriter, r *http.Request) {
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
	if card == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decryptedNum, err := encryption.AESDecrypt(card.Num, ch.SignKey)
	if err != nil {
		log.Println("error in decrypting card number:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	decryptedExpire, err := encryption.AESDecrypt(card.Expire, ch.SignKey)
	if err != nil {
		log.Println("error in decrypting card expires date:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	decryptedCvc, err := encryption.AESDecrypt(card.Cvc, ch.SignKey)
	if err != nil {
		log.Println("error in decrypting card cvc:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	card.Num = decryptedNum
	card.Expire = decryptedExpire
	card.Cvc = decryptedCvc

	resp, err := json.Marshal(NewCardGetByTitleDTO(*card))
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

type cardUpdateRequestDTO struct {
	Num         *string `json:"num,omitempty" validate:"omitnil,omitempty,credit_card,len=16"`   // validate:"omitnil,omitempty,credit_card,len=16"
	Expire      *string `json:"expires_at,omitempty" validate:"omitempty,omitnil,numeric,len=4"` //  validate:"omitempty,omitnil,numeric,len=4"
	Cvc         *string `json:"cvc,omitempty" validate:"omitempty,omitnil,numeric,len=3"`        //  validate:"omitempty,omitnil,numeric,len=3"
	Description *string `json:"description,omitempty"`
}

func (ch *CardHandlers) CardUpdate(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("!!!!!!!!!!!", string(buf.Bytes()))
	// return
	err = validate.Struct(card)
	if err != nil {
		log.Println("error in validating:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var encryptedNum, encryptedExpire, encryptedCvc *string
	if card.Num != nil {
		res, err := encryption.AESEncrypt(*card.Num, ch.SignKey)
		if err != nil {
			log.Println("error in encrypting card number:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		encryptedNum = &res
	}
	if card.Expire != nil {
		res, err := encryption.AESEncrypt(*card.Expire, ch.SignKey)
		if err != nil {
			log.Println("error in encrypting card expires date:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		encryptedExpire = &res
	}
	if card.Cvc != nil {
		res, err := encryption.AESEncrypt(*card.Cvc, ch.SignKey)
		if err != nil {
			log.Println("error in encrypting card cvc:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		encryptedCvc = &res
	}

	err = ch.CardRepo.UpdateCard(context.Background(), title, encryptedNum, encryptedExpire,
		encryptedCvc, card.Description, custID) // card.Num, card.Expire, card.Cvc
	if err != nil {
		log.Println("error in updating card:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (ch *CardHandlers) CardDeleteByTitle(w http.ResponseWriter, r *http.Request) {
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
