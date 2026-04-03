package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
)

const allowedRoot = "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07"

func TestWebSocket_SubscribeValidateRoots(t *testing.T) {
	server := httptest.NewServer(newEventsTestMux())
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := mustDialWS(t, ctx, server.URL+"/events")
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	mustWriteClientMessage(t, ctx, conn, `{"id":"sub-001","kind":"subscribe","rootKeys":["bad-root"]}`)
	got := mustReadServerMessage(t, ctx, conn)
	want := `{"id":"sub-001","kind":"status","status":"invalid","message":"invalid rootKey: not in predefined allowed set"}`
	if got != want {
		t.Fatalf("message mismatch:\nwant %s\ngot  %s", want, got)
	}
}

func TestWebSocket_UnsubscribeNoop(t *testing.T) {
	server := httptest.NewServer(newEventsTestMux())
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := mustDialWS(t, ctx, server.URL+"/events")
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	mustWriteClientMessage(t, ctx, conn, `{"id":"sub-002","kind":"subscribe","rootKeys":["`+allowedRoot+`"]}`)
	_ = mustReadServerMessage(t, ctx, conn)
	mustWriteClientMessage(t, ctx, conn, `{"id":"unsub-001","kind":"unsubscribe","rootKeys":["`+allowedRoot+`","`+allowedRoot+`"]}`)
	got := mustReadServerMessage(t, ctx, conn)
	want := `{"id":"unsub-001","kind":"unsubscribed"}`
	if got != want {
		t.Fatalf("message mismatch:\nwant %s\ngot  %s", want, got)
	}
}

func TestWebSocket_PingPongCorrelation(t *testing.T) {
	server := httptest.NewServer(newEventsTestMux())
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := mustDialWS(t, ctx, server.URL+"/events")
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	mustWriteClientMessage(t, ctx, conn, `{"id":"ping-001","kind":"ping"}`)
	got := mustReadServerMessage(t, ctx, conn)
	want := `{"id":"ping-001","kind":"pong"}`
	if got != want {
		t.Fatalf("message mismatch:\nwant %s\ngot  %s", want, got)
	}
}

func newEventsTestMux() *http.ServeMux {
	mux := http.NewServeMux()
	NewEventsAPI([]string{allowedRoot}).Register(mux)
	return mux
}

func mustDialWS(t *testing.T, ctx context.Context, httpURL string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + httpURL[len("http"):]
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	return conn
}

func mustWriteClientMessage(t *testing.T, ctx context.Context, conn *websocket.Conn, payload string) {
	t.Helper()
	if err := conn.Write(ctx, websocket.MessageText, []byte(payload)); err != nil {
		t.Fatalf("write websocket: %v", err)
	}
}

func mustReadServerMessage(t *testing.T, ctx context.Context, conn *websocket.Conn) string {
	t.Helper()
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read websocket: %v", err)
	}
	return string(data)
}
