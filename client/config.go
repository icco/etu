package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const defaultGRPCTarget = "grpc.etu.timeclimbers.com:443"

// SignupURL is where new users create an account and API key.
const SignupURL = "https://etu.timeclimbers.com"

// ConfigFile represents the persisted config file format.
type ConfigFile struct {
	APIKey     string `json:"api_key"`
	GRPCTarget string `json:"grpc_target"`
}

// ConfigDir returns the etu config directory (e.g. ~/.config/etu on Unix).
// Creates the directory if it does not exist. Use this for config and cache files.
func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	fullDir := filepath.Join(dir, "etu")
	if err := os.MkdirAll(fullDir, 0700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	return fullDir, nil
}

// ConfigPath returns the path to the config file (~/.config/etu/config.json).
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// CachePath returns the path for a cache file under the config directory.
// Example: CachePath("timesince.cache") => ~/.config/etu/timesince.cache
func CachePath(filename string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

// loadConfigFromFile reads api_key and grpc_target from ~/.config/etu/config.json.
// A missing file returns zero values; invalid JSON returns an error.
func loadConfigFromFile() (*ConfigFile, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	// path is from ConfigPath() (fixed config dir under user home), not external input.
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is from fixed config dir, not user-controlled
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigFile{}, nil
		}
		return nil, err
	}
	var cf ConfigFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cf, nil
}

// SaveConfig writes api_key and grpc_target to ~/.config/etu/config.json.
// Creates the config directory if it does not exist.
func SaveConfig(apiKey, grpcTarget string) (*ConfigFile, error) {
	if grpcTarget == "" {
		grpcTarget = defaultGRPCTarget
	}
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	cf := &ConfigFile{APIKey: apiKey, GRPCTarget: grpcTarget}
	// Persisting the API key to the user's local config is intentional.
	data, err := json.Marshal(cf) //nolint:gosec // G117: api_key persistence is the purpose of this file
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, fmt.Errorf("could not write config file: %w", err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		return nil, fmt.Errorf("could not restrict config file permissions: %w", err)
	}
	return cf, nil
}

// SaveAPIKey stores apiKey in the config file, keeping the saved grpc_target.
func SaveAPIKey(apiKey string) error {
	cf, err := loadConfigFromFile()
	if err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		if !errors.As(err, &syntaxErr) && !errors.As(err, &typeErr) {
			return err
		}
		log.Printf("etu: replacing invalid config: %v", err)
		cf = &ConfigFile{}
	}
	_, err = SaveConfig(strings.TrimSpace(apiKey), cf.GRPCTarget)
	return err
}
