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

func TestHTTP_SetKeyValueList_RejectsStaleWrite(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	root := snapshot.Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}
	store.Upsert(root, snapshot.KeyValue{
		Key: snapshot.Key{
			KeyID:       root.KeyID + ":note:n7c401c2:text",
			SecureKeyID: "ok",
			Version:     "v2",
		},
		Value: "latest",
	})

	api := NewNodeAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(`{
  "id":"req-set-node-stale-001",
  "rootKey":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text","secureKeyId":"ok","version":"v1"},"value":"stale update"}
  ]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusConflict; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}

	const want = "{\"id\":\"req-set-node-stale-001\",\"status\":\"outdated\",\"data\":{\"rootKey\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\"]}},\"keyList\":[{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text\",\"secureKeyId\":\"ok\",\"version\":\"v1\",\"kind\":{\"hierarchy\":[\"dashboard\",\"note\",\"text\"]}},\"status\":\"outdated\",\"message\":\"outdated version: write is not based on the latest stored version\"}]}}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
	}

	stored, ok := store.Find(root, root.KeyID+":note:n7c401c2:text")
	if !ok {
		t.Fatal("expected stored node to remain present")
	}
	if stored.Key.Version != "v2" || stored.Value != "latest" {
		t.Fatalf("stale write should not replace stored value, got version=%s value=%s", stored.Key.Version, stored.Value)
	}
}

func TestHTTP_SetKeyValueList_BumpsVersionOnSuccess(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	root := snapshot.Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}
	store.Upsert(root, snapshot.KeyValue{
		Key: snapshot.Key{
			KeyID:       root.KeyID + ":note:n7c401c2:text",
			SecureKeyID: "ok",
			Version:     "v1",
		},
		Value: "before",
	})

	api := NewNodeAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(`{
  "id":"req-set-node-bump-001",
  "rootKey":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text","secureKeyId":"ok","version":"v1"},"value":"after"}
  ]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"version":"v2"`) {
		t.Fatalf("expected bumped version in response, got %s", rec.Body.String())
	}

	stored, ok := store.Find(root, root.KeyID+":note:n7c401c2:text")
	if !ok {
		t.Fatal("expected stored node after successful write")
	}
	if stored.Key.Version != "v2" || stored.Value != "after" {
		t.Fatalf("successful write should bump to v2 and update value, got version=%s value=%s", stored.Key.Version, stored.Value)
	}
}

func TestHTTP_GetKeyValueList_ReturnsRequestedValues(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	root := snapshot.Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}
	store.Upsert(root, snapshot.KeyValue{
		Key: snapshot.Key{
			KeyID:       root.KeyID + ":note:n7c401c2:like:count",
			SecureKeyID: "ok",
			Version:     "v2",
		},
		Value: "3",
	})
	store.Upsert(root, snapshot.KeyValue{
		Key: snapshot.Key{
			KeyID:       root.KeyID + ":note:n7c401c2:text",
			SecureKeyID: "ok",
			Version:     "v1",
		},
		Value: "hello world",
	})

	api := NewNodeAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	reqBody := `{
  "id":"req-get-node-001",
  "rootKey":{
    "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId":"ok"
  },
  "keyList":[
    {
      "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text",
      "secureKeyId":"ok"
    },
    {
      "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:missing:text",
      "secureKeyId":"ok"
    },
    {
      "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count",
      "secureKeyId":"ok"
    }
  ]
}`

	req := httptest.NewRequest(http.MethodGet, "/node", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}

	const want = "{\"id\":\"req-get-node-001\",\"status\":\"ok\",\"data\":{\"rootKey\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\"]}},\"keyValueList\":[{\"keyValue\":{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text\",\"secureKeyId\":\"ok\",\"version\":\"v1\",\"kind\":{\"hierarchy\":[\"dashboard\",\"note\",\"text\"]}},\"value\":\"hello world\"},\"status\":\"ok\"},{\"keyValue\":{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:missing:text\",\"secureKeyId\":\"ok\"}},\"status\":\"invalid\",\"message\":\"invalid key: unsupported label \\\"missing\\\"\"},{\"keyValue\":{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count\",\"secureKeyId\":\"ok\",\"version\":\"v2\",\"kind\":{\"hierarchy\":[\"dashboard\",\"note\",\"like\",\"count\"]}},\"value\":\"3\"},\"status\":\"ok\"}]}}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
	}
}

func TestHTTP_GetKeyValueList_DeterministicRequestOrder(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	root := snapshot.Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}
	store.Upsert(root, snapshot.KeyValue{
		Key: snapshot.Key{
			KeyID:       root.KeyID + ":note:n7c401c2:text",
			SecureKeyID: "ok",
			Version:     "v1",
		},
		Value: "hello world",
	})

	api := NewNodeAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	run := func() string {
		req := httptest.NewRequest(http.MethodGet, "/node", strings.NewReader(`{
  "id":"req-get-node-002",
  "rootKey":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
  "keyList":[
    {"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text","secureKeyId":"ok"},
    {"keyId":"bad-rootless-key","secureKeyId":"ok"}
  ]
}`))
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
