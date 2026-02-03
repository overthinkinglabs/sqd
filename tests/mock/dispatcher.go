package mock

import (
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/commands"
	"github.com/albertoboccolini/sqd/services/files"
)

func NewDispatcher() *commands.Dispatcher {
	utils := services.NewUtils()
	processor := files.NewProcessor(utils)

	parallelizer := files.NewParallelizer(utils)
	dryRunner := commands.NewDryRunner(utils)
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
		dryRunner,
		utils,
		parallelizer,
	)
}
