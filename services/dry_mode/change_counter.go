package dry_mode

import (
	"regexp"

	"github.com/albertoboccolini/sqd/models"
)

type ChangeCounter struct {
	fileReader *FileReader
}

func NewChangeCounter(fileReader *FileReader) *ChangeCounter {
	return &ChangeCounter{fileReader: fileReader}
}

func (changeCounter *ChangeCounter) validateAndCount(file string, command models.Command, stats *models.ExecutionStats) (int, bool) {
	lines, ok := changeCounter.fileReader.ValidateAndReadFile(file, stats)
	if !ok {
		return 0, false
	}

	if command.Action == models.UPDATE {
		return changeCounter.countUpdates(lines, command), true
	}

	return changeCounter.countDeletions(lines, command), true
}

func (changeCounter *ChangeCounter) countUpdatesInLines(lines []string, pattern *regexp.Regexp, negate bool, replace string) int {
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

func (changeCounter *ChangeCounter) countUpdatesInLinesInBatch(lines []string, replacements []models.Replacement) int {
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

func (changeCounter *ChangeCounter) countDeletionsInLines(lines []string, pattern *regexp.Regexp, negate bool) int {
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

func (changeCounter *ChangeCounter) countDeletionsInLinesInBatch(lines []string, deletions []models.Deletion) int {
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

func (changeCounter *ChangeCounter) countUpdates(lines []string, command models.Command) int {
	if command.IsBatch {
		return changeCounter.countUpdatesInLinesInBatch(lines, command.Replacements)
	}

	return changeCounter.countUpdatesInLines(lines, command.Pattern, command.NegateContent, command.Replace)
}

func (changeCounter *ChangeCounter) countDeletions(lines []string, command models.Command) int {
	if command.IsBatch {
		return changeCounter.countDeletionsInLinesInBatch(lines, command.Deletions)
	}

	return changeCounter.countDeletionsInLines(lines, command.Pattern, command.NegateContent)
}
