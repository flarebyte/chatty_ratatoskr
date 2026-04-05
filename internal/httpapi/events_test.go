package httpapi

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
)

const allowedRoot = "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07"

func TestWebSocket_SubscribeValidateRoots(t *testing.T) {
	ctx, conn, _ := startEventsSession(t, newEventsTestMux(), 5*time.Second)
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
	ctx, conn, _ := startEventsSession(t, newEventsTestMux(), 5*time.Second)
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
	ctx, conn, _ := startEventsSession(t, newEventsTestMux(), 5*time.Second)
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

func TestWebSocket_Heartbeat(t *testing.T) {
	pingCh := make(chan struct{}, 1)
	api := NewEventsAPIWithRuntimeOptions([]string{allowedRoot}, nil, nil, 32768, 20*time.Millisecond, time.Second)
	server := startIPv4HTTPServer(t, registerEventsAPI(api))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/events"
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		OnPingReceived: func(ctx context.Context, payload []byte) bool {
			select {
			case pingCh <- struct{}{}:
			default:
			}
			return true
		},
	})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			if _, _, err := conn.Read(ctx); err != nil {
				return
			}
		}
	}()

	select {
	case <-pingCh:
	case <-ctx.Done():
		t.Fatal("timed out waiting for heartbeat ping")
	}

	_ = conn.Close(websocket.StatusNormalClosure, "")
	<-readDone
}

func TestWebSocket_MessageSizeLimit(t *testing.T) {
	api := NewEventsAPIWithRuntimeOptions([]string{allowedRoot}, nil, nil, 32, time.Hour, time.Second)
	ctx, conn, _ := startEventsSession(t, registerEventsAPI(api), 2*time.Second)
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	mustWriteClientMessage(t, ctx, conn, `{"kind":"subscribe","rootKeys":["`+allowedRoot+`"],"padding":"abcdefghijklmnopqrstuvwxyz"}`)
	_, _, err := conn.Read(ctx)
	if err == nil {
		t.Fatal("expected websocket close error for oversized message")
	}
	if got, want := websocket.CloseStatus(err), websocket.StatusMessageTooBig; got != want {
		t.Fatalf("close status mismatch: got %v want %v err=%v", got, want, err)
	}
}

func TestWebSocket_DisconnectClearsSubscriptions(t *testing.T) {
	api := NewEventsAPIWithRuntimeOptions([]string{allowedRoot}, nil, nil, 32768, time.Hour, time.Second)
	ctx, conn, _ := startEventsSession(t, registerEventsAPI(api), 2*time.Second)
	mustWriteClientMessage(t, ctx, conn, `{"kind":"subscribe","rootKeys":["`+allowedRoot+`"]}`)
	_ = mustReadServerMessage(t, ctx, conn)
	if got, want := api.subscriberCount(), 1; got != want {
		t.Fatalf("subscriber count mismatch before close: got %d want %d", got, want)
	}
	if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("close websocket: %v", err)
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if api.subscriberCount() == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected subscriber count to reach 0 after disconnect, got %d", api.subscriberCount())
}

type localHTTPServer struct {
	URL    string
	server *http.Server
}

func (s *localHTTPServer) Close() {
	_ = s.server.Close()
}

func startEventsSession(t *testing.T, mux *http.ServeMux, timeout time.Duration) (context.Context, *websocket.Conn, *localHTTPServer) {
	t.Helper()

	server := startIPv4HTTPServer(t, mux)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	conn := mustDialWS(t, ctx, server.URL+"/events")
	return ctx, conn, server
}

func startIPv4HTTPServer(t *testing.T, mux *http.ServeMux) *localHTTPServer {
	t.Helper()

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback listener unavailable in this environment: %v", err)
	}
	server := &http.Server{Handler: mux}
	go func() {
		_ = server.Serve(listener)
	}()
	local := &localHTTPServer{
		URL:    "http://" + listener.Addr().String(),
		server: server,
	}
	t.Cleanup(local.Close)
	return local
}

func registerEventsAPI(api *EventsAPI) *http.ServeMux {
	mux := http.NewServeMux()
	api.Register(mux)
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
