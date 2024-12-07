// Package jwt provides functionality for generating and verifying JSON Web Tokens (JWTs)
// for authentication and authorization.
package jwt

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the payload of the JWT, including custom fields for user information
// and standard claims from the `jwt` library.
type Claims struct {
	jwt.RegisteredClaims // Standard claims such as expiration time
	Login                string
	UserID               int
}

// TOKEN_EXP defines the expiration duration for generated JWT tokens.
const TOKEN_EXP = time.Hour * 24

// GenerateJwtToken generates a signed JWT token using the provided secret key
// and user details.
func GenerateJwtToken(secret []byte, userLogin string, userID int) (string, error) {
	// Create a new JWT with claims
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		Login:  userLogin, // Custom claim: user login
		UserID: userID,    // Custom claim: user ID
	})

	// Sign the token using the secret key
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		log.Printf("failed to sign jwt: %s\n", err)
		return "", err
	}

	return signedToken, nil
}

type Customer struct {
	ID    int
	Login string
}

// VerifyToken verifies a JWT token, validates its claims, and extracts user information.
//
// Parameters:
//   - secret: The secret key used to validate the token signature.
//   - token: The JWT token to be verified.
//
// Returns:
//   - A `Customer` object containing user ID and login if verification is successful.
//   - A boolean indicating whether the token is valid or not.
func VerifyToken(secret []byte, token string) (Customer, bool) {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		log.Printf("Failed to parse token: %s\n", err)
		return Customer{}, false
	}

	if !jwtToken.Valid {
		return Customer{}, false
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return Customer{}, false
	}

	loginRaw, ok := claims["Login"]
	if !ok {
		return Customer{}, false
	}

	login, ok := loginRaw.(string)
	if !ok {
		return Customer{}, false
	}

	userIDRaw, ok := claims["UserID"]
	if !ok {
		return Customer{}, false
	}
	userID, ok := userIDRaw.(float64)
	if !ok {
		return Customer{}, false
	}

	return Customer{ID: int(userID), Login: login}, true
}
