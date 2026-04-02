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

	const want = "{\"id\":\"req-set-snapshot-001\",\"status\":\"ok\",\"data\":{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\"]}}}}\n"
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

func TestHTTP_GetSnapshot_ReturnsDeterministicSnapshot(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	root := snapshot.Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}
	store.Replace(root, []snapshot.KeyValue{
		{
			Key: snapshot.Key{
				KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count",
				SecureKeyID: "ok",
				Version:     "v1",
			},
			Value: "3",
		},
		{
			Key: snapshot.Key{
				KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text",
				SecureKeyID: "ok",
				Version:     "v1",
			},
			Value: "hello world",
		},
	})

	api := NewSnapshotAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	reqBody := `{
  "id": "req-get-snapshot-001",
  "key": {
    "keyId": "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId": "ok"
  }
}`
	req := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}

	const want = "{\"id\":\"req-get-snapshot-001\",\"status\":\"ok\",\"data\":{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\"]}},\"keyValueList\":[{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count\",\"secureKeyId\":\"ok\",\"version\":\"v1\"},\"value\":\"3\"},{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text\",\"secureKeyId\":\"ok\",\"version\":\"v1\"},\"value\":\"hello world\"}]}}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
	}
}

func TestHTTP_GetSnapshot_MissingSnapshotReturnsEmptyList(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	api := NewSnapshotAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	reqBody := `{
  "id": "req-get-snapshot-001",
  "key": {
    "keyId": "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId": "ok"
  }
}`
	req := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"keyValueList":[]`) {
		t.Fatalf("expected empty keyValueList, got %s", rec.Body.String())
	}
}

func TestHTTP_GetSnapshot_Deterministic(t *testing.T) {
	run := func() string {
		store := snapshot.NewInMemoryStore()
		root := snapshot.Key{
			KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
			SecureKeyID: "ok",
		}
		store.Replace(root, []snapshot.KeyValue{
			{
				Key: snapshot.Key{
					KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text",
					SecureKeyID: "ok",
					Version:     "v1",
				},
				Value: "hello world",
			},
		})

		api := NewSnapshotAPI(store)
		mux := http.NewServeMux()
		api.Register(mux)
		req := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(`{"id":"req-get-snapshot-001","key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"}}`))
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

func TestHTTP_ResponseCorrelationRules(t *testing.T) {
	t.Run("echo request id on put", func(t *testing.T) {
		store := snapshot.NewInMemoryStore()
		api := NewSnapshotAPIWithGenerator(store, func() string { return "generated-001" })
		mux := http.NewServeMux()
		api.Register(mux)

		req := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(`{
  "id":"req-explicit-001",
  "key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
  "keyValueList":[]
}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if !strings.Contains(rec.Body.String(), `"id":"req-explicit-001"`) {
			t.Fatalf("expected explicit id in response, got %s", rec.Body.String())
		}
	})

	t.Run("generate id on get when request id missing", func(t *testing.T) {
		store := snapshot.NewInMemoryStore()
		api := NewSnapshotAPIWithGenerator(store, func() string { return "generated-123" })
		mux := http.NewServeMux()
		api.Register(mux)

		req := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(`{
  "key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"}
}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		const want = "{\"id\":\"generated-123\",\"status\":\"ok\",\"data\":{\"key\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\"]}},\"keyValueList\":[]}}\n"
		if got := rec.Body.String(); got != want {
			t.Fatalf("generated id response mismatch:\nwant %s\ngot  %s", want, got)
		}
	})

	t.Run("generate id on invalid request when request id missing", func(t *testing.T) {
		store := snapshot.NewInMemoryStore()
		api := NewSnapshotAPIWithGenerator(store, func() string { return "generated-invalid-001" })
		mux := http.NewServeMux()
		api.Register(mux)

		req := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(`{"key":{"keyId":123}}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if got, want := rec.Code, http.StatusBadRequest; got != want {
			t.Fatalf("status mismatch: got %d want %d", got, want)
		}
		if !strings.Contains(rec.Body.String(), `"id":"generated-invalid-001"`) {
			t.Fatalf("expected generated invalid id, got %s", rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"status":"invalid"`) {
			t.Fatalf("expected invalid status, got %s", rec.Body.String())
		}
	})
}

func TestHTTP_SnapshotKindIgnoresClientHint(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	api := NewSnapshotAPI(store)
	mux := http.NewServeMux()
	api.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(`{
  "id":"req-kind-001",
  "key":{
    "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId":"ok",
    "kind":{"hierarchy":["wrong","kind"]}
  },
  "keyValueList":[]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"kind":{"hierarchy":["dashboard"]}`) {
		t.Fatalf("expected derived kind in response, got %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"wrong"`) {
		t.Fatalf("client kind hint leaked into response: %s", rec.Body.String())
	}
}
