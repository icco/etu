package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultGRPCTarget = "grpc.etu.natwelch.com:443"

// configFile represents the persisted config file format.
type ConfigFile struct {
	APIKey     string `json:"api_key"`
	GRPCTarget string `json:"grpc_target"`
}

func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	fullDir := filepath.Join(dir, "etu")

	if err := os.MkdirAll(fullDir, 0700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	configPath := filepath.Join(fullDir, "config.json")
	return configPath, nil
}

// loadConfigFromFile reads api_key and grpc_target from ~/.config/etu/config.json.
// Missing file or invalid JSON returns nil error and zero values; caller can use env or defaults.
func loadConfigFromFile() (*ConfigFile, error) {
	path, err := ConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SaveConfig("", "")
		}
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
	data, err := json.Marshal(cf)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, fmt.Errorf("could not write config file: %w", err)
	}

	return cf, nil
}
