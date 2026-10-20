package commands

import (
	"regexp"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

type Updater struct {
	processor  *files.Processor
	utils      *services.Utils
	lineReader *LineReader
}

func NewUpdater(processor *files.Processor, utils *services.Utils, lineReader *LineReader) *Updater {
	return &Updater{
		processor:  processor,
		utils:      utils,
		lineReader: lineReader,
	}
}

func (updater *Updater) Single(file string, pattern *regexp.Regexp, negate bool, replace string) (int, error) {
	return updater.processor.ProcessFile(file, func(line string) (string, bool, bool, error) {
		if !updater.lineReader.MatchesCondition(line, pattern, negate) {
			return line, false, true, nil
		}

		if negate {
			return replace, true, true, nil
		}

		return pattern.ReplaceAllLiteralString(line, replace), true, true, nil
	})
}

func (updater *Updater) Batch(file string, replacements []models.Replacement) (int, error) {
	return updater.processor.ProcessFile(file, func(line string) (string, bool, bool, error) {
		for _, replacement := range replacements {
			if !updater.lineReader.MatchesCondition(line, replacement.Pattern, replacement.Negate) {
				continue
			}

			if replacement.Negate {
				return replacement.Replace, true, true, nil
			}

			return replacement.Pattern.ReplaceAllLiteralString(line, replacement.Replace), true, true, nil
		}

		return line, false, true, nil
	})
}
