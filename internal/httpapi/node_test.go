package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func TestHTTP_SetKeyValueList_RequestOrderAndItemStatus(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	api := NewNodeAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	reqBody := `{
  "id":"req-set-node-001",
  "rootKey":{
    "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId":"ok"
  },
  "keyValueList":[
    {
      "key":{
        "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
        "secureKeyId":"ok",
        "version":"v1"
      },
      "value":"root-is-not-a-node-child"
    },
    {
      "key":{
        "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text",
        "secureKeyId":"ok",
        "version":"v1"
      },
      "value":"hello world"
    }
  ]
}`

	req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}

	const want = "{\"id\":\"req-set-node-001\",\"status\":\"ok\",\"data\":{\"rootKey\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\"]}},\"keyList\":[{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"version\":\"v1\",\"kind\":{\"hierarchy\":[\"dashboard\"]}},\"status\":\"invalid\",\"message\":\"invalid key: node entry must be a descendant of the requested root\"},{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text\",\"secureKeyId\":\"ok\",\"version\":\"v1\",\"kind\":{\"hierarchy\":[\"dashboard\",\"note\",\"text\"]}},\"status\":\"ok\"}]}}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
	}

	got := store.Get(snapshot.Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	})
	if len(got.KeyValueList) != 1 {
		t.Fatalf("expected 1 stored node, got %d", len(got.KeyValueList))
	}
	if got.KeyValueList[0].Key.KeyID != "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text" {
		t.Fatalf("unexpected stored node key: %s", got.KeyValueList[0].Key.KeyID)
	}
}

func TestHTTP_SetKeyValueList_PartialSuccessTopLevelOK(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	api := NewNodeAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(`{
  "id":"req-set-node-002",
  "rootKey":{
    "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId":"ok"
  },
  "keyValueList":[
    {"key":{"keyId":"bad-rootless-key","secureKeyId":"ok"},"value":"x"},
    {"key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count","secureKeyId":"ok"},"value":"3"}
  ]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected top-level ok, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"invalid"`) {
		t.Fatalf("expected item-level invalid, got %s", rec.Body.String())
	}
}
