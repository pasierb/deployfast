package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	SSH SSHConfig `json:"ssh"`
}

type SSHConfig struct {
	Host string `json:"host"`
	User string `json:"user"`
	Port int    `json:"port"`
	// Remove the Password field
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPaths := []string{
		"deployfast.json",
		filepath.Join(homeDir, "deployfast.json"),
	}

	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	if configFile == "" {
		return nil, os.ErrNotExist
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
