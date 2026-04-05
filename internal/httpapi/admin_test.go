package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func TestAdmin_ClearState_SeparatesFromProtocolSurface(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	snapshotAPI := NewSnapshotAPI(store)
	adminAPI := NewAdminAPI(store)
	mux := http.NewServeMux()
	snapshotAPI.Register(mux)
	adminAPI.Register(mux)

	root := snapshot.Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}
	store.Replace(root, []snapshot.KeyValue{
		{
			Key: snapshot.Key{
				KeyID:       root.KeyID + ":note:n7c401c2:text",
				SecureKeyID: "ok",
				Version:     "v1",
			},
			Value: "hello world",
		},
	})

	getBefore := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(`{"id":"req-before","key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"}}`))
	recBefore := httptest.NewRecorder()
	mux.ServeHTTP(recBefore, getBefore)
	if !strings.Contains(recBefore.Body.String(), `"keyValueList":[`) {
		t.Fatalf("expected snapshot data before clear-state, got %s", recBefore.Body.String())
	}
	if strings.Contains(recBefore.Body.String(), `"keyValueList":[]`) {
		t.Fatalf("expected non-empty snapshot before clear-state, got %s", recBefore.Body.String())
	}

	clearReq := httptest.NewRequest(http.MethodPut, "/admin/commands", strings.NewReader(`{
  "id":"req-clear-001",
  "commands":[{"id":"clear-state","comment":"clear all mock-server in-memory stores","arguments":["clear-state"]}]
}`))
	clearRec := httptest.NewRecorder()
	mux.ServeHTTP(clearRec, clearReq)
	if got, want := clearRec.Code, http.StatusOK; got != want {
		t.Fatalf("admin clear status mismatch: got %d want %d", got, want)
	}
	if !strings.Contains(clearRec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected ok admin response, got %s", clearRec.Body.String())
	}

	getAfter := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(`{"id":"req-after","key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"}}`))
	recAfter := httptest.NewRecorder()
	mux.ServeHTTP(recAfter, getAfter)
	if got, want := recAfter.Code, http.StatusOK; got != want {
		t.Fatalf("snapshot after clear status mismatch: got %d want %d", got, want)
	}
	if !strings.Contains(recAfter.Body.String(), `"keyValueList":[]`) {
		t.Fatalf("expected empty snapshot after clear-state, got %s", recAfter.Body.String())
	}
}

func TestAdmin_UnknownCommandReturnsInvalid(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	adminAPI := NewAdminAPI(store)
	mux := http.NewServeMux()
	adminAPI.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/admin/commands", strings.NewReader(`{
  "id":"req-unknown-001",
  "commands":[{"id":"unknown-command"}]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"status":"invalid"`) {
		t.Fatalf("expected invalid status, got %s", rec.Body.String())
	}
}
