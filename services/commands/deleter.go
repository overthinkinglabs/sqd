package commands

import (
	"regexp"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/files"
)

type Deleter struct {
	processor *files.Processor
	utils     *services.Utils
}

func NewDeleter(processor *files.Processor, utils *services.Utils) *Deleter {
	return &Deleter{
		processor: processor,
		utils:     utils,
	}
}

func (deleter *Deleter) DeleteSingle(file string, pattern *regexp.Regexp) (int, error) {
	return deleter.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		filtered := []string{}
		count := 0
		for _, line := range lines {
			if !pattern.MatchString(line) {
				filtered = append(filtered, line)
				continue
			}
			count++
		}
		return filtered, count
	})
}

func (deleter *Deleter) DeleteBatch(file string, deletions []models.Deletion) (int, error) {
	return deleter.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		filtered := []string{}
		count := 0
		for _, line := range lines {
			shouldDelete := false
			for _, deletion := range deletions {
				if deletion.Pattern.MatchString(line) {
					shouldDelete = true
					count++
					break
				}
			}
			if !shouldDelete {
				filtered = append(filtered, line)
			}
		}
		return filtered, count
	})
}
