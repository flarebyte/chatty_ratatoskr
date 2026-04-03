package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func TestMockStatusForcing_SecureKeyID(t *testing.T) {
	t.Run("forced invalid on snapshot get", func(t *testing.T) {
		store := snapshot.NewInMemoryStore()
		api := NewSnapshotAPI(store)
		mux := http.NewServeMux()
		api.Register(mux)

		req := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(`{
  "id":"req-force-invalid-001",
  "key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"invalid"}
}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if got, want := rec.Code, http.StatusBadRequest; got != want {
			t.Fatalf("status mismatch: got %d want %d", got, want)
		}
		const want = "{\"id\":\"req-force-invalid-001\",\"status\":\"invalid\",\"message\":\"forced by mock secureKeyId\",\"data\":{}}\n"
		if got := rec.Body.String(); got != want {
			t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
		}
	})

	t.Run("forced unauthorised on create", func(t *testing.T) {
		api := NewCreateAPI()
		mux := http.NewServeMux()
		api.Register(mux)

		req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(`{
  "id":"req-force-unauthorised-001",
  "rootKey":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"unauthorised"},
  "newKeys":[]
}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if got, want := rec.Code, http.StatusUnauthorized; got != want {
			t.Fatalf("status mismatch: got %d want %d", got, want)
		}
		const want = "{\"id\":\"req-force-unauthorised-001\",\"status\":\"unauthorised\",\"message\":\"forced by mock secureKeyId\",\"data\":{}}\n"
		if got := rec.Body.String(); got != want {
			t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
		}
	})

	t.Run("forced outdated on node item", func(t *testing.T) {
		store := snapshot.NewInMemoryStore()
		api := NewNodeAPI(store)
		mux := http.NewServeMux()
		api.Register(mux)

		req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(`{
  "id":"req-force-outdated-001",
  "rootKey":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text","secureKeyId":"outdated"},"value":"hello world"}
  ]
}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if got, want := rec.Code, http.StatusConflict; got != want {
			t.Fatalf("status mismatch: got %d want %d", got, want)
		}
		const want = "{\"id\":\"req-force-outdated-001\",\"status\":\"outdated\",\"data\":{\"rootKey\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\"]}},\"keyList\":[{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text\",\"secureKeyId\":\"outdated\"},\"status\":\"outdated\",\"message\":\"forced by mock secureKeyId\"}]}}\n"
		if got := rec.Body.String(); got != want {
			t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
		}
	})
}
