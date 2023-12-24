package main

import (
	"github.com/devgianlu/go-fileshare"
	"github.com/devgianlu/go-fileshare/auth"
	"github.com/devgianlu/go-fileshare/http"
	"github.com/devgianlu/go-fileshare/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Port     int
	Secret   string
	Path     string
	LogLevel string

	DefaultACL []fileshare.PathACL `mapstructure:"default_acl"`

	Users []fileshare.User
}

func loadConfig() (*Config, error) {
	viper.SetDefault("logLevel", "info")

	// load config from local "server.yml" file
	viper.AddConfigPath(".")
	viper.SetConfigName("server")
	viper.SetConfigType("yml")

	// try to load from env
	viper.SetEnvPrefix("FILESHARE_")
	viper.AutomaticEnv()

	// load from file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) {
	for _, user := range cfg.Users {
		// check admin does not have ACL
		if user.Admin && len(user.ACL) > 0 {
			log.WithField("module", "config").Warnf("redundant ACL for admin user %s", user.Nickname)
		}
	}
}

type Server struct {
	Storage fileshare.AuthenticatedStorageProvider
	Users   fileshare.UsersProvider
	Auth    fileshare.AuthProvider
	HTTP    fileshare.HttpServer
}

func main() {
	// load config
	cfg, err := loadConfig()
	if err != nil {
		log.WithError(err).Fatal("cannot load config")
	}

	// parse and set log level
	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.WithError(err).Fatalf("invalid log level")
	}
	log.SetLevel(logLevel)

	// validate config and log errors/warnings
	validateConfig(cfg)

	s := Server{}

	// setup users provider
	s.Users = auth.NewConfigUsersProvider(cfg.Users)

	// setup authentication with JWT
	if s.Auth, err = auth.NewJWTAuthProvider([]byte(cfg.Secret)); err != nil {
		log.WithError(err).WithField("module", "auth").Fatalf("failed creating JWT auth provider")
	}

	// setup storage with ACL
	if s.Storage, err = storage.NewACLStorageProvider(storage.NewLocalStorageProvider(cfg.Path), cfg.DefaultACL); err != nil {
		log.WithError(err).WithField("module", "storage").Fatalf("failed creating access controlled storage provider")
	}

	// setup HTTP server
	s.HTTP = http.NewHTTPServer(cfg.Port, s.Storage, s.Users, s.Auth)

	// listen
	if err := s.HTTP.ListenForever(); err != nil {
		log.WithError(err).Fatalf("failed listening")
	}
}
