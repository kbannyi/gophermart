package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AuthUser struct {
	UserID string
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const HeaderName = "Authorization"
const CookieName = "Authorization"
const tokenExp = time.Hour * 24
const secretKey = "supersecretkey"

type contextkey string

const contextKey contextkey = "UserContext"

var ErrContextAuthNotfound = errors.New("couldn't extract auth from context")

func FromContext(ctx context.Context) (AuthUser, error) {
	u, ok := ctx.Value(contextKey).(AuthUser)
	if !ok {
		return AuthUser{}, ErrContextAuthNotfound
	}

	return u, nil
}

func ToContext(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, contextKey, user)
}

func BuildJWTString(user AuthUser) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: user.UserID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ReadJWTString(tokenString string) (AuthUser, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
	if err != nil {
		return AuthUser{}, err
	}

	return AuthUser{UserID: claims.UserID}, nil
}
