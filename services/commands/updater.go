package commands

import (
	"path/filepath"
	"regexp"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/files"
)

type Updater struct {
	processor *files.Processor
	utils     *services.Utils
}

func NewUpdater(processor *files.Processor, utils *services.Utils) *Updater {
	return &Updater{
		processor: processor,
		utils:     utils,
	}
}

func (updater *Updater) Single(file string, pattern *regexp.Regexp, replace string) (int, error) {
	return updater.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		count := 0
		for i, line := range lines {
			if pattern.MatchString(line) {
				lines[i] = pattern.ReplaceAllLiteralString(line, replace)
				count++
			}
		}
		return lines, count
	})
}

func (updater *Updater) Batch(file string, replacements []models.Replacement) (int, error) {
	return updater.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		count := 0
		filename := filepath.Base(file)

		for _, replacement := range replacements {
			if replacement.WhereTarget == models.Name && !replacement.Pattern.MatchString(filename) {
				continue
			}

			if replacement.WhereTarget == models.Content {
				for i, line := range lines {
					if replacement.Pattern.MatchString(line) {
						lines[i] = replacement.Pattern.ReplaceAllLiteralString(line, replacement.Replace)
						count++
					}
				}
			}
		}
		return lines, count
	})
}
