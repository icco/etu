package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir: %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("returned non-absolute path: %s", dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("path is not a directory")
	}
}

func TestConfigDirIdempotent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dir1, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	dir2, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir1 != dir2 {
		t.Errorf("ConfigDir not idempotent: %q != %q", dir1, dir2)
	}
}

func TestConfigPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("expected config.json, got %s", filepath.Base(path))
	}
}

func TestCachePath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path, err := CachePath("timesince.cache")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "timesince.cache" {
		t.Errorf("expected timesince.cache, got %s", filepath.Base(path))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cf, err := SaveConfig("test-api-key", "localhost:50051")
	if err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if cf.APIKey != "test-api-key" {
		t.Errorf("APIKey = %q, want %q", cf.APIKey, "test-api-key")
	}
	if cf.GRPCTarget != "localhost:50051" {
		t.Errorf("GRPCTarget = %q, want %q", cf.GRPCTarget, "localhost:50051")
	}

	loaded, err := loadConfigFromFile()
	if err != nil {
		t.Fatalf("loadConfigFromFile: %v", err)
	}
	if loaded.APIKey != "test-api-key" {
		t.Errorf("loaded APIKey = %q, want %q", loaded.APIKey, "test-api-key")
	}
	if loaded.GRPCTarget != "localhost:50051" {
		t.Errorf("loaded GRPCTarget = %q, want %q", loaded.GRPCTarget, "localhost:50051")
	}
}

func TestSaveConfigDefaultTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cf, err := SaveConfig("key", "")
	if err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if cf.GRPCTarget != defaultGRPCTarget {
		t.Errorf("GRPCTarget = %q, want default %q", cf.GRPCTarget, defaultGRPCTarget)
	}
}

func TestLoadConfigCreatesFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cf, err := loadConfigFromFile()
	if err != nil {
		t.Fatalf("loadConfigFromFile: %v", err)
	}
	if cf == nil {
		t.Fatal("expected non-nil config")
	}

	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ETU_API_KEY", "env-key")

	cfg := LoadConfig()
	if cfg.ApiKey != "env-key" {
		t.Errorf("ApiKey = %q, want %q", cfg.ApiKey, "env-key")
	}
	// GRPCTarget defaults from the config file (SaveConfig sets it),
	// so ETU_GRPC_TARGET env only applies when the file has an empty target.
	if cfg.GRPCTarget != defaultGRPCTarget {
		t.Errorf("GRPCTarget = %q, want default %q", cfg.GRPCTarget, defaultGRPCTarget)
	}
}

func TestLoadConfigTrimsWhitespace(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ETU_API_KEY", "  my-key\n")

	cfg := LoadConfig()
	if cfg.ApiKey != "my-key" {
		t.Errorf("ApiKey = %q, want %q", cfg.ApiKey, "my-key")
	}
}

func TestLoadConfigFileOverridesDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Save a config with a custom key
	_, err := SaveConfig("file-key", "file-target:443")
	if err != nil {
		t.Fatal(err)
	}

	// Unset env vars so file values are used
	t.Setenv("ETU_API_KEY", "")
	t.Setenv("ETU_GRPC_TARGET", "")

	cfg := LoadConfig()
	if cfg.ApiKey != "file-key" {
		t.Errorf("ApiKey = %q, want %q", cfg.ApiKey, "file-key")
	}
	if cfg.GRPCTarget != "file-target:443" {
		t.Errorf("GRPCTarget = %q, want %q", cfg.GRPCTarget, "file-target:443")
	}
}
