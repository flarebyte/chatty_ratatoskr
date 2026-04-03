package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/flarebyte/chatty-ratatoskr/internal/runtimeconfig"
)

func TestCLI_HelpAndVersion(t *testing.T) {
	restore := stubBuildInfo("1.2.3", "abc123def456", "2026-04-02T12:34:56Z")
	defer restore()

	t.Run("help command", func(t *testing.T) {
		stdout, stderr, err := runCLI(t, context.Background(), []string{"help"}, cliOptions{})
		if err != nil {
			t.Fatalf("help returned error: %v", err)
		}
		if stderr != "" {
			t.Fatalf("help wrote unexpected stderr: %q", stderr)
		}
		assertContains(t, stdout, "Repository: "+repoURL)
		assertContains(t, stdout, "serve")
		assertContains(t, stdout, "version")
	})

	t.Run("help flag", func(t *testing.T) {
		stdout, _, err := runCLI(t, context.Background(), []string{"--help"}, cliOptions{})
		if err != nil {
			t.Fatalf("--help returned error: %v", err)
		}
		assertContains(t, stdout, "flarebyte/chatty-ratatoskr")
	})

	t.Run("version command", func(t *testing.T) {
		stdout, stderr, err := runCLI(t, context.Background(), []string{"version"}, cliOptions{})
		if err != nil {
			t.Fatalf("version returned error: %v", err)
		}
		if stderr != "" {
			t.Fatalf("version wrote unexpected stderr: %q", stderr)
		}
		if got, want := stdout, "chatty version=1.2.3 commit=abc123def456 date=2026-04-02T12:34:56Z\n"; got != want {
			t.Fatalf("unexpected version output:\nwant %q\ngot  %q", want, got)
		}
	})

	t.Run("version flag", func(t *testing.T) {
		stdout, _, err := runCLI(t, context.Background(), []string{"--version"}, cliOptions{})
		if err != nil {
			t.Fatalf("--version returned error: %v", err)
		}
		assertContains(t, stdout, "version=1.2.3")
		assertContains(t, stdout, "commit=abc123def456")
		assertContains(t, stdout, "date=2026-04-02T12:34:56Z")
	})

	t.Run("unknown command points to help", func(t *testing.T) {
		_, _, err := runCLI(t, context.Background(), []string{"bogus"}, cliOptions{})
		if err == nil {
			t.Fatal("expected error for unknown command")
		}
		assertContains(t, err.Error(), `unknown command "bogus"`)
		assertContains(t, err.Error(), "chatty --help")
	})
}

func TestCLI_ServeStartsAndStops(t *testing.T) {
	t.Run("starts and stops", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		configPath := writeTempConfig(t, `{"listen":"127.0.0.1:0","websocketEnabled":false,"adminEnabled":false}`)
		ready := make(chan string, 1)
		errCh := make(chan error, 1)
		listener := newBlockingListener("127.0.0.1:0")

		go func() {
			_, _, err := runCLI(t, ctx, []string{"serve", "--config", configPath}, cliOptions{
				serveReady: func(addr string) {
					ready <- addr
				},
				listen: func(network, address string) (net.Listener, error) {
					return listener, nil
				},
			})
			errCh <- err
		}()

		var addr string
		select {
		case addr = <-ready:
		case err := <-errCh:
			t.Fatalf("serve returned early: %v", err)
		case <-time.After(2 * time.Second):
			t.Fatal("serve command did not become ready")
		}

		if got, want := addr, "127.0.0.1:0"; got != want {
			t.Fatalf("unexpected ready address: got %q want %q", got, want)
		}

		cancel()

		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("serve returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("serve command did not stop after context cancellation")
		}
	})

	t.Run("invalid config path", func(t *testing.T) {
		_, _, err := runCLI(t, context.Background(), []string{"serve", "--config", filepath.Join(t.TempDir(), "missing.json")}, cliOptions{})
		if err == nil {
			t.Fatal("expected invalid config path error")
		}
		assertContains(t, err.Error(), "read config")
	})

	t.Run("invalid config contents", func(t *testing.T) {
		configPath := writeTempConfig(t, `{"listen":123}`)
		_, _, err := runCLI(t, context.Background(), []string{"serve", "--config", configPath}, cliOptions{})
		if err == nil {
			t.Fatal("expected invalid config content error")
		}
		assertContains(t, err.Error(), "decode config")
	})

	t.Run("admin disabled omits route", func(t *testing.T) {
		cfg := writeTempConfig(t, `{"listen":"127.0.0.1:0","websocketEnabled":false,"adminEnabled":false}`)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ready := make(chan string, 1)
		errCh := make(chan error, 1)
		listener := newBlockingListener("127.0.0.1:0")
		go func() {
			_, _, err := runCLI(t, ctx, []string{"serve", "--config", cfg}, cliOptions{
				serveReady: func(addr string) {
					ready <- addr
				},
				listen: func(network, address string) (net.Listener, error) {
					return listener, nil
				},
			})
			errCh <- err
		}()

		select {
		case <-ready:
		case err := <-errCh:
			t.Fatalf("serve returned early: %v", err)
		case <-time.After(2 * time.Second):
			t.Fatal("serve command did not become ready")
		}

		mux := newServerMux(runtimeconfig.ServeConfig{
			Listen:                     "127.0.0.1:0",
			AdminEnabled:               false,
			WebSocketMessageLimitBytes: 32768,
			HTTPPayloadLimitBytes:      1 << 20,
		})
		req := httptest.NewRequest(http.MethodPut, "/admin/commands", strings.NewReader(`{"id":"req-admin","commands":[{"id":"clear-state"}]}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if got, want := rec.Code, http.StatusNotFound; got != want {
			t.Fatalf("expected admin route to be absent: got %d want %d body=%s", got, want, rec.Body.String())
		}

		cancel()
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("serve returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("serve command did not stop after context cancellation")
		}
	})

	t.Run("unsafe admin exposure rejected on non-loopback", func(t *testing.T) {
		_, _, err := runCLI(t, context.Background(), []string{"serve", "--config", writeTempConfig(t, `{"listen":"0.0.0.0:18082","websocketEnabled":false,"adminEnabled":true}`)}, cliOptions{})
		if err == nil {
			t.Fatal("expected unsafe admin exposure error")
		}
		assertContains(t, err.Error(), "loopback listen address")
	})
}

func runCLI(t *testing.T, ctx context.Context, args []string, options cliOptions) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := executeWithOptions(ctx, args, &stdout, &stderr, options)
	return stdout.String(), stderr.String(), err
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func stubBuildInfo(version, commit, date string) func() {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	Version, Commit, Date = version, commit, date
	return func() {
		Version, Commit, Date = oldVersion, oldCommit, oldDate
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}

type blockingListener struct {
	addr   net.Addr
	closed chan struct{}
	once   sync.Once
}

func newBlockingListener(address string) *blockingListener {
	return &blockingListener{
		addr:   stubAddr(address),
		closed: make(chan struct{}),
	}
}

func (l *blockingListener) Accept() (net.Conn, error) {
	<-l.closed
	return nil, net.ErrClosed
}

func (l *blockingListener) Close() error {
	l.once.Do(func() {
		close(l.closed)
	})
	return nil
}

func (l *blockingListener) Addr() net.Addr {
	return l.addr
}

type stubAddr string

func (a stubAddr) Network() string { return "tcp" }

func (a stubAddr) String() string { return string(a) }
