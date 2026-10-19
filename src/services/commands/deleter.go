package commands

import (
	"regexp"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

type Deleter struct {
	processor  *files.Processor
	lineReader *LineReader
	utils      *services.Utils
}

func NewDeleter(processor *files.Processor, utils *services.Utils, lineReader *LineReader) *Deleter {
	return &Deleter{
		processor:  processor,
		lineReader: lineReader,
		utils:      utils,
	}
}

func (deleter *Deleter) Single(file string, pattern *regexp.Regexp, negate bool) (int, error) {
	return deleter.processor.ProcessFile(file, func(line string) (string, bool, bool, error) {
		if deleter.lineReader.MatchesCondition(line, pattern, negate) {
			return "", true, false, nil
		}

		return line, false, true, nil
	})
}

func (deleter *Deleter) Batch(file string, deletions []models.Deletion) (int, error) {
	return deleter.processor.ProcessFile(file, func(line string) (string, bool, bool, error) {
		for _, deletion := range deletions {
			if deleter.lineReader.MatchesCondition(line, deletion.Pattern, deletion.Negate) {
				return "", true, false, nil
			}
		}

		return line, false, true, nil
	})
}
