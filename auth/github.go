package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"net/http"
)

const AuthProviderTypeGithub = "github"

type githubAuthProvider struct {
	cfg *oauth2.Config

	state string
}

func NewGithubAuthProvider(auth fileshare.AuthGithub) (fileshare.AuthProvider, error) {
	if len(auth.ClientID) == 0 || len(auth.ClientSecret) == 0 {
		return nil, fmt.Errorf("invalid config")
	}

	// I think this is good enough for us
	stateBytes := make([]byte, 16)
	_, _ = rand.Read(stateBytes)

	return &githubAuthProvider{
		state: hex.EncodeToString(stateBytes),
		cfg: &oauth2.Config{
			RedirectURL:  fmt.Sprintf("%s/login/github/callback", auth.CallbackBaseURL),
			ClientID:     auth.ClientID,
			ClientSecret: auth.ClientSecret,
			Scopes:       []string{}, // no scopes required, nickname is public info
			Endpoint:     github.Endpoint,
		},
	}, nil
}

func (p *githubAuthProvider) Callback() (string, error) {
	url := p.cfg.AuthCodeURL(p.state)
	return url, nil
}

func (p *githubAuthProvider) Authenticate(payload_ any) (string, error) {
	payload, ok := payload_.(fileshare.OAuth2ProviderPayload)
	if !ok {
		panic("invalid payload type")
	}

	if payload.State != p.state {
		return "", fmt.Errorf("invalid state")
	}

	token, err := p.cfg.Exchange(context.Background(), payload.Code)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}

	token.SetAuthHeader(req)

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github bad status code: %d", resp.StatusCode)
	}

	var body struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}

	return body.Login, nil
}
