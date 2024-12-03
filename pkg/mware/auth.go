package mware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/adettelle/go-keeper/internal/jwt"
)

type JwtChecker interface {
	TokenIsValid(ctx context.Context, token string) (bool, error)
}

// AuthMwr добавляет аутентификацию пользователя и возвращает новый http.Handler
func AuthMwr(h http.HandlerFunc, jwtSignKey []byte, jwtChecker JwtChecker) http.HandlerFunc {
	authFn := func(w http.ResponseWriter, r *http.Request) {
		// получаем http header вида 'Bearer {jwt}'
		authHeaderValue := r.Header.Get("Authorization")
		if authHeaderValue == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// проверяем доступы
		bearerToken := strings.Split(authHeaderValue, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		cust, ok := jwt.VerifyToken(jwtSignKey, bearerToken[1])
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		isValid, err := jwtChecker.TokenIsValid(context.Background(), bearerToken[1])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Set("x-user", cust.Login)
		r.Header.Set("x-user-id", strconv.Itoa(cust.ID))
		// log.Println("login, id:", cust.Login, cust.ID)
		h.ServeHTTP(w, r) // передали следующей функции, которую мы обрамляем middleware'ом
	}

	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(authFn)
}
