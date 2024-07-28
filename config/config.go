package config

import (
	"github.com/pelletier/go-toml"
)

type Config struct {
	Git     GitConfig     `toml:"git"`
	AI      AIConfig      `toml:"ai"`
	Printer PrinterConfig `toml:"printer"`
}

type GitConfig struct {
	Provider string `toml:"provider"`
	Token    string `toml:"token"`
}

type AIConfig struct {
	Provider    string `toml:"provider"`
	OllamaURL   string `toml:"ollama_url"`
	OllamaModel string `toml:"ollama_model"`
	GroqAPIKey  string `toml:"groq_api_key"`
	GroqModel   string `toml:"groq_model"`
}

type PrinterConfig struct {
	Kind string `toml:"kind"`
}

func LoadConfig(filepath string) (*Config, error) {
	config := &Config{}
	tree, err := toml.LoadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = tree.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
