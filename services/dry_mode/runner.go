package dry_mode

import (
	"fmt"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
)

type Runner struct {
	changeDisplayer *ChangeDisplayer
	changeCounter   *ChangeCounter
	fileReader      *FileReader
	errorHandler    *ErrorHandler
	utils           *services.Utils
}

func NewRunner(changeDisplayer *ChangeDisplayer, changeCounter *ChangeCounter, fileReader *FileReader, errorHandler *ErrorHandler, utils *services.Utils) *Runner {
	return &Runner{changeDisplayer: changeDisplayer, changeCounter: changeCounter, fileReader: fileReader, errorHandler: errorHandler, utils: utils}
}

func (runner *Runner) Validate(command models.Command, files []string, stats *models.ExecutionStats, useTransaction bool, showFileNames bool) bool {
	total := 0

	for _, file := range files {
		if showFileNames {
			if command.Action == models.UPDATE {
				runner.changeDisplayer.ShowUpdatesForFile(file, command)
			}

			if command.Action == models.DELETE {
				runner.changeDisplayer.ShowDeletionsForFile(file, command)
			}
		}

		count, ok := runner.changeCounter.validateAndCount(file, command, stats)
		if !ok {
			if useTransaction {
				return false
			}

			continue
		}

		total += count
		stats.Processed++
	}

	if command.Action == models.UPDATE {
		runner.utils.PrintUpdateMessage(total)
	}

	if command.Action == models.DELETE {
		fmt.Printf("Deleted: %d lines\n", total)
	}

	runner.utils.PrintStats(*stats)
	return true
}
