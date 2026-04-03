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
	Listen                     string
	WebSocketEnabled           bool
	WebSocketMessageLimitBytes int64
	AdminEnabled               bool
	HTTPPayloadLimitBytes      int64
}

type cueExport struct {
	HTTP struct {
		Port       int   `json:"port"`
		LimitBytes int64 `json:"limitBytes"`
	} `json:"http"`
	WebSocket struct {
		Supported  bool  `json:"supported"`
		LimitBytes int64 `json:"limitBytes"`
	} `json:"websocket"`
	Admin any `json:"admin"`
}

func DefaultServeConfig() ServeConfig {
	return ServeConfig{
		Listen:                     "127.0.0.1:8080",
		WebSocketEnabled:           false,
		WebSocketMessageLimitBytes: 32768,
		AdminEnabled:               false,
		HTTPPayloadLimitBytes:      1 << 20,
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
	if cfg.HTTPPayloadLimitBytes <= 0 {
		return fmt.Errorf("invalid config: http payload limit must be greater than zero")
	}
	if cfg.WebSocketMessageLimitBytes <= 0 {
		return fmt.Errorf("invalid config: websocket message limit must be greater than zero")
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
		Listen:                     fmt.Sprintf("127.0.0.1:%d", raw.HTTP.Port),
		WebSocketEnabled:           raw.WebSocket.Supported,
		WebSocketMessageLimitBytes: DefaultServeConfig().WebSocketMessageLimitBytes,
		AdminEnabled:               raw.Admin != nil,
	}
	if raw.WebSocket.LimitBytes > 0 {
		cfg.WebSocketMessageLimitBytes = raw.WebSocket.LimitBytes
	}
	if raw.HTTP.LimitBytes > 0 {
		cfg.HTTPPayloadLimitBytes = raw.HTTP.LimitBytes
	} else {
		cfg.HTTPPayloadLimitBytes = DefaultServeConfig().HTTPPayloadLimitBytes
	}
	if err := ValidateServeConfig(cfg); err != nil {
		return ServeConfig{}, err
	}
	return cfg, nil
}
