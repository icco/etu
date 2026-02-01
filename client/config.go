package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const defaultGRPCTarget = "grpc.etu.natwelch.com:443"

// configFile represents the persisted config file format.
type configFile struct {
	APIKey     string `json:"api_key"`
	GRPCTarget string `json:"grpc_target"`
}

// configDir returns the etu config directory. On Unix (Linux and macOS) we use
// XDG_CONFIG_HOME or $HOME/.config so config lives at ~/.config/etu. On Windows
// we use os.UserConfigDir() (e.g. %AppData%).
func configDir() (string, error) {
	if runtime.GOOS == "windows" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("user config dir: %w", err)
		}
		return filepath.Join(dir, "etu"), nil
	}
	// Unix: XDG_CONFIG_HOME or $HOME/.config
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "etu"), nil
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// EnsureConfigFileExists creates the config file at the path from XDG_CONFIG_HOME or $HOME/.config/etu
// (~/.config/etu on Unix) with empty api_key and default grpc_target if the file does not exist.
// Call at startup so users always have a config to edit.
// Returns the path where the file was created, or "" if it already existed or on error.
func EnsureConfigFileExists() (createdPath string, err error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	_, statErr := os.Stat(path)
	if statErr == nil {
		return "", nil // file exists
	}
	if !os.IsNotExist(statErr) {
		return "", statErr
	}
	apiKey := os.Getenv("ETU_API_KEY")
	grpcTarget := os.Getenv("ETU_GRPC_TARGET")
	if grpcTarget == "" {
		grpcTarget = defaultGRPCTarget
	}
	if err := SaveConfig(apiKey, grpcTarget); err != nil {
		return path, err // return path so caller can report where we tried to write
	}
	return path, nil // just created
}

// loadConfigFromFile reads api_key and grpc_target from ~/.config/etu/config.json.
// Missing file or invalid JSON returns nil error and zero values; caller can use env or defaults.
func loadConfigFromFile() (apiKey, grpcTarget string, err error) {
	path, err := configPath()
	if err != nil {
		return "", "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("read config: %w", err)
	}
	var cf configFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return "", "", fmt.Errorf("parse config: %w", err)
	}
	return cf.APIKey, cf.GRPCTarget, nil
}

// SaveConfig writes api_key and grpc_target to ~/.config/etu/config.json.
// Creates the config directory if it does not exist.
func SaveConfig(apiKey, grpcTarget string) error {
	if grpcTarget == "" {
		grpcTarget = defaultGRPCTarget
	}
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	cf := configFile{APIKey: apiKey, GRPCTarget: grpcTarget}
	data, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ConfigDir returns the application config directory for use in error messages or docs.
func ConfigDir() (string, error) {
	return configDir()
}

// ConfigPath returns the full path to the config file (from os.UserConfigDir()/etu/config.json).
func ConfigPath() (string, error) {
	return configPath()
}
