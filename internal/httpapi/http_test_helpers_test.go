package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func newSnapshotTestMux(store snapshot.Store) *http.ServeMux {
	mux := http.NewServeMux()
	NewSnapshotAPI(store).Register(mux)
	return mux
}

func newNodeTestMux(store snapshot.Store) *http.ServeMux {
	mux := http.NewServeMux()
	NewNodeAPI(store).Register(mux)
	return mux
}

func serveJSONRequest(mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func requireHTTPStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if got := rec.Code; got != want {
		t.Fatalf("status mismatch: got %d want %d body=%s", got, want, rec.Body.String())
	}
}
