package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/auth"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/content"
)

const testWorkID = "00000000-0000-0000-0000-000000000001"

type fakeWorks struct {
	list          []content.WorkSummary
	work          *content.Work
	err           error
	includeDrafts bool
	created       *content.Work
	updated       *content.Work
	deletedID     string
	publishedID   string
	published     *bool
}

func (f *fakeWorks) List(_ context.Context, includeDrafts bool) ([]content.WorkSummary, error) {
	f.includeDrafts = includeDrafts
	return f.list, f.err
}

func (f *fakeWorks) Get(_ context.Context, _ string) (*content.Work, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.work, nil
}

func (f *fakeWorks) Create(_ context.Context, work *content.Work) error {
	f.created = work
	return f.err
}

func (f *fakeWorks) Update(_ context.Context, work *content.Work) error {
	f.updated = work
	return f.err
}

func (f *fakeWorks) Delete(_ context.Context, id string) error {
	f.deletedID = id
	return f.err
}

func (f *fakeWorks) Publish(_ context.Context, id string, published bool) error {
	f.publishedID = id
	f.published = &published
	return f.err
}

func authenticatedRequest(t *testing.T, sessions *auth.Sessions, method, path, body string, csrf bool) *http.Request {
	t.Helper()

	token, session, err := sessions.Create(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	if csrf {
		request.Header.Set(auth.CSRFTokenHeader, session.CSRFToken)
	}
	return request
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
}

func TestWorksRequireSession(t *testing.T) {
	handler, _ := newHandlerWithWorks(t, fakePinger{}, &fakeWorks{})

	requests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/works"},
		{http.MethodGet, "/api/works/" + testWorkID},
		{http.MethodPost, "/api/works"},
		{http.MethodPut, "/api/works/" + testWorkID},
		{http.MethodDelete, "/api/works/" + testWorkID},
		{http.MethodPost, "/api/works/" + testWorkID + "/publish"},
	}

	for _, request := range requests {
		t.Run(request.method+" "+request.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(request.method, request.path, nil))

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", recorder.Code)
			}
		})
	}
}

func TestWorksListUsesDraftsParam(t *testing.T) {
	works := &fakeWorks{list: []content.WorkSummary{{ID: testWorkID, Slug: "monban"}}}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodGet, "/api/works?drafts=1", "", false))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !works.includeDrafts {
		t.Error("includeDrafts = false, want true for drafts=1")
	}

	var body struct {
		Works []content.WorkSummary `json:"works"`
	}
	decodeResponse(t, recorder, &body)
	if len(body.Works) != 1 || body.Works[0].Slug != "monban" {
		t.Errorf("works = %+v, want one monban", body.Works)
	}
}

func TestWorksListHidesDraftsByDefault(t *testing.T) {
	works := &fakeWorks{}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodGet, "/api/works", "", false))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if works.includeDrafts {
		t.Error("includeDrafts = true, want false without drafts param")
	}
}

func TestWorksListEmptyReturnsArray(t *testing.T) {
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, &fakeWorks{})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodGet, "/api/works", "", false))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"works":[]`) {
		t.Errorf("body = %s, want an empty works array", recorder.Body.String())
	}
}

func TestWorksGetNotFound(t *testing.T) {
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, &fakeWorks{err: content.ErrNotFound})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodGet, "/api/works/"+testWorkID, "", false))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	var body map[string]string
	decodeResponse(t, recorder, &body)
	if body["error"] != "not found" {
		t.Errorf("error = %q, want not found", body["error"])
	}
}

func TestWorksValidationError(t *testing.T) {
	works := &fakeWorks{err: &content.ValidationError{Field: "summary", Message: "is required"}}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	request := authenticatedRequest(t, sessions, http.MethodPost, "/api/works", `{"slug":"example"}`, true)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	var body map[string]string
	decodeResponse(t, recorder, &body)
	if body["field"] != "summary" || body["error"] != "is required" {
		t.Errorf("body = %v, want field summary and message is required", body)
	}
}

func TestWorksCreateReturnsCreatedWork(t *testing.T) {
	works := &fakeWorks{}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	payload := `{"id":"11111111-1111-1111-1111-111111111111","slug":"example","title":"Example","summary":"Summary","date":"2025-01-02","updated":"2025-02-03"}`
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPost, "/api/works", payload, true))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", recorder.Code, recorder.Body.String())
	}
	if works.created == nil || works.created.Slug != "example" {
		t.Fatalf("created = %+v, want slug example", works.created)
	}
	if works.created.ID != "" {
		t.Errorf("created id = %q, want an empty id before the repository assigns one", works.created.ID)
	}
}

func TestWorksPublishRequiresCSRF(t *testing.T) {
	works := &fakeWorks{}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPost,
		"/api/works/"+testWorkID+"/publish", `{"published":true}`, false))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
	if works.published != nil {
		t.Error("Publish was called without CSRF")
	}
}

func TestWorksPublishRequiresPublishedField(t *testing.T) {
	works := &fakeWorks{}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPost,
		"/api/works/"+testWorkID+"/publish", `{}`, true))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	var body map[string]string
	decodeResponse(t, recorder, &body)
	if body["field"] != "published" {
		t.Errorf("field = %q, want published", body["field"])
	}
}

func TestWorksPublish(t *testing.T) {
	works := &fakeWorks{}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPost,
		"/api/works/"+testWorkID+"/publish", `{"published":true}`, true))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if works.published == nil || !*works.published {
		t.Fatalf("published = %v, want true", works.published)
	}
	if works.publishedID != testWorkID {
		t.Errorf("published id = %q, want %q", works.publishedID, testWorkID)
	}
}

func TestWorksDelete(t *testing.T) {
	works := &fakeWorks{}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodDelete,
		"/api/works/"+testWorkID, "", true))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if works.deletedID != testWorkID {
		t.Errorf("deleted id = %q, want %q", works.deletedID, testWorkID)
	}
}

func TestWorksUpdateUsesPathID(t *testing.T) {
	works := &fakeWorks{}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	payload := `{"id":"11111111-1111-1111-1111-111111111111","slug":"example","title":"Example","summary":"Summary","date":"2025-01-02","updated":"2025-02-03"}`
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPut,
		"/api/works/"+testWorkID, payload, true))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if works.updated == nil || works.updated.ID != testWorkID {
		t.Fatalf("updated id = %+v, want path id", works.updated)
	}
}

func TestWorksInvalidJSON(t *testing.T) {
	works := &fakeWorks{}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPost, "/api/works", `{`, true))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}

func TestWorksInternalErrorIsNotLeaked(t *testing.T) {
	works := &fakeWorks{err: errors.New("pq: relation works does not exist")}
	handler, sessions := newHandlerWithWorks(t, fakePinger{}, works)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodGet, "/api/works", "", false))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "relation") {
		t.Fatalf("response leaks the database error: %s", recorder.Body.String())
	}
	var body map[string]string
	decodeResponse(t, recorder, &body)
	if body["error"] != "internal server error" {
		t.Errorf("error = %q, want internal server error", body["error"])
	}
}
