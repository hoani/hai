package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	aiconfig "github.com/hoani/hai/ai/config"
	"github.com/pkg/errors"
)

type Config struct {
	AI aiconfig.Config `json:"ai"`
}

type SaveOption func(*Config)

func WithOpenAIKey(key string) SaveOption {
	return func(c *Config) {
		c.AI.OpenAI.Key = key
	}
}

func Save(opts ...SaveOption) error {
	c, err := Load()
	if err != nil {
		c = &Config{} // Just assign an empty config
	}

	for _, opt := range opts {
		opt(c)
	}

	configpath, err := configPath()
	if err != nil {
		return err
	}

	f, err := os.Open(configpath)
	if errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(filepath.Dir(configpath), 0700)
		f, err = os.Create(configpath)
		if err != nil {
			return errors.Wrap(err, "cannot create config file")
		}
	}
	f.Close()

	b, err := json.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "unable to encode config")
	}

	if err := os.WriteFile(configpath, b, 0644); err != nil {
		return errors.Wrap(err, "unable to write config to file")
	}
	return nil
}

func Load() (*Config, error) {
	configpath, err := configPath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(configpath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open config file")
	}

	var c *Config
	if err := json.Unmarshal(b, c); err != nil {
		return nil, errors.Wrap(err, "config file is corrupted")
	}

	return c, nil
}

func configPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "cannot get home directory")
	}
	return filepath.Join(homedir, ".hai", "hai.json"), nil
}
