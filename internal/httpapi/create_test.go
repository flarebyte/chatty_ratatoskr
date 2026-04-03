package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTP_CreateNewKeys_PreservesLocalKeyIDAndOrder(t *testing.T) {
	api := NewCreateAPIWithGenerator(sequenceGenerator("n9001", "n9002"))
	mux := http.NewServeMux()
	api.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(`{
  "id":"req-create-001",
  "rootKey":{
    "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
    "secureKeyId":"ok"
  },
  "newKeys":[
    {
      "key":{
        "localKeyId":"tmp-note-1",
        "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
        "secureKeyId":"ok"
      },
      "expectedKind":"note",
      "children":[
        {"localKeyId":"tmp-text-1","expectedKind":"text"},
        {"localKeyId":"tmp-thumb-1","expectedKind":"thumbnail"}
      ]
    },
    {
      "key":{
        "localKeyId":"tmp-note-2",
        "keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07"
      },
      "expectedKind":"invalid-kind",
      "children":[]
    }
  ]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status mismatch: got %d want %d", got, want)
	}

	const want = "{\"id\":\"req-create-001\",\"status\":\"ok\",\"data\":{\"rootKey\":{\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\"]}},\"newKeys\":[{\"key\":{\"localKeyId\":\"tmp-note-1\",\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n9001\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\",\"note\"]}},\"status\":\"ok\",\"children\":[{\"key\":{\"localKeyId\":\"tmp-text-1\",\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n9001:text\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\",\"note\",\"text\"]}},\"status\":\"ok\"},{\"key\":{\"localKeyId\":\"tmp-thumb-1\",\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n9001:thumbnail:_\",\"secureKeyId\":\"ok\",\"kind\":{\"hierarchy\":[\"dashboard\",\"note\",\"thumbnail\"]}},\"status\":\"ok\"}]},{\"key\":{\"localKeyId\":\"tmp-note-2\",\"keyId\":\"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07\"},\"status\":\"invalid\",\"message\":\"invalid expectedKind: unsupported create kind \\\"invalid-kind\\\"\",\"children\":[]}]}}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("response mismatch:\nwant %s\ngot  %s", want, got)
	}
}

func TestHTTP_CreateNewKeys_DeterministicGenerator(t *testing.T) {
	run := func() string {
		api := NewCreateAPIWithGenerator(sequenceGenerator("n7001"))
		mux := http.NewServeMux()
		api.Register(mux)

		req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(`{
  "id":"req-create-002",
  "rootKey":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
  "newKeys":[
    {
      "key":{"localKeyId":"tmp-note-1","keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
      "expectedKind":"note",
      "children":[{"localKeyId":"tmp-text-1","expectedKind":"text"}]
    }
  ]
}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Body.String()
	}

	first := run()
	second := run()
	if first != second {
		t.Fatalf("deterministic create mismatch:\nfirst=%s\nsecond=%s", first, second)
	}
}

func sequenceGenerator(ids ...string) func() string {
	index := 0
	return func() string {
		if index >= len(ids) {
			return "generated-overflow"
		}
		out := ids[index]
		index++
		return out
	}
}
