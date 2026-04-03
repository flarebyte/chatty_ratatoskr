package runtimeconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ServeConfig struct {
	Listen           string
	WebSocketEnabled bool
	AdminEnabled     bool
}

type cueExport struct {
	HTTP struct {
		Port int `json:"port"`
	} `json:"http"`
	WebSocket struct {
		Supported bool `json:"supported"`
	} `json:"websocket"`
	Admin any `json:"admin"`
}

func DefaultServeConfig() ServeConfig {
	return ServeConfig{
		Listen:           "127.0.0.1:8080",
		WebSocketEnabled: false,
		AdminEnabled:     false,
	}
}

func LoadServeConfig(ctx context.Context, path string) (ServeConfig, error) {
	cfg := DefaultServeConfig()
	if path == "" {
		return cfg, nil
	}

	switch filepath.Ext(path) {
	case ".cue":
		return loadCueConfig(ctx, path)
	default:
		return loadJSONConfig(path)
	}
}

func ValidateServeConfig(cfg ServeConfig) error {
	if cfg.Listen == "" {
		return fmt.Errorf("invalid config: listen must not be empty")
	}
	return nil
}

func loadJSONConfig(path string) (ServeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ServeConfig{}, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg := DefaultServeConfig()
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return ServeConfig{}, fmt.Errorf("decode config %q: %w", path, err)
	}
	if err := ValidateServeConfig(cfg); err != nil {
		return ServeConfig{}, err
	}
	return cfg, nil
}

func loadCueConfig(ctx context.Context, path string) (ServeConfig, error) {
	cmd := exec.CommandContext(ctx, "cue", "export", path, "--out", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ServeConfig{}, fmt.Errorf("export cue config %q: %w: %s", path, err, bytes.TrimSpace(output))
	}

	var raw cueExport
	if err := json.Unmarshal(output, &raw); err != nil {
		return ServeConfig{}, fmt.Errorf("decode cue config %q: %w", path, err)
	}

	cfg := ServeConfig{
		Listen:           fmt.Sprintf("127.0.0.1:%d", raw.HTTP.Port),
		WebSocketEnabled: raw.WebSocket.Supported,
		AdminEnabled:     raw.Admin != nil,
	}
	if err := ValidateServeConfig(cfg); err != nil {
		return ServeConfig{}, err
	}
	return cfg, nil
}
