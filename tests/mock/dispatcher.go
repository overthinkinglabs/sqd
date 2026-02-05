package mock

import (
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/commands"
	"github.com/albertoboccolini/sqd/services/dry_mode"
	"github.com/albertoboccolini/sqd/services/files"
)

func NewDispatcher() *commands.Dispatcher {
	utils := services.NewUtils()
	processor := files.NewProcessor(utils)

	parallelizer := files.NewParallelizer(utils)
	dryModeErrorHandler := dry_mode.NewErrorHandler()
	dryModeFileReader := dry_mode.NewFileReader(dryModeErrorHandler, utils)
	dryModeChangeDisplayer := dry_mode.NewChangeDisplayer(dryModeFileReader)
	dryModeChangeCounter := dry_mode.NewChangeCounter(dryModeFileReader)
	dryModeRunner := dry_mode.NewRunner(dryModeChangeDisplayer, dryModeChangeCounter, dryModeFileReader, dryModeErrorHandler, utils)
	transactioner := commands.NewTransactioner(utils)
	sorter := commands.NewSorter()
	searcher := commands.NewSearcher(parallelizer, sorter, utils)
	counter := commands.NewCounter(parallelizer, searcher)
	updater := commands.NewUpdater(processor, utils)
	deleter := commands.NewDeleter(processor, utils)

	return commands.NewDispatcher(
		searcher,
		counter,
		updater,
		deleter,
		transactioner,
		dryModeRunner,
		utils,
		parallelizer,
	)
}
