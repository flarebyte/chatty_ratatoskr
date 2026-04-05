package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

func TestAdmin_ReadLogs(t *testing.T) {
	logs := NewLogStore(20)
	store := snapshot.NewInMemoryStore()
	mux := http.NewServeMux()
	NewSnapshotAPI(store).Register(mux)
	NewAdminAPIWithOptions(store, logs, defaultHTTPPayloadLimitBytes).Register(mux)
	handler := WithRequestLogging(mux, logs)

	setReq := httptest.NewRequest(http.MethodPut, "/snapshot", strings.NewReader(`{
  "id":"req-set-snapshot-logs-001",
  "key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07","secureKeyId":"ok"},
  "keyValueList":[
    {"key":{"keyId":"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text","secureKeyId":"ok","version":"v1"},"value":"hello world"}
  ]
}`))
	setRec := httptest.NewRecorder()
	handler.ServeHTTP(setRec, setReq)
	if got, want := setRec.Code, http.StatusOK; got != want {
		t.Fatalf("set snapshot status mismatch: got %d want %d body=%s", got, want, setRec.Body.String())
	}

	clearReq := httptest.NewRequest(http.MethodPut, "/admin/commands", strings.NewReader(`{
  "id":"req-clear-logs-001",
  "commands":[{"id":"clear-state","comment":"clear all mock-server in-memory stores","arguments":["clear-state"]}]
}`))
	clearRec := httptest.NewRecorder()
	handler.ServeHTTP(clearRec, clearReq)
	if got, want := clearRec.Code, http.StatusOK; got != want {
		t.Fatalf("clear-state status mismatch: got %d want %d body=%s", got, want, clearRec.Body.String())
	}

	readReq := httptest.NewRequest(http.MethodGet, "/admin/commands", strings.NewReader(`{
  "id":"req-read-logs-001",
  "command":{"id":"read-logs","comment":"Read the logs","arguments":["logs"]}
}`))
	readRec := httptest.NewRecorder()
	handler.ServeHTTP(readRec, readReq)
	if got, want := readRec.Code, http.StatusOK; got != want {
		t.Fatalf("read-logs status mismatch: got %d want %d body=%s", got, want, readRec.Body.String())
	}

	goldenPath := filepath.Join("..", "..", "testdata", "admin", "read-logs.response.golden.json")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden fixture: %v", err)
	}
	if got := compactJSON(t, readRec.Body.String()); got != compactJSON(t, string(want)) {
		t.Fatalf("read-logs response mismatch:\nwant %s\ngot  %s", string(want), readRec.Body.String())
	}
}

func compactJSON(t *testing.T, raw string) string {
	t.Helper()

	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatalf("parse json: %v", err)
	}
	compacted, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal compact json: %v", err)
	}
	return string(compacted)
}
