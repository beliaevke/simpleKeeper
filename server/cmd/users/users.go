package users

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/beliaevke/simpleKeeper/internal/db/postgres"
	"github.com/beliaevke/simpleKeeper/internal/repository/usersrepo"
	"github.com/beliaevke/simpleKeeper/internal/service"

	"github.com/golang-jwt/jwt/v4"
)

type database interface {
	CreateUser(ctx context.Context, u usersrepo.UserInfo) (int, error)
	GetUser(ctx context.Context, u usersrepo.UserInfo) (int, error)
	LoginUser(ctx context.Context, u usersrepo.UserInfo) (int, error)
	Timeout() time.Duration
}

const tokenExpiresAt = time.Minute * 30

func AuthenticateUser(userID int) (string, string, error) {
	// создаём случайный ключ
	key, err := service.GenerateRandom(16)
	if err != nil {
		return "", "", err
	}
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(tokenExpiresAt).Unix(),
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", "", err
	}

	return tokenString, base64.StdEncoding.EncodeToString(key), nil
}

func ValidateToken(tokenString string, encodedKey string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return false, errors.New("unexpected token signing method")
		}
		key, err := base64.StdEncoding.DecodeString(encodedKey)
		if err != nil {
			return false, err
		}
		return key, nil
	})

	if err != nil {
		return false, err
	}

	// Проверка, что токен действителен
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Проверка времени истечения
		if exp, ok := claims["exp"].(int64); ok {
			if time.Now().Unix() > exp {
				return false, errors.New("token has expired")
			}
		} else if expFloat, ok := claims["exp"].(float64); ok {
			exp := int64(expFloat)
			if time.Now().Unix() > exp {
				return false, errors.New("token has expired")
			}
		}
		return true, nil
	}

	return false, errors.New("invalid token")
}

func NewRepo(db *postgres.DB) database {
	return usersrepo.NewUser(db)
}
