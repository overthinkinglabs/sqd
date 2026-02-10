package dry_mode

import (
	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
)

type Runner struct {
	changeProcessor *ChangeProcessor
	utils           *services.Utils
}

func NewRunner(changeProcessor *ChangeProcessor, utils *services.Utils) *Runner {
	return &Runner{
		changeProcessor: changeProcessor,
		utils:           utils,
	}
}

func (runner *Runner) printSummary(action models.TokenType, totalChanges int) {
	switch action {
	case models.UPDATE:
		runner.utils.PrintUpdateMessage(totalChanges)
	case models.DELETE:
		runner.utils.PrintDeleteMessage(totalChanges)
	}
}

func (runner *Runner) Validate(command models.Command, files []string, stats *models.ExecutionStats, useTransaction bool, showDetailedOutputInDryMode bool) error {
	totalChanges := 0
	errorCollection := models.NewErrorCollection()

	if showDetailedOutputInDryMode {
		runner.changeProcessor = runner.changeProcessor.WithPrinting()
	}

	for _, file := range files {
		changeCount, err := runner.changeProcessor.ProcessCommand(file, command, stats)
		if err != nil {
			errorCollection.Add(err)
			continue
		}

		totalChanges += changeCount
		stats.Processed++
	}

	runner.printSummary(command.Action, totalChanges)
	runner.utils.PrintStats(*stats)

	if errorCollection.HasErrors() {
		return errorCollection
	}

	return nil
}
