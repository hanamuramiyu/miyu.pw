package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/auth"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/config"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/content"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/db"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/httpapi"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/publish"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/webui"
)

const (
	authRateLimit  = 10
	authRateWindow = time.Minute
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("miyu-admin: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	openCtx, cancelOpen := context.WithTimeout(ctx, 10*time.Second)
	database, err := db.Open(openCtx, cfg.DatabaseURL)
	cancelOpen()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer database.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	err = database.Ping(pingCtx)
	cancelPing()
	if err != nil {
		return fmt.Errorf("database ping: %w", err)
	}

	sessions := auth.NewSessions(auth.NewPostgresStore(database.Pool), auth.SessionTTL)
	oauth := auth.NewOAuth(
		cfg.GitHubClientID,
		cfg.GitHubClientSecret,
		cfg.CallbackURL(),
		cfg.GitHubAllowedLogins,
		auth.DefaultEndpoints(),
		nil,
	)
	handlers := auth.NewHandlers(auth.HandlerOptions{
		OAuth:        oauth,
		Sessions:     sessions,
		States:       auth.NewStateStore(auth.StateTTL),
		SecureCookie: cfg.CookieSecure,
		SessionTTL:   auth.SessionTTL,
		StateTTL:     auth.StateTTL,
		Logger:       logger,
	})
	limiter := auth.NewRateLimiter(authRateLimit, authRateWindow)
	works := content.NewRepository(database.Pool)
	publisher := publish.NewService(works, cfg.PublishDir)

	server := &http.Server{
		Addr: cfg.Address(),
		Handler: httpapi.New(httpapi.Options{
			Sessions:  sessions,
			Auth:      handlers,
			Limiter:   limiter,
			Health:    database,
			Works:     works,
			Publisher: publisher,
			UI:        webui.Handler(),
			Logger:    logger,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Printf("miyu-admin listening on http://%s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server: %w", err)
		}
	case <-ctx.Done():
		stop()

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	}

	return nil
}
