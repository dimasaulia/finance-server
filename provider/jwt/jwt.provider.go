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
	j.RegisteredClaims
}

func GenerateJWT(d TokenData) (string, error) {

	var secretKey = []byte(configuration.ENV.GetString("JWT_SECRET"))
	d.RegisteredClaims.ExpiresAt = j.NewNumericDate(time.Now().Add(1 * time.Hour))
	d.RegisteredClaims.IssuedAt = j.NewNumericDate(time.Now())

	t := j.NewWithClaims(j.SigningMethodHS256, d)

	tokenString, err := t.SignedString(secretKey)
	if err != nil {
		return "resp", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func VerifyJWT(tokenString string) (*TokenData, error) {
	var secretKey = []byte(configuration.ENV.GetString("JWT_SECRET"))

	// Parse and validate the JWT token
	token, err := j.ParseWithClaims(tokenString, &TokenData{}, func(token *j.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*j.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(*TokenData); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
