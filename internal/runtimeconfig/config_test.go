package runtimeconfig

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfig_LoadAndValidate(t *testing.T) {
	ctx := context.Background()

	t.Run("valid cue config", func(t *testing.T) {
		cfg, err := LoadServeConfig(ctx, filepath.Join("..", "..", "testdata", "config", "basic.cue"))
		if err != nil {
			t.Fatalf("LoadServeConfig valid cue: %v", err)
		}
		if got, want := cfg.Listen, "127.0.0.1:18080"; got != want {
			t.Fatalf("Listen mismatch: got %q want %q", got, want)
		}
		if !cfg.WebSocketEnabled {
			t.Fatal("expected WebSocketEnabled true")
		}
		if !cfg.AdminEnabled {
			t.Fatal("expected AdminEnabled true")
		}
	})

	t.Run("valid json config", func(t *testing.T) {
		cfg, err := LoadServeConfig(ctx, filepath.Join("..", "..", "testdata", "config", "basic.json"))
		if err != nil {
			t.Fatalf("LoadServeConfig valid json: %v", err)
		}
		if got, want := cfg.Listen, "127.0.0.1:18081"; got != want {
			t.Fatalf("Listen mismatch: got %q want %q", got, want)
		}
		if cfg.WebSocketEnabled {
			t.Fatal("expected WebSocketEnabled false")
		}
		if cfg.AdminEnabled {
			t.Fatal("expected AdminEnabled false")
		}
	})

	t.Run("invalid cue config", func(t *testing.T) {
		_, err := LoadServeConfig(ctx, filepath.Join("..", "..", "testdata", "config", "invalid.cue"))
		if err == nil {
			t.Fatal("expected invalid cue config error")
		}
		if !strings.Contains(err.Error(), "config") {
			t.Fatalf("unexpected invalid cue error: %v", err)
		}
	})

	t.Run("invalid validation", func(t *testing.T) {
		err := ValidateServeConfig(ServeConfig{})
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !strings.Contains(err.Error(), "listen must not be empty") {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})
}
