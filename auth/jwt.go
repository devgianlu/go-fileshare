package auth

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type customClaims struct {
	jwt.RegisteredClaims
	Permissions []string `json:"permissions"`
}

type jwtAuthProvider struct {
	secret []byte
	parser *jwt.Parser
}

func (p *jwtAuthProvider) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
	}

	return p.secret, nil
}

func (p *jwtAuthProvider) GetUser(tokenString string) (*fileshare.User, error) {
	token, err := p.parser.ParseWithClaims(tokenString, &customClaims{}, p.keyFunc)
	if err != nil {
		return nil, fileshare.NewError("", fileshare.ErrAuthMalformed, err)
	}

	if !token.Valid {
		return nil, fileshare.NewError("", fileshare.ErrAuthInvalid, err)
	}

	claims := token.Claims.(*customClaims)
	return &fileshare.User{Nickname: claims.Subject, Permissions: claims.Permissions}, nil
}

func (p *jwtAuthProvider) GetToken(user *fileshare.User) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Nickname,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
		Permissions: user.Permissions,
	})

	tokenString, err := token.SignedString(p.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func NewJWTAuthProvider(secret []byte) fileshare.AuthProvider {
	p := jwtAuthProvider{}
	p.secret = secret
	p.parser = jwt.NewParser(
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	return &p
}
