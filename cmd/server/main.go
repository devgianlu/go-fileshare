package main

import (
	"github.com/devgianlu/go-fileshare"
	"github.com/devgianlu/go-fileshare/auth"
	"github.com/devgianlu/go-fileshare/http"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	Auth fileshare.AuthProvider
	HTTP fileshare.HttpServer
}

func main() {
	const PORT = 8080         // FIXME
	const SECRET = "test1234" // FIXME

	log.SetLevel(log.TraceLevel)

	s := Server{}
	s.Auth = auth.NewJWTAuthProvider([]byte(SECRET))
	s.HTTP = http.NewHTTPServer(PORT, s.Auth)

	if err := s.HTTP.ListenForever(); err != nil {
		log.WithError(err).Fatalf("failed listening")
	}
}
