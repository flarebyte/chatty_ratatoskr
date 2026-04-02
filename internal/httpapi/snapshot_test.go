package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func TestHTTP_SetSnapshot_ReplacesBaseline(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	api := NewSnapshotAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	reqBody := `{
  "id": "req-set-snapshot-001",
  "key": {
    "keyId": "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId": "ok"
  },
  "keyValueList": [
    {
      "key": {
        "keyId": "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text",
        "secureKeyId": "ok",
        "version": "v1"
      },
      "value": "hello world"
    },
    {
      "key": {
        "keyId": "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count",
        "secureKeyId": "ok",
        "version": "v1"
      },
      "value": "3"
    }
  ]
}`

	req := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}

	const want = "{\"id\":\"req-set-snapshot-001\",\"status\":\"ok\",\"data\":{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\"}}}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
	}

	saved := store.Get(snapshot.Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	})
	if len(saved.KeyValueList) != 2 {
		t.Fatalf("expected 2 snapshot entries, got %d", len(saved.KeyValueList))
	}
}

func TestHTTP_SetSnapshot_InvalidPayload(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	api := NewSnapshotAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(`{"key":{"keyId":123}}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"status":"invalid"`) {
		t.Fatalf("expected invalid status body, got %s", rec.Body.String())
	}
}

func TestHTTP_SetSnapshot_Deterministic(t *testing.T) {
	requestBody := `{
  "id": "req-set-snapshot-001",
  "key": {
    "keyId": "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId": "ok"
  },
  "keyValueList": [
    {
      "key": {
        "keyId": "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text",
        "secureKeyId": "ok",
        "version": "v1"
      },
      "value": "hello world"
    }
  ]
}`

	run := func() string {
		store := snapshot.NewInMemoryStore()
		api := NewSnapshotAPI(store)
		mux := http.NewServeMux()
		api.Register(mux)
		req := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(requestBody))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Body.String()
	}

	first := run()
	second := run()
	if first != second {
		t.Fatalf("deterministic response mismatch:\nfirst=%s\nsecond=%s", first, second)
	}
}
