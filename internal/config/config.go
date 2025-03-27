package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	CurrentUser string
}
type State struct {
	Config *Config
}

func (c *Config) SetUser(username string) error {
	c.CurrentUser = username
	return write(c)
}

func Read() (Config, error) {

	path, err := getConfigFile()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func write(c *Config) error {
	path, err := getConfigFile()
	if err != nil {
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(c)
}

func getConfigFile() (string, error) {
	const configFileName = ".gatorconfig.json"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(homeDir, configFileName)
	return path, err
}
