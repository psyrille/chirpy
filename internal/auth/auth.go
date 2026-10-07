package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {

	match, _, err := argon2id.CheckHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil

}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	currentDate := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(currentDate.UTC()),
		ExpiresAt: jwt.NewNumericDate(currentDate.Add(time.Duration(3600) * time.Second)),
		Subject:   userID.String(),
	})

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	return parsedID, err

}

func GetBearerToken(header http.Header) (string, error) {

	if len(header.Values("Authorization")) <= 0 {
		err := errors.New("Authorization header does not exist")
		return "", err
	}

	bearerToken := http.Header.Get(header, "Authorization")
	if bearerToken == "" {
		return "", errors.New("Missing header key: Authorization")
	}

	token := strings.Split(bearerToken, " ")

	if len(token) == 0 {
		return "", errors.New("Authorization key does not exist")
	}
	if len(token) == 1 {
		return "", errors.New("Please provide token")
	}

	return token[1], nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func GetAPIKey(header http.Header) (string, error) {
	if len(header.Values("Authorization")) <= 0 {
		err := errors.New("Authorization header does not exist")
		return "", err
	}

	bearerToken := http.Header.Get(header, "Authorization")
	if bearerToken == "" {
		return "", errors.New("Missing header key: Authorization")
	}

	token := strings.Split(bearerToken, " ")

	if len(token) == 0 {
		return "", errors.New("Authorization key does not exist")
	}
	if len(token) == 1 {
		return "", errors.New("Please provide API key")
	}

	return token[1], nil
}
