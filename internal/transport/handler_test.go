package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vizurth/url_shortener/internal/service"
	"github.com/vizurth/url_shortener/internal/transport"
	"github.com/vizurth/url_shortener/mocks"
	"github.com/vizurth/url_shortener/pkg/logger"
)

func newTestRouter(svc service.URLService) http.Handler {
	h := transport.NewHandler(svc, "http://localhost:8080")
	log, _ := logger.New()
	return transport.NewRouter(h, log)
}

func anyCtx() any {
	return mock.MatchedBy(func(ctx context.Context) bool { return true })
}

func postJSON(t *testing.T, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func getReq(t *testing.T, path string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, path, http.NoBody)
	require.NoError(t, err)
	return req
}

func TestShorten_Created(t *testing.T) {
	t.Parallel()

	mockSvc := mocks.NewMockURLService(t)
	mockSvc.EXPECT().
		ShortenURL(anyCtx(), "https://example.com").
		Return("abc1234567", true, nil)

	rec := httptest.NewRecorder()
	newTestRouter(mockSvc).ServeHTTP(rec, postJSON(t, `{"url":"https://example.com"}`))

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "abc1234567", resp["short_code"])
	assert.Equal(t, "http://localhost:8080/abc1234567", resp["short_url"])
}

func TestShorten_AlreadyExists(t *testing.T) {
	t.Parallel()

	mockSvc := mocks.NewMockURLService(t)
	mockSvc.EXPECT().
		ShortenURL(anyCtx(), "https://example.com").
		Return("abc1234567", false, nil)

	rec := httptest.NewRecorder()
	newTestRouter(mockSvc).ServeHTTP(rec, postJSON(t, `{"url":"https://example.com"}`))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShorten_InvalidURL(t *testing.T) {
	t.Parallel()

	mockSvc := mocks.NewMockURLService(t)
	mockSvc.EXPECT().
		ShortenURL(anyCtx(), "not-a-url").
		Return("", false, service.ErrInvalidURL)

	rec := httptest.NewRecorder()
	newTestRouter(mockSvc).ServeHTTP(rec, postJSON(t, `{"url":"not-a-url"}`))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShorten_EmptyURL(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	newTestRouter(mocks.NewMockURLService(t)).ServeHTTP(rec, postJSON(t, `{"url":""}`))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShorten_InvalidBody(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	newTestRouter(mocks.NewMockURLService(t)).ServeHTTP(rec, postJSON(t, "not json"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestResolve_Redirect(t *testing.T) {
	t.Parallel()

	mockSvc := mocks.NewMockURLService(t)
	mockSvc.EXPECT().
		Resolve(anyCtx(), "abc1234567").
		Return("https://example.com", nil)

	rec := httptest.NewRecorder()
	newTestRouter(mockSvc).ServeHTTP(rec, getReq(t, "/abc1234567"))

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://example.com", rec.Header().Get("Location"))
}

func TestResolve_NotFound(t *testing.T) {
	t.Parallel()

	mockSvc := mocks.NewMockURLService(t)
	mockSvc.EXPECT().
		Resolve(anyCtx(), "abc1234567").
		Return("", service.ErrNotFound)

	rec := httptest.NewRecorder()
	newTestRouter(mockSvc).ServeHTTP(rec, getReq(t, "/abc1234567"))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
