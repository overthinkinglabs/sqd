package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/overthinkinglabs/sqd/models"
	"github.com/overthinkinglabs/sqd/services"
	"github.com/overthinkinglabs/sqd/services/files"
)

type Searcher struct {
	parallelizer *files.Parallelizer
	sorter       *Sorter
	utils        *services.Utils
}

type searchResult struct {
	filePath    string
	lineNumber  int
	lineContent string
}

type fileResults struct {
	results  []searchResult
	hasMatch bool
}

func NewSearcher(parallelizer *files.Parallelizer, sorter *Sorter, utils *services.Utils) *Searcher {
	return &Searcher{
		parallelizer: parallelizer,
		sorter:       sorter,
		utils:        utils,
	}
}

func (searcher *Searcher) filterFilesByName(files []string, pattern *regexp.Regexp, negate bool) []string {
	filtered := make([]string, 0, len(files))
	for _, file := range files {
		fileName := filepath.Base(file)
		matches := pattern.MatchString(fileName)
		if negate {
			matches = !matches
		}

		if matches {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func countMatchingLines(file string, pattern *regexp.Regexp, negate bool) (int, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
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

	return count, nil
}

func (searcher *Searcher) Select(files []string, command models.Command) models.ExecutionStats {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		files = searcher.filterFilesByName(files, command.WherePattern, command.NegateFileName)
	}

	if command.SelectTarget == models.NAME && command.WhereTarget == models.NAME {
		results := make([]searchResult, 0, len(files))
		for _, file := range files {
			results = append(results, searchResult{filePath: file})
		}

		searcher.sorter.sortResults(results, command.OrderBy)
		for _, result := range results {
			fmt.Printf("%s\n", searcher.utils.HighlightName(result.filePath, command.WherePattern))
		}

		stats.Processed = len(files)
		return stats
	}

	if command.WhereTarget == models.NAME && (command.SelectTarget == models.ASTERISK || command.SelectTarget == models.CONTENT) {
		results := make([]searchResult, 0)
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				stats.Skipped++
				continue
			}

			lines := strings.Split(string(data), "\n")
			for i, line := range lines {
				results = append(results, searchResult{
					filePath:    file,
					lineNumber:  i + 1,
					lineContent: line,
				})
			}
			stats.Processed++
		}

		searcher.sorter.sortResults(results, command.OrderBy)
		for _, result := range results {
			switch command.SelectTarget {
			case models.CONTENT:
				fmt.Printf("%s\n", result.lineContent)
			case models.ASTERISK:
				fmt.Printf("%s:%d: %s\n", searcher.utils.HighlightName(result.filePath, command.WherePattern), result.lineNumber, result.lineContent)
			}
		}

		return stats
	}

	allFileResults := make([]fileResults, len(files))

	searcher.parallelizer.ProcessFilesInParallelWithIndex(files, func(index int, file string) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		lines := strings.Split(string(data), "\n")
		searchResults := fileResults{results: make([]searchResult, 0)}

		for i, line := range lines {
			matches := command.Pattern.MatchString(line)
			if command.NegateContent {
				matches = !matches
			}

			if matches {
				searchResults.hasMatch = true
				searchResults.results = append(searchResults.results, searchResult{
					filePath:    file,
					lineNumber:  i + 1,
					lineContent: line,
				})
			}
		}

		allFileResults[index] = searchResults
		return nil
	}, &stats)

	results := make([]searchResult, 0)
	filesWithMatches := make([]string, 0)
	for i, searchResults := range allFileResults {
		if searchResults.hasMatch {
			results = append(results, searchResults.results...)
			if command.SelectTarget == models.NAME {
				filesWithMatches = append(filesWithMatches, files[i])
			}
		}
	}

	if command.SelectTarget == models.NAME {
		nameResults := make([]searchResult, 0, len(filesWithMatches))
		for _, file := range filesWithMatches {
			nameResults = append(nameResults, searchResult{filePath: file})
		}

		searcher.sorter.sortResults(nameResults, command.OrderBy)
		for _, result := range nameResults {
			fmt.Printf("%s\n", searcher.utils.HighlightName(result.filePath, command.Pattern))
		}
		return stats
	}

	searcher.sorter.sortResults(results, command.OrderBy)
	for _, result := range results {
		switch command.SelectTarget {
		case models.CONTENT:
			fmt.Printf("%s\n", searcher.utils.HighlightMatch(result.lineContent, command.Pattern))
		case models.ASTERISK:
			fmt.Printf("%s:%d: %s\n", result.filePath, result.lineNumber, searcher.utils.HighlightMatch(result.lineContent, command.Pattern))
		}
	}

	return stats
}
