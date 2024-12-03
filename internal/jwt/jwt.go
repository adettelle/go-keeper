package jwt

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Login  string
	UserID int
}

const TOKEN_EXP = time.Hour * 24

func GenerateJwtToken(secret []byte, userLogin string, userID int) (string, error) {
	// создаём payload
	// claims := jwt.MapClaims{
	// 	"login":       userLogin,
	// 	"customer_id": userID,
	// }
	// создаём jwt и указываем payload
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		Login:  userLogin,
		UserID: userID,
	})

	// получаем подписанный токен
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		log.Printf("failed to sign jwt: %s\n", err)
		return "", err
	}
	// log.Println("Result token: " + signedToken)

	return signedToken, nil
}

// VerifyToken — функция, которая выполняет аутентификацию и авторизацию пользователя.
// token — JWT пользователя.
// если у пользователь ввел правильные данные, и у него есть необходимая привилегия -
// возвращаем true и логин пользователя, иначе - false
type Customer struct {
	ID    int
	Login string
}

// HELP TODO *Customer или Customer???????????
func VerifyToken(secret []byte, token string) (Customer, bool) { //  HELP TODO как secret используется?????!!!!!
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secret, nil // почему здесь secret, если д.б. interface????
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
	// log.Println("claims[Login]:", claims["Login"])
	if !ok {
		return Customer{}, false
	}

	login, ok := loginRaw.(string)
	if !ok {
		return Customer{}, false
	}
	// log.Println("login:", login)
	// log.Println("claims[UserID]:", claims["UserID"])

	userIDRaw, ok := claims["UserID"]
	if !ok {
		return Customer{}, false
	}
	//log.Println("userIDRaw:", reflect.TypeOf(userIDRaw), userIDRaw)
	userID, ok := userIDRaw.(float64)
	if !ok {
		return Customer{}, false
	}

	//log.Println("claims[UserID]:", claims["UserID"])
	return Customer{ID: int(userID), Login: login}, true
}
