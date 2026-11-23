package httpapi

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/auth"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/httpx"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/publish"
)

const readinessTimeout = 2 * time.Second

type Pinger interface {
	Ping(ctx context.Context) error
}

type Publisher interface {
	Request(ctx context.Context, login string) error
	Status(ctx context.Context) (publish.Status, error)
}

type Options struct {
	Sessions  *auth.Sessions
	Auth      *auth.Handlers
	Limiter   *auth.RateLimiter
	Health    Pinger
	Works     WorkStore
	Publisher Publisher
	UI        http.Handler
	Logger    *log.Logger
}

func New(opts Options) http.Handler {
	logger := opts.Logger
	if logger == nil {
		logger = log.Default()
	}

	middleware := &auth.Middleware{Sessions: opts.Sessions}
	works := &worksAPI{store: opts.Works, logger: logger}
	publisher := &publishAPI{publisher: opts.Publisher, logger: logger}

	mux := http.NewServeMux()
	mux.Handle("/healthz", methodSet{
		http.MethodGet: http.HandlerFunc(handleHealthz),
	})
	mux.Handle("/readyz", methodSet{
		http.MethodGet: handleReadyz(opts.Health, logger),
	})

	mux.Handle("/auth/github", methodSet{
		http.MethodGet: limit(opts.Limiter, http.HandlerFunc(opts.Auth.GitHub)),
	})
	mux.Handle("/auth/callback", methodSet{
		http.MethodGet: limit(opts.Limiter, http.HandlerFunc(opts.Auth.Callback)),
	})
	mux.Handle("/auth/logout", methodSet{
		http.MethodPost: limit(opts.Limiter,
			middleware.RequireSession(auth.CSRF(http.HandlerFunc(opts.Auth.Logout))),
		),
	})

	mux.Handle("/api/me", methodSet{
		http.MethodGet: middleware.RequireSession(http.HandlerFunc(handleMe)),
	})

	mux.Handle("/api/works", methodSet{
		http.MethodGet:  middleware.RequireSession(http.HandlerFunc(works.list)),
		http.MethodPost: middleware.RequireSession(auth.CSRF(http.HandlerFunc(works.create))),
	})
	mux.Handle("/api/works/{id}", methodSet{
		http.MethodGet:    middleware.RequireSession(http.HandlerFunc(works.get)),
		http.MethodPut:    middleware.RequireSession(auth.CSRF(http.HandlerFunc(works.update))),
		http.MethodDelete: middleware.RequireSession(auth.CSRF(http.HandlerFunc(works.delete))),
	})
	mux.Handle("/api/works/{id}/publish", methodSet{
		http.MethodPost: middleware.RequireSession(auth.CSRF(http.HandlerFunc(works.publish))),
	})
	mux.Handle("/api/publish", methodSet{
		http.MethodPost: middleware.RequireSession(auth.CSRF(http.HandlerFunc(publisher.request))),
	})
	mux.Handle("/api/publish/status", methodSet{
		http.MethodGet: middleware.RequireSession(http.HandlerFunc(publisher.status)),
	})

	mux.HandleFunc("/api", handleNotFound)
	mux.HandleFunc("/api/", handleNotFound)
	mux.HandleFunc("/auth", handleNotFound)
	mux.HandleFunc("/auth/", handleNotFound)

	if opts.UI != nil {
		mux.Handle("/", opts.UI)
	} else {
		mux.HandleFunc("/", handleNotFound)
	}

	return recoverPanic(logger, securityHeaders(mux))
}

type methodSet map[string]http.Handler

func (m methodSet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := m[r.Method]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	methods := make([]string, 0, len(m))
	for method := range m {
		methods = append(methods, method)
	}
	sort.Strings(methods)

	w.Header().Set("Allow", strings.Join(methods, ", "))
	httpx.Error(w, http.StatusMethodNotAllowed, "method not allowed")
}

func recoverPanic(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Printf("panic serving %s %s: %v\n%s", r.Method, r.URL.Path, recovered, debug.Stack())
				httpx.Error(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func handleReadyz(health Pinger, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		if health == nil {
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]bool{"ok": false})
			return
		}
		if err := health.Ping(ctx); err != nil {
			logger.Printf("readiness check failed: %v", err)
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]bool{"ok": false})
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{
		"login":     session.Login,
		"csrfToken": session.CSRFToken,
	})
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	httpx.Error(w, http.StatusNotFound, "not found")
}

func limit(limiter *auth.RateLimiter, next http.Handler) http.Handler {
	if limiter == nil {
		return next
	}
	return limiter.Middleware(next)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("Cache-Control", "no-store")
		header.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		next.ServeHTTP(w, r)
	})
}
