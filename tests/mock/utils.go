package mock

import (
	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
)

func NewUtils() *services.Utils {
	defaultConfig := models.NewDefaultConfig()

	return services.NewUtils(defaultConfig)
}
