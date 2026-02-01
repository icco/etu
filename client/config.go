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

// loadConfigFromFile reads api_key and grpc_target from ~/.config/etu/config.json.
// Missing file or invalid JSON returns nil error and zero values; caller can use env or defaults.
func loadConfigFromFile() (apiKey, grpcTarget string, err error) {
	dir, err := os.UserConfigDir()
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
	dir, err := os.UserConfigDir()
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
