package mock

import (
	"github.com/overthinkinglabs/sqd/services"
	"github.com/overthinkinglabs/sqd/services/dry_mode"
)

func NewDryModeRunner() *dry_mode.Runner {
	utils := services.NewUtils()
	dryModeFileReader := dry_mode.NewFileReader(utils)
	dryModeChangeProcessor := dry_mode.NewChangeProcessor(dryModeFileReader, utils)
	return dry_mode.NewRunner(dryModeChangeProcessor, utils)
}
