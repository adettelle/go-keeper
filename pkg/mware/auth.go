package mware

import (
	"net/http"
	"strings"

	"github.com/adettelle/go-keeper/internal/jwt"
)

// AuthMwr добавляет аутентификацию пользователя и возвращает новый http.Handler
func AuthMwr(h http.HandlerFunc, jwtSignKey []byte) http.HandlerFunc {
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

		login, ok := jwt.VerifyToken(jwtSignKey, bearerToken[1])
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Header.Set("x-user", login)
		h.ServeHTTP(w, r) // передали следующей функции, которую мы обрамляем middleware'ом
	}

	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(authFn)
}
