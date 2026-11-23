package auth

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAllowedIsCaseInsensitive(t *testing.T) {
	oauth := NewOAuth("id", "secret", "http://admin.example/auth/callback", []string{"OctoCat"}, Endpoints{}, nil)

	if !oauth.Allowed("octocat") {
		t.Error("Allowed(octocat) = false, want true")
	}
	if !oauth.Allowed("OCTOCAT") {
		t.Error("Allowed(OCTOCAT) = false, want true")
	}
	if oauth.Allowed("someone-else") {
		t.Error("Allowed(someone-else) = true, want false")
	}
}

func TestAllowedMatchesAnyLogin(t *testing.T) {
	oauth := NewOAuth("id", "secret", "http://admin.example/auth/callback",
		[]string{"first-user", "second-user", "third-user"}, Endpoints{}, nil)

	for _, login := range []string{"first-user", "Second-User", "THIRD-USER"} {
		if !oauth.Allowed(login) {
			t.Errorf("Allowed(%q) = false, want true", login)
		}
	}
	if oauth.Allowed("fourth-user") {
		t.Error("Allowed(fourth-user) = true, want false")
	}
}

func TestAllowedRejectsEmptyAllowList(t *testing.T) {
	oauth := NewOAuth("id", "secret", "http://admin.example/auth/callback", nil, Endpoints{}, nil)

	if oauth.Allowed("octocat") {
		t.Error("Allowed(octocat) = true with an empty allow list")
	}
	if oauth.Allowed("") {
		t.Error("Allowed(empty) = true with an empty allow list")
	}
}

func TestAuthorizationURL(t *testing.T) {
	oauth := NewOAuth("client-id", "client-secret", "http://admin.example/auth/callback", []string{"octocat"}, Endpoints{}, nil)

	raw, err := oauth.AuthorizationURL("state-value")
	if err != nil {
		t.Fatalf("AuthorizationURL: %v", err)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse authorization url: %v", err)
	}

	query := parsed.Query()
	if got := query.Get("client_id"); got != "client-id" {
		t.Errorf("client_id = %q, want client-id", got)
	}
	if got := query.Get("scope"); got != "read:user" {
		t.Errorf("scope = %q, want read:user", got)
	}
	if got := query.Get("state"); got != "state-value" {
		t.Errorf("state = %q, want state-value", got)
	}
	if got := query.Get("redirect_uri"); got != "http://admin.example/auth/callback" {
		t.Errorf("redirect_uri = %q", got)
	}
}

func newGitHubStub(t *testing.T, login string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login/oauth/access_token":
			if err := r.ParseForm(); err != nil {
				http.Error(w, "bad form", http.StatusBadRequest)
				return
			}
			if r.PostForm.Get("client_id") != "client-id" ||
				r.PostForm.Get("client_secret") != "client-secret" ||
				r.PostForm.Get("code") != "test-code" {
				http.Error(w, "bad credentials", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "test-access-token"})
		case "/user":
			if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
				http.Error(w, "bad token", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"login": login})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func newTestHandlers(t *testing.T, allowedLogin string) (*Handlers, *fakeStore, *httptest.Server) {
	t.Helper()

	github := newGitHubStub(t, "octocat")
	endpoints := Endpoints{
		AuthorizeURL: github.URL + "/login/oauth/authorize",
		TokenURL:     github.URL + "/login/oauth/access_token",
		UserURL:      github.URL + "/user",
	}
	oauth := NewOAuth("client-id", "client-secret", "http://admin.example/auth/callback", []string{allowedLogin}, endpoints, github.Client())

	store := newFakeStore()
	sessions := NewSessions(store, SessionTTL)
	handlers := NewHandlers(HandlerOptions{
		OAuth:        oauth,
		Sessions:     sessions,
		States:       NewStateStore(StateTTL),
		SecureCookie: false,
		Logger:       log.New(io.Discard, "", 0),
	})
	return handlers, store, github
}

func startLogin(t *testing.T, handlers *Handlers) (state string, stateCookie *http.Cookie) {
	t.Helper()

	recorder := httptest.NewRecorder()
	handlers.GitHub(recorder, httptest.NewRequest(http.MethodGet, "/auth/github", nil))

	if recorder.Code != http.StatusFound {
		t.Fatalf("GitHub status = %d, want 302", recorder.Code)
	}

	location, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect: %v", err)
	}
	state = location.Query().Get("state")
	if state == "" {
		t.Fatal("redirect has no state")
	}

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == StateCookieName {
			stateCookie = cookie
			break
		}
	}
	if stateCookie == nil || stateCookie.Value != state {
		t.Fatal("state cookie is missing or does not match the redirect")
	}
	return state, stateCookie
}

func callbackRequest(state string, stateCookie *http.Cookie) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?code=test-code&state="+url.QueryEscape(state), nil)
	if stateCookie != nil {
		request.AddCookie(stateCookie)
	}
	return request
}

func TestCallbackCreatesSessionOnce(t *testing.T) {
	handlers, store, _ := newTestHandlers(t, "octocat")

	state, stateCookie := startLogin(t, handlers)

	recorder := httptest.NewRecorder()
	handlers.Callback(recorder, callbackRequest(state, stateCookie))

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("Callback status = %d, want 303, body = %s", recorder.Code, recorder.Body.String())
	}

	var sessionCookie *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("session cookie was not set")
	}
	if !sessionCookie.HttpOnly {
		t.Error("session cookie is not HttpOnly")
	}
	if sessionCookie.SameSite != http.SameSiteStrictMode {
		t.Error("session cookie is not SameSite=Strict")
	}
	if store.count() != 1 {
		t.Fatalf("session records = %d, want 1", store.count())
	}

	replay := httptest.NewRecorder()
	handlers.Callback(replay, callbackRequest(state, stateCookie))
	if replay.Code != http.StatusBadRequest {
		t.Fatalf("replayed callback status = %d, want 400", replay.Code)
	}
	if store.count() != 1 {
		t.Fatalf("replay created a session, records = %d", store.count())
	}
}

func TestCallbackRequiresStateCookie(t *testing.T) {
	handlers, store, _ := newTestHandlers(t, "octocat")

	state, _ := startLogin(t, handlers)

	recorder := httptest.NewRecorder()
	handlers.Callback(recorder, callbackRequest(state, nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	if store.count() != 0 {
		t.Errorf("session created without state cookie")
	}
}

func TestCallbackRejectsUnknownState(t *testing.T) {
	handlers, store, _ := newTestHandlers(t, "octocat")

	recorder := httptest.NewRecorder()
	handlers.Callback(recorder, callbackRequest("unknown-state", &http.Cookie{
		Name:  StateCookieName,
		Value: "unknown-state",
	}))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	if store.count() != 0 {
		t.Errorf("session created for an unknown state")
	}
}

func TestCallbackRejectsLoginOutsideAllowList(t *testing.T) {
	handlers, store, _ := newTestHandlers(t, "someone-else")

	state, stateCookie := startLogin(t, handlers)

	recorder := httptest.NewRecorder()
	handlers.Callback(recorder, callbackRequest(state, stateCookie))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
	if store.count() != 0 {
		t.Errorf("session created for a disallowed login")
	}
}
