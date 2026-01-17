package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	Endpoints string `json:"endpoints"`
	CACert    string `json:"cacert"`
	Key       string `json:"key"`
	Cert      string `json:"cert"`
}

const configDir = ".etcd-tui"
const configFile = "config.json"

var (
	customConfigPath string
	configPathMutex  sync.RWMutex
)

// SetConfigPath sets a custom config file path.
func SetConfigPath(path string) {
	configPathMutex.Lock()
	defer configPathMutex.Unlock()
	customConfigPath = path
}

func getConfigPath() (string, error) {
	configPathMutex.RLock()
	customPath := customConfigPath
	configPathMutex.RUnlock()

	if customPath != "" {
		return customPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, configDir, configFile), nil
}

func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Config file doesn't exist, return nil (will use env vars)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func GetEndpoints() string {
	cfg, _ := Load()
	if cfg != nil && cfg.Endpoints != "" {
		return cfg.Endpoints
	}
	return os.Getenv("ETCDCTL_ENDPOINTS")
}

func GetCACert() string {
	cfg, _ := Load()
	if cfg != nil && cfg.CACert != "" {
		return cfg.CACert
	}
	return os.Getenv("ETCDCTL_CACERT")
}

func GetKey() string {
	cfg, _ := Load()
	if cfg != nil && cfg.Key != "" {
		return cfg.Key
	}
	return os.Getenv("ETCDCTL_KEY")
}

func GetCert() string {
	cfg, _ := Load()
	if cfg != nil && cfg.Cert != "" {
		return cfg.Cert
	}
	return os.Getenv("ETCDCTL_CERT")
}
