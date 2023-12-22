package auth

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/golang-jwt/jwt/v5"
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
	return &fileshare.User{Permissions: claims.Permissions}, nil
}

func NewJWTAuthProvider(secret []byte) fileshare.AuthProvider {
	p := jwtAuthProvider{}
	p.secret = secret
	p.parser = jwt.NewParser()
	return &p
}
