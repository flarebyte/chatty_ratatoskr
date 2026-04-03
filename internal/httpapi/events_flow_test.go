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

func TestWebSocket_SetEventFlow(t *testing.T) {
	mux := newEventFlowMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := mustDialWS(t, ctx, server.URL+"/events")
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	mustWriteClientMessage(t, ctx, conn, `{"id":"sub-100","kind":"subscribe","rootKeys":["`+allowedRoot+`"]}`)
	_ = mustReadServerMessage(t, ctx, conn)

	req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(`{
  "id":"req-set-node-event-001",
  "rootKey":{"keyId":"`+allowedRoot+`","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"`+allowedRoot+`:note:n7c401c2:text","secureKeyId":"ok"},"value":"hello world"}
  ]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("node write status mismatch: got %d want %d", got, want)
	}

	got := mustReadServerMessage(t, ctx, conn)
	want := `{"kind":"event","event":{"eventId":"event-001","rootKey":{"keyId":"` + allowedRoot + `","secureKeyId":"ok","kind":{"hierarchy":["dashboard"]}},"operation":"set","created":"2026-04-03T10:00:00Z","key":{"keyId":"` + allowedRoot + `:note:n7c401c2:text","secureKeyId":"ok","version":"v1","kind":{"hierarchy":["dashboard","note","text"]}},"keyValue":{"key":{"keyId":"` + allowedRoot + `:note:n7c401c2:text","secureKeyId":"ok","version":"v1","kind":{"hierarchy":["dashboard","note","text"]}},"value":"hello world"}}}`
	if got != want {
		t.Fatalf("event mismatch:\nwant %s\ngot  %s", want, got)
	}
}

func TestWebSocket_SnapshotReplacedFlow(t *testing.T) {
	mux := newEventFlowMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := mustDialWS(t, ctx, server.URL+"/events")
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	mustWriteClientMessage(t, ctx, conn, `{"id":"sub-101","kind":"subscribe","rootKeys":["`+allowedRoot+`"]}`)
	_ = mustReadServerMessage(t, ctx, conn)

	req := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(`{
  "id":"req-set-snapshot-event-001",
  "key":{"keyId":"`+allowedRoot+`","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"`+allowedRoot+`:note:n7c401c2:text","secureKeyId":"ok","version":"v1"},"value":"hello world"}
  ]
}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("snapshot write status mismatch: got %d want %d", got, want)
	}

	got := mustReadServerMessage(t, ctx, conn)
	want := `{"kind":"event","event":{"eventId":"event-001","rootKey":{"keyId":"` + allowedRoot + `","secureKeyId":"ok","kind":{"hierarchy":["dashboard"]}},"operation":"snapshot-replaced","created":"2026-04-03T10:00:00Z","snapshotVersion":"snapshot-v1"}}`
	if got != want {
		t.Fatalf("event mismatch:\nwant %s\ngot  %s", want, got)
	}
}

func newEventFlowMux() *http.ServeMux {
	mux := http.NewServeMux()
	store := snapshot.NewInMemoryStore()
	events := NewEventsAPIWithOptions([]string{allowedRoot}, sequenceGenerator("event-001"), func() string {
		return "2026-04-03T10:00:00Z"
	})
	events.Register(mux)
	NewSnapshotAPIWithEvents(store, events).Register(mux)
	NewNodeAPIWithEvents(store, events).Register(mux)
	return mux
}
