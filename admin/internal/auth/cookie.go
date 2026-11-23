package auth

import (
	"net/http"
	"time"
)

const (
	SessionCookieName = "miyu_admin_session"
	StateCookieName   = "miyu_admin_oauth_state"
	CSRFTokenHeader   = "X-CSRF-Token"
)

const StateTTL = 10 * time.Minute

func SessionCookie(token string, secure bool, ttl time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(ttl / time.Second),
	}
}

func ClearSessionCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
	}
}

func StateCookie(state string, secure bool, ttl time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     StateCookieName,
		Value:    state,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl / time.Second),
	}
}

func ClearStateCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     StateCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
	}
}
