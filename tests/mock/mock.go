package mock

import (
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/commands"
	"github.com/albertoboccolini/sqd/services/files"
	"github.com/albertoboccolini/sqd/services/sql"
)

func CreateParser() *sql.Parser {
	extractor := sql.NewExtractor()
	parser := sql.NewParser(extractor)
	return parser
}

func CreateDispatcher() *commands.Dispatcher {
	utils := services.NewUtils()
	processor := files.NewProcessor(utils)

	parallelizer := files.NewParallelizer(utils)
	dryRunner := commands.NewDryRunner(utils)
	transactioner := commands.NewTransactioner(utils)
	searcher := commands.NewSearcher(parallelizer, utils)
	updater := commands.NewUpdater(processor, utils)
	deleter := commands.NewDeleter(processor, utils)

	return commands.NewDispatcher(
		searcher,
		updater,
		deleter,
		transactioner,
		dryRunner,
		utils,
		parallelizer,
	)
}
