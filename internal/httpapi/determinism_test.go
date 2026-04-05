package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func TestDeterminism_HTTPSnapshotTranscript(t *testing.T) {
	run := func() string {
		store := snapshot.NewInMemoryStore()
		mux := http.NewServeMux()
		NewSnapshotAPI(store).Register(mux)

		setReq := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(`{
  "id":"req-set-snapshot-determinism-001",
  "key":{"keyId":"`+allowedRoot+`","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"`+allowedRoot+`:note:n7c401c2:text","secureKeyId":"ok","version":"v1"},"value":"hello world"},
    {"key":{"keyId":"`+allowedRoot+`:note:n7c401c2:like:count","secureKeyId":"ok","version":"v1"},"value":"3"}
  ]
}`))
		setRec := httptest.NewRecorder()
		mux.ServeHTTP(setRec, setReq)
		if got, want := setRec.Code, http.StatusOK; got != want {
			t.Fatalf("set snapshot status mismatch: got %d want %d body=%s", got, want, setRec.Body.String())
		}

		getReq := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(`{
  "id":"req-get-snapshot-determinism-001",
  "key":{"keyId":"`+allowedRoot+`","secureKeyId":"ok"}
}`))
		getRec := httptest.NewRecorder()
		mux.ServeHTTP(getRec, getReq)
		if got, want := getRec.Code, http.StatusOK; got != want {
			t.Fatalf("get snapshot status mismatch: got %d want %d body=%s", got, want, getRec.Body.String())
		}

		return setRec.Body.String() + "\n---\n" + getRec.Body.String()
	}

	first := run()
	second := run()
	if first != second {
		t.Fatalf("deterministic HTTP transcript mismatch:\nfirst=%s\nsecond=%s", first, second)
	}
}

func TestDeterminism_WebSocketSetTranscript(t *testing.T) {
	run := func() string {
		mux := newEventFlowMux()
		server := startIPv4HTTPServer(t, mux)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn := mustDialWS(t, ctx, server.URL+"/events")
		defer func() {
			_ = conn.Close(websocket.StatusNormalClosure, "")
		}()

		mustWriteClientMessage(t, ctx, conn, `{"id":"sub-determinism-001","kind":"subscribe","rootKeys":["`+allowedRoot+`"]}`)
		subscribed := mustReadServerMessage(t, ctx, conn)

		req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(`{
  "id":"req-set-node-determinism-001",
  "rootKey":{"keyId":"`+allowedRoot+`","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"`+allowedRoot+`:note:n7c401c2:text","secureKeyId":"ok"},"value":"hello world"}
  ]
}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if got, want := rec.Code, http.StatusOK; got != want {
			t.Fatalf("node write status mismatch: got %d want %d body=%s", got, want, rec.Body.String())
		}

		event := mustReadServerMessage(t, ctx, conn)
		return subscribed + "\n---\n" + rec.Body.String() + "\n---\n" + event
	}

	first := run()
	second := run()
	if first != second {
		t.Fatalf("deterministic websocket transcript mismatch:\nfirst=%s\nsecond=%s", first, second)
	}
}

func TestDeterminism_WorkersComparison(t *testing.T) {
	one := runConcurrentNodeWorkflow(t, 1)
	many := runConcurrentNodeWorkflow(t, 6)
	if one != many {
		t.Fatalf("workers comparison mismatch:\nworkers1=%s\nworkersN=%s", one, many)
	}
}
