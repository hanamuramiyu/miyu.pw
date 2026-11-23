package config

import (
	"reflect"
	"strings"
	"testing"
)

func fullEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL":         "postgres://user:pass@127.0.0.1:5432/miyu",
		"GITHUB_CLIENT_ID":     "client-id",
		"GITHUB_CLIENT_SECRET": "client-secret",
		"GITHUB_ALLOWED_LOGIN": "octocat",
		"PUBLIC_BASE_URL":      "https://admin.example",
	}
}

func envFunc(values map[string]string) func(string) string {
	return func(name string) string {
		return values[name]
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := load(envFunc(fullEnv()))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if !cfg.CookieSecure {
		t.Error("CookieSecure = false, want true by default")
	}
	if got, want := cfg.Address(), "127.0.0.1:8080"; got != want {
		t.Errorf("Address() = %q, want %q", got, want)
	}
	if got, want := cfg.CallbackURL(), "https://admin.example/auth/callback"; got != want {
		t.Errorf("CallbackURL() = %q, want %q", got, want)
	}
	if got, want := cfg.GitHubAllowedLogins, []string{"octocat"}; !reflect.DeepEqual(got, want) {
		t.Errorf("GitHubAllowedLogins = %v, want %v", got, want)
	}
	if cfg.PublishDir != defaultPublishDir {
		t.Errorf("PublishDir = %q, want %q", cfg.PublishDir, defaultPublishDir)
	}
}

func TestLoadParsesAllowedLogins(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "several",
			raw:  "user-one,user-two",
			want: []string{"user-one", "user-two"},
		},
		{
			name: "trims and lowercases",
			raw:  " OctoCat , , SECOND-user ,",
			want: []string{"octocat", "second-user"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := fullEnv()
			values["GITHUB_ALLOWED_LOGIN"] = test.raw

			cfg, err := load(envFunc(values))
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if !reflect.DeepEqual(cfg.GitHubAllowedLogins, test.want) {
				t.Errorf("GitHubAllowedLogins = %v, want %v", cfg.GitHubAllowedLogins, test.want)
			}
		})
	}
}

func TestLoadRejectsEmptyAllowedLogins(t *testing.T) {
	for _, raw := range []string{"", "   ", " , , "} {
		values := fullEnv()
		values["GITHUB_ALLOWED_LOGIN"] = raw

		_, err := load(envFunc(values))
		if err == nil {
			t.Fatalf("load succeeded for GITHUB_ALLOWED_LOGIN %q, want error", raw)
		}
		if !strings.Contains(err.Error(), "GITHUB_ALLOWED_LOGIN") {
			t.Errorf("error %q does not mention GITHUB_ALLOWED_LOGIN", err)
		}
	}
}

func TestLoadOverrides(t *testing.T) {
	values := fullEnv()
	values["PORT"] = "9090"
	values["COOKIE_SECURE"] = "false"
	values["PUBLIC_BASE_URL"] = "http://127.0.0.1:8080/"
	values["PUBLISH_DIR"] = "/tmp/miyu-test-state"

	cfg, err := load(envFunc(values))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.CookieSecure {
		t.Error("CookieSecure = true, want false")
	}
	if got, want := cfg.CallbackURL(), "http://127.0.0.1:8080/auth/callback"; got != want {
		t.Errorf("CallbackURL() = %q, want %q", got, want)
	}
	if cfg.PublishDir != "/tmp/miyu-test-state" {
		t.Errorf("PublishDir = %q, want /tmp/miyu-test-state", cfg.PublishDir)
	}
}

func TestLoadReportsAllMissingVariables(t *testing.T) {
	_, err := load(envFunc(map[string]string{}))
	if err == nil {
		t.Fatal("load succeeded, want error")
	}
	for _, name := range []string{
		"DATABASE_URL",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
		"GITHUB_ALLOWED_LOGIN",
		"PUBLIC_BASE_URL",
	} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not mention %s", err, name)
		}
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]string)
		want   string
	}{
		{
			name:   "base url without scheme",
			mutate: func(v map[string]string) { v["PUBLIC_BASE_URL"] = "admin.example" },
			want:   "PUBLIC_BASE_URL",
		},
		{
			name:   "base url with ftp scheme",
			mutate: func(v map[string]string) { v["PUBLIC_BASE_URL"] = "ftp://admin.example" },
			want:   "PUBLIC_BASE_URL",
		},
		{
			name:   "non numeric port",
			mutate: func(v map[string]string) { v["PORT"] = "http" },
			want:   "PORT",
		},
		{
			name:   "port out of range",
			mutate: func(v map[string]string) { v["PORT"] = "70000" },
			want:   "PORT",
		},
		{
			name:   "cookie secure not boolean",
			mutate: func(v map[string]string) { v["COOKIE_SECURE"] = "yes" },
			want:   "COOKIE_SECURE",
		},
		{
			name:   "relative publish dir",
			mutate: func(v map[string]string) { v["PUBLISH_DIR"] = "state" },
			want:   "PUBLISH_DIR",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := fullEnv()
			test.mutate(values)

			_, err := load(envFunc(values))
			if err == nil {
				t.Fatal("load succeeded, want error")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Errorf("error %q does not mention %s", err, test.want)
			}
		})
	}
}
