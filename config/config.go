package config

import (
	"os"
	"path/filepath"

	"github.com/HanmaDevin/schlama/ollama"
	"gopkg.in/yaml.v3"
)

var home, _ = os.UserHomeDir()
var config_Path string = filepath.Dir(home + "/.config/schlama/")
var filename string = config_Path + "/config.yaml"

type Config struct {
	Model string `yaml:"model"`
}

func ReadConfig() *ollama.Ollama {
	var cfg Config
	data, err := os.ReadFile(filename)
	if err != nil {
		WriteConfig(Config{
			Model: "",
		})
	}
	// ignore errors, there shouldn't be any
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil
	}

	return parseConfig(cfg)
}

func WriteConfig(cfg Config) error {
	// Ensure the config directory exists
	if err := os.MkdirAll(config_Path, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func parseConfig(cfg Config) *ollama.Ollama {
	Body := ollama.NewOllama()
	Body.Model = cfg.Model
	Body.Messages = []ollama.Message{
		{
			Role:    "user",
			Content: "",
			Images:  nil,
		},
	}
	Body.Stream = false
	return Body
}
