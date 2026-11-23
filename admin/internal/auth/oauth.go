package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserURL      = "https://api.github.com/user"

	maxOAuthResponse = 1 << 20
)

type Endpoints struct {
	AuthorizeURL string
	TokenURL     string
	UserURL      string
}

func DefaultEndpoints() Endpoints {
	return Endpoints{
		AuthorizeURL: githubAuthorizeURL,
		TokenURL:     githubTokenURL,
		UserURL:      githubUserURL,
	}
}

type OAuth struct {
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	AllowedLogins []string
	Endpoints     Endpoints
	HTTPClient    *http.Client
}

func NewOAuth(clientID, clientSecret, redirectURL string, allowedLogins []string, endpoints Endpoints, client *http.Client) *OAuth {
	defaults := DefaultEndpoints()
	if endpoints.AuthorizeURL == "" {
		endpoints.AuthorizeURL = defaults.AuthorizeURL
	}
	if endpoints.TokenURL == "" {
		endpoints.TokenURL = defaults.TokenURL
	}
	if endpoints.UserURL == "" {
		endpoints.UserURL = defaults.UserURL
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &OAuth{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURL:   redirectURL,
		AllowedLogins: allowedLogins,
		Endpoints:     endpoints,
		HTTPClient:    client,
	}
}

func (o *OAuth) AuthorizationURL(state string) (string, error) {
	parsed, err := url.Parse(o.Endpoints.AuthorizeURL)
	if err != nil {
		return "", fmt.Errorf("parse authorize url: %w", err)
	}

	query := parsed.Query()
	query.Set("client_id", o.ClientID)
	query.Set("redirect_uri", o.RedirectURL)
	query.Set("scope", "read:user")
	query.Set("state", state)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func (o *OAuth) Exchange(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", o.ClientID)
	form.Set("client_secret", o.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", o.RedirectURL)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, o.Endpoints.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response, err := o.HTTPClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("request access token: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxOAuthResponse))
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned status %d", response.StatusCode)
	}

	var payload struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if payload.Error != "" {
		if payload.ErrorDescription != "" {
			return "", fmt.Errorf("token endpoint error: %s", payload.ErrorDescription)
		}
		return "", fmt.Errorf("token endpoint error: %s", payload.Error)
	}
	if payload.AccessToken == "" {
		return "", errors.New("token endpoint returned an empty token")
	}

	return payload.AccessToken, nil
}

func (o *OAuth) UserLogin(ctx context.Context, accessToken string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, o.Endpoints.UserURL, nil)
	if err != nil {
		return "", fmt.Errorf("build user request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "miyu-admin")

	response, err := o.HTTPClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("request github user: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxOAuthResponse))
	if err != nil {
		return "", fmt.Errorf("read user response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("user endpoint returned status %d", response.StatusCode)
	}

	var payload struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode user response: %w", err)
	}
	if payload.Login == "" {
		return "", errors.New("user endpoint returned an empty login")
	}

	return payload.Login, nil
}

func (o *OAuth) Allowed(login string) bool {
	candidate := strings.TrimSpace(login)
	if candidate == "" {
		return false
	}
	for _, allowed := range o.AllowedLogins {
		if strings.EqualFold(candidate, strings.TrimSpace(allowed)) {
			return true
		}
	}
	return false
}
