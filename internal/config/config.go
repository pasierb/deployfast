package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"text/template"
)

type Config struct {
	SSH        SSHConfig `json:"ssh"`
	Repository string    `json:"repository"`
	AppName    string    `json:"appName"`
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

func (c *Config) RenderTemplate(content string) (string, error) {
	tmpl, err := template.New("script").Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		return "", err
	}

	return buf.String(), nil
}
