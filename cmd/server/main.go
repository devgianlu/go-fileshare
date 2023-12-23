package main

import (
	"github.com/devgianlu/go-fileshare"
	"github.com/devgianlu/go-fileshare/auth"
	"github.com/devgianlu/go-fileshare/http"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Port     int
	Secret   string
	LogLevel string
}

func loadConfig() (*Config, error) {
	viper.SetDefault("logLevel", "info")

	// load config from local "server.yml" file
	viper.AddConfigPath(".")
	viper.SetConfigName("server")
	viper.SetConfigType("yml")

	// try to load from env
	viper.AutomaticEnv()

	// load from file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

type Server struct {
	Auth fileshare.AuthProvider
	HTTP fileshare.HttpServer
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.WithError(err).Fatal("cannot load config")
	}

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.WithError(err).Fatalf("invalid log level")
	}

	log.SetLevel(logLevel)

	s := Server{}
	s.Auth = auth.NewJWTAuthProvider([]byte(cfg.Secret))
	s.HTTP = http.NewHTTPServer(cfg.Port, s.Auth)

	if err := s.HTTP.ListenForever(); err != nil {
		log.WithError(err).Fatalf("failed listening")
	}
}
