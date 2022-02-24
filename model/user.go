package model

import (
	"errors"
	"log"
	"os"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrBadSignature = errors.New("bad signature")
	ErrUnauthorized = errors.New("unauthorized")

	jwtSecret = os.Getenv("JWT_SECRET")
	keyFunc   = func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	}
)

type UserClaims struct {
	Login    string
	Password string

	jwt.StandardClaims
}

func SignUser(login, password string) (string, error) {
	claims := &UserClaims{
		Login:    login,
		Password: password,
		StandardClaims: jwt.StandardClaims{
			Issuer: login,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Println("bad signature", err)

		return "", ErrBadSignature
	}

	return signed, nil
}

func ValidateUser(signature string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(signature, &UserClaims{}, keyFunc)

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(*UserClaims); ok {
			return claims, nil
		}
	}

	return nil, ErrUnauthorized
}
