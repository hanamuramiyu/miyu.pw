package auth

import (
	"crypto/subtle"
	"log"
	"net/http"
	"time"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/httpx"
)

type HandlerOptions struct {
	OAuth        *OAuth
	Sessions     *Sessions
	States       *StateStore
	SecureCookie bool
	SessionTTL   time.Duration
	StateTTL     time.Duration
	Logger       *log.Logger
}

type Handlers struct {
	oauth        *OAuth
	sessions     *Sessions
	states       *StateStore
	secureCookie bool
	sessionTTL   time.Duration
	stateTTL     time.Duration
	logger       *log.Logger
}

func NewHandlers(opts HandlerOptions) *Handlers {
	sessionTTL := opts.SessionTTL
	if sessionTTL <= 0 {
		sessionTTL = SessionTTL
	}
	stateTTL := opts.StateTTL
	if stateTTL <= 0 {
		stateTTL = StateTTL
	}
	logger := opts.Logger
	if logger == nil {
		logger = log.Default()
	}

	return &Handlers{
		oauth:        opts.OAuth,
		sessions:     opts.Sessions,
		states:       opts.States,
		secureCookie: opts.SecureCookie,
		sessionTTL:   sessionTTL,
		stateTTL:     stateTTL,
		logger:       logger,
	}
}

func (h *Handlers) GitHub(w http.ResponseWriter, r *http.Request) {
	state, err := h.states.Issue()
	if err != nil {
		h.logger.Printf("issue oauth state: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	target, err := h.oauth.AuthorizationURL(state)
	if err != nil {
		h.logger.Printf("build authorization url: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, StateCookie(state, h.secureCookie, h.stateTTL))
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *Handlers) Callback(w http.ResponseWriter, r *http.Request) {
	defer func() {
		http.SetCookie(w, ClearStateCookie(h.secureCookie))
	}()

	query := r.URL.Query()
	queryState := query.Get("state")
	cookie, err := r.Cookie(StateCookieName)
	if err != nil || queryState == "" || cookie.Value == "" ||
		subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(queryState)) != 1 {
		httpx.Error(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	if !h.states.Consume(queryState) {
		httpx.Error(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	if query.Get("error") != "" {
		h.logger.Printf("github authorization denied: %s", query.Get("error"))
		httpx.Error(w, http.StatusBadRequest, "authorization denied")
		return
	}

	code := query.Get("code")
	if code == "" {
		httpx.Error(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	accessToken, err := h.oauth.Exchange(r.Context(), code)
	if err != nil {
		h.logger.Printf("github token exchange failed: %v", err)
		httpx.Error(w, http.StatusBadGateway, "github authentication failed")
		return
	}

	login, err := h.oauth.UserLogin(r.Context(), accessToken)
	if err != nil {
		h.logger.Printf("github user request failed: %v", err)
		httpx.Error(w, http.StatusBadGateway, "github authentication failed")
		return
	}

	if !h.oauth.Allowed(login) {
		h.logger.Printf("rejected github login %q", login)
		httpx.Error(w, http.StatusForbidden, "forbidden")
		return
	}

	token, _, err := h.sessions.Create(r.Context(), login)
	if err != nil {
		h.logger.Printf("create session: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, SessionCookie(token, h.secureCookie, h.sessionTTL))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil && cookie.Value != "" {
		if err := h.sessions.Delete(r.Context(), cookie.Value); err != nil {
			h.logger.Printf("delete session: %v", err)
			httpx.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	http.SetCookie(w, ClearSessionCookie(h.secureCookie))
	w.WriteHeader(http.StatusNoContent)
}
