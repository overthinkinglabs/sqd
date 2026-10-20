package mock

import (
	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services/default_config"
)

func NewLoader() *default_config.Loader {
	defaultConfig := models.NewDefaultConfig()
	return default_config.NewLoader(defaultConfig)
}
