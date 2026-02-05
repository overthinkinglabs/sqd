package dry_mode

import (
	"fmt"
	"regexp"

	"github.com/albertoboccolini/sqd/models"
)

type ChangeDisplayer struct {
	fileReader *FileReader
}

func NewChangeDisplayer(fileReader *FileReader) *ChangeDisplayer {
	return &ChangeDisplayer{fileReader: fileReader}
}

func (changeDisplayer *ChangeDisplayer) printUpdateMessage(lineNumber int, file string, line string) {
	fmt.Printf("Updated line %d on %s content: %s\n", lineNumber+1, file, line)
}

func (changeDisplayer *ChangeDisplayer) printDeletionMessage(lineNumber int, file string, line string) {
	fmt.Printf("Deleted line %d on %s content: %s\n", lineNumber+1, file, line)
}

func (changeDisplayer *ChangeDisplayer) showSingleUpdates(file string, lines []string, pattern *regexp.Regexp, negate bool, replace string) {
	for lineNumber, line := range lines {
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
				changeDisplayer.printUpdateMessage(lineNumber, file, newLine)
			}
		}
	}
}

func (changeDisplayer *ChangeDisplayer) showBatchUpdates(file string, lines []string, replacements []models.Replacement) {
	for lineNumber, line := range lines {
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
			changeDisplayer.printUpdateMessage(lineNumber, file, line)
		}
	}
}

func (changeDisplayer *ChangeDisplayer) showSingleDeletions(file string, lines []string, pattern *regexp.Regexp, negate bool) {
	for lineNumber, line := range lines {
		matches := pattern.MatchString(line)
		if negate {
			matches = !matches
		}

		if matches {
			changeDisplayer.printDeletionMessage(lineNumber, file, line)
		}
	}
}

func (changeDisplayer *ChangeDisplayer) showBatchDeletions(file string, lines []string, deletions []models.Deletion) {
	for lineNumber, line := range lines {
		for _, deletion := range deletions {
			matches := deletion.Pattern.MatchString(line)
			if deletion.Negate {
				matches = !matches
			}

			if matches {
				changeDisplayer.printDeletionMessage(lineNumber, file, line)
				break
			}
		}
	}
}

func (changeDisplayer *ChangeDisplayer) ShowDeletionsForFile(file string, command models.Command) {
	lines, ok := changeDisplayer.fileReader.ValidateAndReadFile(file, &models.ExecutionStats{})
	if !ok {
		return
	}

	if command.IsBatch {
		changeDisplayer.showBatchDeletions(file, lines, command.Deletions)
		return
	}

	changeDisplayer.showSingleDeletions(file, lines, command.Pattern, command.NegateContent)
}

func (changeDisplayer *ChangeDisplayer) ShowUpdatesForFile(file string, command models.Command) {
	lines, ok := changeDisplayer.fileReader.ValidateAndReadFile(file, &models.ExecutionStats{})
	if !ok {
		return
	}

	if command.IsBatch {
		changeDisplayer.showBatchUpdates(file, lines, command.Replacements)
		return
	}

	changeDisplayer.showSingleUpdates(file, lines, command.Pattern, command.NegateContent, command.Replace)
}
