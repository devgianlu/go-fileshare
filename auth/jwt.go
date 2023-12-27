package auth

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type customClaims struct {
	jwt.RegisteredClaims

	// TODO: perhaps support anonymous user JWT without a subject?
}

type jsonWebTokenProvider struct {
	secret []byte
	parser *jwt.Parser
}

func (p *jsonWebTokenProvider) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
	}

	return p.secret, nil
}

func (p *jsonWebTokenProvider) GetUser(tokenString string) (string, error) {
	token, err := p.parser.ParseWithClaims(tokenString, &customClaims{}, p.keyFunc)
	if err != nil {
		return "", fileshare.NewError("", fileshare.ErrAuthMalformed, err)
	}

	if !token.Valid {
		return "", fileshare.NewError("", fileshare.ErrAuthInvalid, err)
	}

	claims := token.Claims.(*customClaims)
	return claims.Subject, nil
}

func (p *jsonWebTokenProvider) GetToken(nickname string) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   nickname,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	})

	tokenString, err := token.SignedString(p.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func NewJsonWebTokenProvider(secret []byte) (fileshare.TokenProvider, error) {
	if len(secret) == 0 {
		return nil, fmt.Errorf("missing secret")
	}

	p := jsonWebTokenProvider{}
	p.secret = secret
	p.parser = jwt.NewParser(
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	return &p, nil
}
