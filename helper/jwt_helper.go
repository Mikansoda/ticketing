package helper

import (
	"errors"
	"time"

	"ticketing/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(uid, email, role string) (string, time.Time, error) {
	exp := time.Now().Add(time.Duration(config.C.AccessTTLMin) * time.Minute)
	claims := JWTClaims{
		UserID: uid,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(config.C.JWTAccessKey))
	return s, exp, err
}

func GenerateRefreshToken(uid, email, role string) (string, time.Time, error) {
	exp := time.Now().Add(time.Duration(config.C.RefreshTTLDays) * 24 * time.Hour)
	claims := JWTClaims{
		UserID: uid,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(config.C.JWTRefreshKey))
	return s, exp, err
}

func ParseAccess(token string) (*JWTClaims, error) {
	t, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.C.JWTAccessKey), nil
	})
	if err != nil {
		return nil, err
	}
	if cl, ok := t.Claims.(*JWTClaims); ok && t.Valid {
		return cl, nil
	}
	return nil, errors.New("invalid")
}

func ParseRefresh(token string) (*JWTClaims, string, string, error) {
	t, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.C.JWTRefreshKey), nil
	})
	if err != nil {
		return nil, "", "", err
	}
	if cl, ok := t.Claims.(*JWTClaims); ok && t.Valid {
		return cl, cl.Email, cl.Role, nil
	}
	return nil, "", "", errors.New("invalid")
}
