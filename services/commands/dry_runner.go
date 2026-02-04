package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
)

type DryRunner struct {
	utils *services.Utils
}

func NewDryRunner(utils *services.Utils) *DryRunner {
	return &DryRunner{utils: utils}
}

func (dryRunner *DryRunner) validateAndCount(file string, command models.Command, stats *models.ExecutionStats) (int, bool) {
	lines, ok := dryRunner.validateAndReadFile(file, stats)
	if !ok {
		return 0, false
	}

	if command.Action == models.UPDATE {
		return dryRunner.countUpdates(lines, command), true
	}

	return dryRunner.countDeletions(lines, command), true
}

func (dryRunner *DryRunner) countUpdatesInLines(lines []string, pattern *regexp.Regexp, negate bool, replace string) int {
	count := 0
	for _, line := range lines {
		matches := pattern.MatchString(line)
		if negate {
			matches = !matches
		}

		if matches {
			newLine := replace
			if !negate {
				newLine = pattern.ReplaceAllLiteralString(line, replace)
			}

			if newLine != line {
				count++
			}
		}
	}

	return count
}

func (dryRunner *DryRunner) countUpdatesInLinesInBatch(lines []string, replacements []models.Replacement) int {
	count := 0
	for _, line := range lines {
		original := line
		for _, replacement := range replacements {
			matches := replacement.Pattern.MatchString(line)
			if replacement.Negate {
				matches = !matches
			}

			if matches {
				line = replacement.Replace
				if !replacement.Negate {
					line = replacement.Pattern.ReplaceAllLiteralString(line, replacement.Replace)
				}
				break
			}
		}

		if line != original {
			count++
		}
	}

	return count
}

func (dryRunner *DryRunner) countDeletionsInLines(lines []string, pattern *regexp.Regexp, negate bool) int {
	count := 0
	for _, line := range lines {
		matches := pattern.MatchString(line)
		if negate {
			matches = !matches
		}

		if matches {
			count++
		}
	}

	return count
}

func (dryRunner *DryRunner) countDeletionsInLinesInBatch(lines []string, deletions []models.Deletion) int {
	count := 0
	for _, line := range lines {
		for _, deletion := range deletions {
			matches := deletion.Pattern.MatchString(line)
			if deletion.Negate {
				matches = !matches
			}

			if matches {
				count++
				break
			}
		}
	}

	return count
}

func (dryRunner *DryRunner) countUpdates(lines []string, command models.Command) int {
	if command.IsBatch {
		return dryRunner.countUpdatesInLinesInBatch(lines, command.Replacements)
	}

	return dryRunner.countUpdatesInLines(lines, command.Pattern, command.NegateContent, command.Replace)
}

func (dryRunner *DryRunner) countDeletions(lines []string, command models.Command) int {
	if command.IsBatch {
		return dryRunner.countDeletionsInLinesInBatch(lines, command.Deletions)
	}

	return dryRunner.countDeletionsInLines(lines, command.Pattern, command.NegateContent)
}

func (dryRunner *DryRunner) validateAndReadFile(file string, stats *models.ExecutionStats) ([]string, bool) {
	if !dryRunner.utils.IsPathInsideCwd(file) {
		dryRunner.fail("invalid path: "+file, stats)
		return nil, false
	}

	if !dryRunner.utils.CanWriteFile(file) {
		dryRunner.fail("permission denied: "+file, stats)
		return nil, false
	}

	data, err := os.ReadFile(file)
	if err != nil {
		dryRunner.fail(err.Error(), stats)
		return nil, false
	}

	return strings.Split(string(data), "\n"), true
}

func (dryRunner *DryRunner) fail(msg string, stats *models.ExecutionStats) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	stats.Skipped++
}

func (dryRunner *DryRunner) Validate(command models.Command, files []string, stats *models.ExecutionStats, useTransaction bool) bool {
	total := 0

	for _, file := range files {
		count, ok := dryRunner.validateAndCount(file, command, stats)
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
		dryRunner.utils.PrintUpdateMessage(total)
	} else {
		fmt.Printf("Deleted: %d lines\n", total)
	}

	dryRunner.utils.PrintStats(*stats)
	return true
}
