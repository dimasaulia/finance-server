package jwt

import (
	"errors"
	"finance/provider/configuration"
	"fmt"
	"time"

	j "github.com/golang-jwt/jwt/v5"
)

type TokenData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Fullname string `json:"fullname"`
}

func GenerateJWT(d TokenData) (string, error) {
	var secretKey = []byte(configuration.ENV.GetString("JWT_SECRET"))
	t := j.NewWithClaims(j.SigningMethodHS256, j.MapClaims{
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"data": d,
	})

	tokenString, err := t.SignedString(secretKey)
	if err != nil {
		return "resp", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func VerifyJWT(tokenString string) error {
	var secretKey = []byte(configuration.ENV.GetString("JWT_SECRET"))

	token, err := j.Parse(tokenString, func(token *j.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
