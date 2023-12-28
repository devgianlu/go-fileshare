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
	"path/filepath"
)

type Config struct {
	Port     int    `yaml:"port"`
	Secret   string `yaml:"secret"`
	Path     string `yaml:"path"`
	LogLevel string `yaml:"log_level"`

	AnonymousAccess bool `yaml:"anonymous_access"`

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
	checkAcl := func(list []fileshare.PathACL) error {
		for i, item := range list {
			// check path is the shortest version
			if item.Path != filepath.Clean(item.Path) {
				return fmt.Errorf("path is not clean: %s", item.Path)
			}

			// check no duplicates
			for j, item_ := range list {
				if i == j {
					continue
				} else if item.Path == item_.Path {
					return fmt.Errorf("duplicate path %s", item.Path)
				}
			}

			// check it makes sense
			if !item.Read && item.Write {
				return fmt.Errorf("invalid read denied write allowed for %s", item.Path)
			}
		}

		return nil
	}

	// check default ACL
	if err := checkAcl(cfg.DefaultACL); err != nil {
		log.WithField("module", "config").WithError(err).Fatal("invalid default ACL")
	}

	var anonymousOk bool
	for i, user := range cfg.Users {
		// check no duplicates
		for j, user_ := range cfg.Users {
			if i == j {
				continue
			} else if user.Nickname == user_.Nickname {
				log.WithField("module", "config").Fatalf("duplicate user %s", user.Nickname)
			}
		}

		// check admin does not have ACL
		if user.Admin && len(user.ACL) > 0 {
			log.WithField("module", "config").Warnf("redundant ACL for admin user %s", user.Nickname)
		}

		// check user ACL
		if err := checkAcl(user.ACL); err != nil {
			log.WithField("module", "config").WithError(err).Fatalf("invalid ACL for %s", user.Nickname)
		}

		// check if user is anonymous
		if user.Anonymous() {
			if cfg.AnonymousAccess {
				anonymousOk = true
			} else {
				log.WithField("module", "config").Warn("redundant anonymous user")
			}
		}
	}

	// check there is an "anonymous" user if anonymous access is enabled
	if cfg.AnonymousAccess && !anonymousOk {
		log.WithField("module", "config").Fatal("missing anonymous user")
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
				log.WithError(err).WithField("module", "auth").Fatal("failed unmarshalling password auth provider config")
			}

			provider, err = auth.NewPasswordAuthProvider(providerCfg)
		case auth.AuthProviderTypeGithub:
			var providerCfg fileshare.AuthGithub
			if err := val.Decode(&providerCfg); err != nil {
				log.WithError(err).WithField("module", "auth").Fatal("failed unmarshalling github auth provider config")
			}

			provider, err = auth.NewGithubAuthProvider(providerCfg)
		default:
			err = fmt.Errorf("unknown provider %s", key)
		}

		if err != nil {
			log.WithError(err).WithField("module", "auth").Fatalf("failed creating provider %s", key)
		}

		s.Auth[key] = provider
	}

	// setup storage with ACL
	storage.NewACLStorageProvider(storage.NewLocalStorageProvider(cfg.Path), cfg.DefaultACL)

	// setup HTTP server
	s.HTTP = http.NewHTTPServer(cfg.Port, cfg.AnonymousAccess, s.Storage, s.Auth, s.Users, s.Tokens)

	// listen
	if err := s.HTTP.ListenForever(); err != nil {
		log.WithError(err).Fatalf("failed listening")
	}
}
