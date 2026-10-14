package client

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	setTestHome(t)

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
	setTestHome(t)

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
	setTestHome(t)

	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("expected config.json, got %s", filepath.Base(path))
	}
}

func TestCachePath(t *testing.T) {
	setTestHome(t)

	path, err := CachePath("timesince.cache")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "timesince.cache" {
		t.Errorf("expected timesince.cache, got %s", filepath.Base(path))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	setTestHome(t)

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
	setTestHome(t)

	cf, err := SaveConfig("key", "")
	if err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if cf.GRPCTarget != defaultGRPCTarget {
		t.Errorf("GRPCTarget = %q, want default %q", cf.GRPCTarget, defaultGRPCTarget)
	}
}

func TestLoadConfigDoesNotCreateFile(t *testing.T) {
	setTestHome(t)

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
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("config file was created: %v", err)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	setTestHome(t)
	t.Setenv("ETU_API_KEY", "env-key")

	cfg := LoadConfig()
	if cfg.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "env-key")
	}
	if cfg.GRPCTarget != defaultGRPCTarget {
		t.Errorf("GRPCTarget = %q, want default %q", cfg.GRPCTarget, defaultGRPCTarget)
	}
}

func TestLoadConfigTrimsWhitespace(t *testing.T) {
	setTestHome(t)
	t.Setenv("ETU_API_KEY", "  my-key\n")

	cfg := LoadConfig()
	if cfg.APIKey != "my-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "my-key")
	}
}

func TestLoadConfigFileOverridesDefault(t *testing.T) {
	setTestHome(t)

	// Save a config with a custom key
	_, err := SaveConfig("file-key", "file-target:443")
	if err != nil {
		t.Fatal(err)
	}

	// Unset env vars so file values are used
	t.Setenv("ETU_API_KEY", "")
	t.Setenv("ETU_GRPC_TARGET", "")

	cfg := LoadConfig()
	if cfg.APIKey != "file-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "file-key")
	}
	if cfg.GRPCTarget != "file-target:443" {
		t.Errorf("GRPCTarget = %q, want %q", cfg.GRPCTarget, "file-target:443")
	}
}

func TestLoadConfigEnvOverridesFile(t *testing.T) {
	setTestHome(t)
	if _, err := SaveConfig("file-key", "file-target:443"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ETU_API_KEY", "env-key")
	t.Setenv("ETU_GRPC_TARGET", "env-target:443")

	cfg := LoadConfig()
	if cfg.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "env-key")
	}
	if cfg.GRPCTarget != "env-target:443" {
		t.Errorf("GRPCTarget = %q, want %q", cfg.GRPCTarget, "env-target:443")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	setTestHome(t)
	t.Setenv("ETU_API_KEY", "")
	t.Setenv("ETU_GRPC_TARGET", "")
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := LoadConfig()
	if cfg.APIKey != "" {
		t.Errorf("APIKey = %q, want empty", cfg.APIKey)
	}
	if cfg.GRPCTarget != defaultGRPCTarget {
		t.Errorf("GRPCTarget = %q, want %q", cfg.GRPCTarget, defaultGRPCTarget)
	}
}

func TestSaveConfigRestrictsPermissions(t *testing.T) {
	setTestHome(t)
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil { //nolint:gosec // test setup
		t.Fatal(err)
	}
	if _, err := SaveConfig("key", ""); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("perm = %o, want 600", perm)
	}
}

func TestSaveAPIKeyKeepsTarget(t *testing.T) {
	setTestHome(t)
	if _, err := SaveConfig("old", "custom:443"); err != nil {
		t.Fatal(err)
	}
	if err := SaveAPIKey(" new\n"); err != nil {
		t.Fatal(err)
	}
	cf, err := loadConfigFromFile()
	if err != nil {
		t.Fatal(err)
	}
	if cf.APIKey != "new" || cf.GRPCTarget != "custom:443" {
		t.Errorf("got %+v", cf)
	}
}

func TestSaveAPIKeyRepairsMalformedConfig(t *testing.T) {
	for _, key := range []string{"new-key", ""} {
		for _, data := range []string{`{not json`, `{"api_key":123}`} {
			t.Run(key+"/"+data, func(t *testing.T) {
				setTestHome(t)
				path, err := ConfigPath()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
				if err := SaveAPIKey(key); err != nil {
					t.Fatal(err)
				}
				cf, err := loadConfigFromFile()
				if err != nil {
					t.Fatal(err)
				}
				if cf.APIKey != key || cf.GRPCTarget != defaultGRPCTarget {
					t.Errorf("got %+v", cf)
				}
			})
		}
	}
}

func TestSaveAPIKeyPropagatesReadError(t *testing.T) {
	setTestHome(t)
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	var pathErr *os.PathError
	if err := SaveAPIKey("new-key"); !errors.As(err, &pathErr) || pathErr.Op != "read" {
		t.Fatalf("expected original read error, got %v", err)
	}
}

func TestSaveAPIKeyClearsTimesinceCache(t *testing.T) {
	for _, key := range []string{"new-key", ""} {
		t.Run(key, func(t *testing.T) {
			setTestHome(t)
			path, err := CachePath("timesince.cache")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("previous account's duration"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := SaveAPIKey(key); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("expected cache to be removed, got %v", err)
			}
		})
	}
}
