package mock

import (
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/dry_mode"
)

func NewDryModeRunner() *dry_mode.Runner {
	utils := services.NewUtils()
	dryModeErrorHandler := dry_mode.NewErrorHandler()
	dryModeFileReader := dry_mode.NewFileReader(dryModeErrorHandler, utils)
	dryModeChangeDisplayer := dry_mode.NewChangeDisplayer(dryModeFileReader)
	dryModeChangeCounter := dry_mode.NewChangeCounter(dryModeFileReader)
	return dry_mode.NewRunner(dryModeChangeDisplayer, dryModeChangeCounter, dryModeFileReader, dryModeErrorHandler, utils)
}
