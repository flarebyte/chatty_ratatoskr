package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func TestHTTP_PayloadLimitEnforced(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	api := NewSnapshotAPIWithLimit(store, 32)
	mux := http.NewServeMux()
	api.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(`{
  "id":"req-payload-too-large-001",
  "key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
  "keyValueList":[{"key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text","secureKeyId":"ok"},"value":"hello world"}]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusRequestEntityTooLarge; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}

	const want = "{\"id\":\"generated\",\"status\":\"invalid\",\"message\":\"payload too large\",\"data\":{}}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
	}
}
