package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DbURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {
	jsonPath, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		return Config{}, err
	}
	defer jsonFile.Close()

	var c Config
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&c); err != nil {
		return Config{}, err
	}
	return c, nil
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName
	return write(*c)
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, configFileName), nil
}

func write(c Config) error {
	jsonPath, err := getConfigPath()
	if err != nil {
		return err
	}

	jsonFile, err := os.Create(jsonPath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	return encoder.Encode(c)
}