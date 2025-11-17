// Package mware provides middleware for HTTP handlers, including user authentication using JWT.
package mware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/adettelle/go-keeper/internal/jwt"
)

// JwtChecker defines an interface for verifying the validity of a JWT token.
type JwtChecker interface {
	TokenIsValid(ctx context.Context, token string) (bool, error)
}

// AuthMwr is middleware that adds user authentication to an HTTP handler.
// It validates JWT tokens and attaches user information to the request headers.
// It returns a new HTTP handler function that includes authentication logic.
func AuthMwr(h http.HandlerFunc, signKey []byte, jwtChecker JwtChecker) http.HandlerFunc {
	authFn := func(w http.ResponseWriter, r *http.Request) {
		// Extract the Authorization header (expected format: 'Bearer {jwt}').
		authHeaderValue := r.Header.Get("Authorization")
		if authHeaderValue == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Split the header value to retrieve the token.
		bearerToken := strings.Split(authHeaderValue, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Verify the JWT token using the provided signing key.
		cust, ok := jwt.VerifyToken(signKey, bearerToken[1])
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check if the token is valid using the JwtChecker implementation.
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
		// Call the next handler in the chain.
		h.ServeHTTP(w, r)
	}

	// Return the enhanced handler with authentication logic.
	return http.HandlerFunc(authFn)
}
