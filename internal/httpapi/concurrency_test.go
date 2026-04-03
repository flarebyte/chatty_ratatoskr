package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func TestConcurrent_NodeWritesAndSnapshotReads(t *testing.T) {
	store := snapshot.NewInMemoryStore()
	mux := http.NewServeMux()
	NewSnapshotAPI(store).Register(mux)
	NewNodeAPI(store).Register(mux)

	rootKey := "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07"
	const workers = 8
	const writesPerWorker = 20

	var wg sync.WaitGroup
	errCh := make(chan error, workers*2)

	for worker := range workers {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for step := range writesPerWorker {
				keyID := fmt.Sprintf("%s:note:n%04d:text", rootKey, worker*writesPerWorker+step)
				req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(fmt.Sprintf(`{
  "id":"req-node-%d-%d",
  "rootKey":{"keyId":"%s","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"%s","secureKeyId":"ok"},"value":"value-%d-%d"}
  ]
}`, worker, step, rootKey, keyID, worker, step)))
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					errCh <- fmt.Errorf("node write status=%d body=%s", rec.Code, rec.Body.String())
					return
				}
			}
		}(worker)

		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for step := range writesPerWorker {
				req := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(fmt.Sprintf(`{
  "id":"req-snapshot-%d-%d",
  "key":{"keyId":"%s","secureKeyId":"ok"}
}`, worker, step, rootKey)))
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					errCh <- fmt.Errorf("snapshot read status=%d body=%s", rec.Code, rec.Body.String())
					return
				}
			}
		}(worker)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}

	got := store.Get(snapshot.Key{KeyID: rootKey, SecureKeyID: "ok"})
	if len(got.KeyValueList) != workers*writesPerWorker {
		t.Fatalf("unexpected final node count: got %d want %d", len(got.KeyValueList), workers*writesPerWorker)
	}
}

func TestWebSocket_EventFanoutConcurrentSubscribers(t *testing.T) {
	mux := newEventFlowMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const subscribers = 4
	type received struct {
		message string
		err     error
	}
	results := make(chan received, subscribers)

	var closeFns []func()
	for range subscribers {
		conn := mustDialWS(t, ctx, server.URL+"/events")
		closeFns = append(closeFns, func() {
			_ = conn.Close(websocket.StatusNormalClosure, "")
		})
		mustWriteClientMessage(t, ctx, conn, `{"kind":"subscribe","rootKeys":["`+allowedRoot+`"]}`)
		_ = mustReadServerMessage(t, ctx, conn)

		go func(conn *websocket.Conn) {
			_, data, err := conn.Read(ctx)
			if err != nil {
				results <- received{err: err}
				return
			}
			results <- received{message: string(data)}
		}(conn)
	}
	defer func() {
		for _, closeFn := range closeFns {
			closeFn()
		}
	}()

	req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(`{
  "id":"req-set-node-fanout-001",
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

	want := `{"kind":"event","event":{"eventId":"event-001","rootKey":{"keyId":"` + allowedRoot + `","secureKeyId":"ok","kind":{"hierarchy":["dashboard"]}},"operation":"set","created":"2026-04-03T10:00:00Z","key":{"keyId":"` + allowedRoot + `:note:n7c401c2:text","secureKeyId":"ok","version":"v1","kind":{"hierarchy":["dashboard","note","text"]}},"keyValue":{"key":{"keyId":"` + allowedRoot + `:note:n7c401c2:text","secureKeyId":"ok","version":"v1","kind":{"hierarchy":["dashboard","note","text"]}},"value":"hello world"}}}`
	for range subscribers {
		select {
		case result := <-results:
			if result.err != nil {
				t.Fatalf("subscriber read failed: %v", result.err)
			}
			if result.message != want {
				t.Fatalf("fanout message mismatch:\nwant %s\ngot  %s", want, result.message)
			}
		case <-ctx.Done():
			t.Fatal("timed out waiting for websocket fanout")
		}
	}
}

func TestWorkersComparison_SnapshotBytesStable(t *testing.T) {
	one := runConcurrentNodeWorkflow(t, 1)
	many := runConcurrentNodeWorkflow(t, 6)
	if one != many {
		t.Fatalf("workers comparison mismatch:\nworkers1=%s\nworkersN=%s", one, many)
	}
}

func runConcurrentNodeWorkflow(t *testing.T, workers int) string {
	t.Helper()

	store := snapshot.NewInMemoryStore()
	mux := http.NewServeMux()
	NewSnapshotAPI(store).Register(mux)
	NewNodeAPI(store).Register(mux)

	rootKey := "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07"
	payloads := []string{
		fmt.Sprintf(`{"id":"req-a","rootKey":{"keyId":"%s","secureKeyId":"ok"},"keyValueList":[{"key":{"keyId":"%s:note:n0001:text","secureKeyId":"ok"},"value":"alpha"}]}`, rootKey, rootKey),
		fmt.Sprintf(`{"id":"req-b","rootKey":{"keyId":"%s","secureKeyId":"ok"},"keyValueList":[{"key":{"keyId":"%s:note:n0002:text","secureKeyId":"ok"},"value":"beta"}]}`, rootKey, rootKey),
		fmt.Sprintf(`{"id":"req-c","rootKey":{"keyId":"%s","secureKeyId":"ok"},"keyValueList":[{"key":{"keyId":"%s:note:n0003:text","secureKeyId":"ok"},"value":"gamma"}]}`, rootKey, rootKey),
		fmt.Sprintf(`{"id":"req-d","rootKey":{"keyId":"%s","secureKeyId":"ok"},"keyValueList":[{"key":{"keyId":"%s:note:n0004:text","secureKeyId":"ok"},"value":"delta"}]}`, rootKey, rootKey),
	}

	jobs := make(chan string, len(payloads))
	for _, payload := range payloads {
		jobs <- payload
	}
	close(jobs)

	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for payload := range jobs {
				req := httptest.NewRequest(http.MethodPut, "/node", strings.NewReader(payload))
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					errCh <- fmt.Errorf("workers comparison write status=%d body=%s", rec.Code, rec.Body.String())
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/snapshot", strings.NewReader(fmt.Sprintf(`{"id":"req-snapshot-final","key":{"keyId":"%s","secureKeyId":"ok"}}`, rootKey)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("final snapshot status mismatch: got %d want %d", got, want)
	}
	return rec.Body.String()
}
