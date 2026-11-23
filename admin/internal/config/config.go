package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultPort       = "8080"
	defaultPublishDir = "/var/lib/miyu-admin"
)

type Config struct {
	DatabaseURL         string
	GitHubClientID      string
	GitHubClientSecret  string
	GitHubAllowedLogins []string
	PublicBaseURL       string
	Port                string
	CookieSecure        bool
	PublishDir          string
}

func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	cfg := Config{
		DatabaseURL:         strings.TrimSpace(getenv("DATABASE_URL")),
		GitHubClientID:      strings.TrimSpace(getenv("GITHUB_CLIENT_ID")),
		GitHubClientSecret:  strings.TrimSpace(getenv("GITHUB_CLIENT_SECRET")),
		GitHubAllowedLogins: parseLogins(getenv("GITHUB_ALLOWED_LOGIN")),
		PublicBaseURL:       strings.TrimRight(strings.TrimSpace(getenv("PUBLIC_BASE_URL")), "/"),
		Port:                strings.TrimSpace(getenv("PORT")),
		CookieSecure:        true,
		PublishDir:          strings.TrimSpace(getenv("PUBLISH_DIR")),
	}

	required := []struct {
		name    string
		present bool
	}{
		{"DATABASE_URL", cfg.DatabaseURL != ""},
		{"GITHUB_CLIENT_ID", cfg.GitHubClientID != ""},
		{"GITHUB_CLIENT_SECRET", cfg.GitHubClientSecret != ""},
		{"GITHUB_ALLOWED_LOGIN", len(cfg.GitHubAllowedLogins) > 0},
		{"PUBLIC_BASE_URL", cfg.PublicBaseURL != ""},
	}
	var missing []string
	for _, field := range required {
		if !field.present {
			missing = append(missing, field.name)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if err := validateBaseURL(cfg.PublicBaseURL); err != nil {
		return Config{}, err
	}

	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if err := validatePort(cfg.Port); err != nil {
		return Config{}, err
	}

	if cfg.PublishDir == "" {
		cfg.PublishDir = defaultPublishDir
	}
	if !filepath.IsAbs(cfg.PublishDir) {
		return Config{}, fmt.Errorf("PUBLISH_DIR must be an absolute path")
	}

	if raw := strings.TrimSpace(getenv("COOKIE_SECURE")); raw != "" {
		secure, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("COOKIE_SECURE must be true or false")
		}
		cfg.CookieSecure = secure
	}

	return cfg, nil
}

func (c Config) Address() string {
	return "127.0.0.1:" + c.Port
}

func (c Config) CallbackURL() string {
	return c.PublicBaseURL + "/auth/callback"
}

func parseLogins(raw string) []string {
	var logins []string
	for _, part := range strings.Split(raw, ",") {
		login := strings.ToLower(strings.TrimSpace(part))
		if login != "" {
			logins = append(logins, login)
		}
	}
	return logins
}

func validateBaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("PUBLIC_BASE_URL is not a valid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("PUBLIC_BASE_URL must use http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("PUBLIC_BASE_URL must include a host")
	}
	return nil
}

func validatePort(raw string) error {
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PORT must be an integer between 1 and 65535")
	}
	return nil
}
