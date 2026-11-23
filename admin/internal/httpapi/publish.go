package httpapi

import (
	"errors"
	"log"
	"net/http"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/auth"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/httpx"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/publish"
)

type publishAPI struct {
	publisher Publisher
	logger    *log.Logger
}

func (a *publishAPI) request(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
		return
	}

	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := a.publisher.Request(r.Context(), session.Login); err != nil {
		var invalid *publish.InvalidWorkError
		if errors.As(err, &invalid) {
			httpx.JSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		a.logger.Printf("publish request: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.JSON(w, http.StatusAccepted, map[string]string{"state": "queued"})
}

func (a *publishAPI) status(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
		return
	}

	status, err := a.publisher.Status(r.Context())
	if err != nil {
		a.logger.Printf("publish status: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.JSON(w, http.StatusOK, status)
}

func (a *publishAPI) ready(w http.ResponseWriter) bool {
	if a.publisher == nil {
		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return false
	}
	return true
}
