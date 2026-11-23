package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/content"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/httpx"
)

const maxWorkBody = 1 << 20

type WorkStore interface {
	List(ctx context.Context, includeDrafts bool) ([]content.WorkSummary, error)
	Get(ctx context.Context, id string) (*content.Work, error)
	Create(ctx context.Context, w *content.Work) error
	Update(ctx context.Context, w *content.Work) error
	Delete(ctx context.Context, id string) error
	Publish(ctx context.Context, id string, published bool) error
}

type worksAPI struct {
	store  WorkStore
	logger *log.Logger
}

func (a *worksAPI) list(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
		return
	}

	includeDrafts := false
	switch strings.ToLower(r.URL.Query().Get("drafts")) {
	case "1", "true":
		includeDrafts = true
	}

	works, err := a.store.List(r.Context(), includeDrafts)
	if err != nil {
		a.writeError(w, err)
		return
	}
	if works == nil {
		works = []content.WorkSummary{}
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"works": works})
}

func (a *worksAPI) get(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
		return
	}

	work, err := a.store.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, work)
}

func (a *worksAPI) create(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
		return
	}

	var work content.Work
	if !decodeJSON(w, r, &work) {
		return
	}
	work.ID = ""

	if err := a.store.Create(r.Context(), &work); err != nil {
		a.writeError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, &work)
}

func (a *worksAPI) update(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
		return
	}

	var work content.Work
	if !decodeJSON(w, r, &work) {
		return
	}
	work.ID = r.PathValue("id")

	if !validateImagePaths(w, work.ID, work.Images) {
		return
	}

	if err := a.store.Update(r.Context(), &work); err != nil {
		a.writeError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, &work)
}

func validateImagePaths(w http.ResponseWriter, workID string, images []content.Image) bool {
	for i, image := range images {
		err := content.ValidateImagePath(workID, image.Path)
		if err == nil {
			continue
		}

		var validation *content.ValidationError
		message := "is invalid"
		if errors.As(err, &validation) {
			message = validation.Message
		}
		httpx.JSON(w, http.StatusBadRequest, map[string]string{
			"error": message,
			"field": fmt.Sprintf("images[%d].path", i),
		})
		return false
	}
	return true
}

func (a *worksAPI) delete(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
		return
	}

	if err := a.store.Delete(r.Context(), r.PathValue("id")); err != nil {
		a.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *worksAPI) publish(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
		return
	}

	var request struct {
		Published *bool `json:"published"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.Published == nil {
		httpx.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "is required",
			"field": "published",
		})
		return
	}

	id := r.PathValue("id")
	if err := a.store.Publish(r.Context(), id, *request.Published); err != nil {
		a.writeError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"id":        id,
		"published": *request.Published,
	})
}

func (a *worksAPI) ready(w http.ResponseWriter) bool {
	if a.store == nil {
		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return false
	}
	return true
}

func (a *worksAPI) writeError(w http.ResponseWriter, err error) {
	writeStoreError(w, a.logger, err)
}

func writeStoreError(w http.ResponseWriter, logger *log.Logger, err error) {
	var validation *content.ValidationError
	switch {
	case errors.Is(err, content.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, "not found")
	case errors.Is(err, content.ErrValidation) && errors.As(err, &validation):
		httpx.JSON(w, http.StatusBadRequest, map[string]string{
			"error": validation.Message,
			"field": validation.Field,
		})
	case errors.Is(err, content.ErrValidation):
		httpx.Error(w, http.StatusBadRequest, "invalid work")
	default:
		if logger != nil {
			logger.Printf("works api: %v", err)
		}
		httpx.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxWorkBody)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return false
	}
	return true
}
