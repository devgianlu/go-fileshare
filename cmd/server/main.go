package main

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/devgianlu/go-fileshare/auth"
	"github.com/devgianlu/go-fileshare/http"
	"github.com/devgianlu/go-fileshare/storage"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Port     int    `yaml:"port"`
	Secret   string `yaml:"secret"`
	Path     string `yaml:"path"`
	LogLevel string `yaml:"log_level"`

	DefaultACL []fileshare.PathACL `yaml:"default_acl"`

	Users []fileshare.User     `yaml:"users"`
	Auths map[string]yaml.Node `yaml:"auths"`
}

func loadConfig() (*Config, error) {
	f, err := os.OpenFile("server.yml", os.O_RDONLY, 0000)
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	dec := yaml.NewDecoder(f)

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
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
	Auth    map[string]fileshare.AuthProvider
	Users   fileshare.UsersProvider
	Tokens  fileshare.TokenProvider
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

	// setup tokens with JWT
	if s.Tokens, err = auth.NewJsonWebTokenProvider([]byte(cfg.Secret)); err != nil {
		log.WithError(err).WithField("module", "auth").Fatalf("failed creating JWT provider")
	}

	// setup authentication providers
	s.Auth = map[string]fileshare.AuthProvider{}
	for key, val := range cfg.Auths {
		var provider fileshare.AuthProvider
		switch key {
		case auth.AuthProviderTypePassword:
			var providerCfg fileshare.AuthPassword
			if err := val.Decode(&providerCfg); err != nil {
				log.WithError(err).WithField("module", "auth").Fatal("failed unmarshalling password provider config")
			}

			provider, err = auth.NewPasswordAuthProvider(providerCfg)
		default:
			err = fmt.Errorf("unknown provider %s", key)
		}

		if err != nil {
			log.WithError(err).WithField("module", "auth").Fatalf("failed creating provider %s", key)
		}

		s.Auth[key] = provider
	}

	// setup storage with ACL
	if s.Storage, err = storage.NewACLStorageProvider(storage.NewLocalStorageProvider(cfg.Path), cfg.DefaultACL); err != nil {
		log.WithError(err).WithField("module", "storage").Fatalf("failed creating access controlled storage provider")
	}

	// setup HTTP server
	s.HTTP = http.NewHTTPServer(cfg.Port, s.Storage, s.Auth, s.Users, s.Tokens)

	// listen
	if err := s.HTTP.ListenForever(); err != nil {
		log.WithError(err).Fatalf("failed listening")
	}
}
