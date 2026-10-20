package default_config

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/overthinkinglabs/sqd/src/models"
)

type Loader struct {
	defaultConfig *models.DefaultConfig
}

func NewLoader(defaultConfig *models.DefaultConfig) *Loader {
	return &Loader{
		defaultConfig: defaultConfig,
	}
}

func (loader *Loader) expandTilde(path string) string {
	if !strings.HasPrefix(path, "~/") && path != "~" {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return homeDir
	}

	return filepath.Join(homeDir, path[2:])
}

func (loader *Loader) configPath() string {
	configDir := loader.expandTilde(filepath.Join("~", ".config", "sqd"))
	return filepath.Join(configDir, "config.json")
}

func (loader *Loader) LoadConfig() (*models.DefaultConfig, error) {
	config := loader.defaultConfig
	configFile := loader.configPath()
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}

		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if config.FromAliases == nil {
		config.FromAliases = make(map[string]string)
	}

	return config, nil
}
