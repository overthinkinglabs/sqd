package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/files"
)

type Searcher struct {
	parallelizer *files.Parallelizer
	utils        *services.Utils
}

func NewSearcher(parallelizer *files.Parallelizer, utils *services.Utils) *Searcher {
	return &Searcher{
		parallelizer: parallelizer,
		utils:        utils,
	}
}

func (searcher *Searcher) filterFilesByName(files []string, pattern *regexp.Regexp) []string {
	filtered := make([]string, 0, len(files))
	for _, file := range files {
		fileName := filepath.Base(file)
		if pattern.MatchString(fileName) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func countMatchingLines(file string, pattern *regexp.Regexp) (int, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for _, line := range lines {
		if pattern.MatchString(line) {
			count++
		}
	}

	return count, nil
}

func (searcher *Searcher) Count(files []string, command models.Command) (int, models.ExecutionStats) {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if command.WhereTarget == models.WHERE_NAME && command.WherePattern != nil {
		files = searcher.filterFilesByName(files, command.WherePattern)
	}

	switch command.SelectTarget {
	case models.NAME:
		if command.WhereTarget == models.WHERE_NAME && command.WherePattern != nil {
			stats.Processed = len(files)
			return len(files), stats
		}

		total := searcher.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			data, err := os.ReadFile(file)
			if err != nil {
				return 0, err
			}

			lines := strings.SplitSeq(string(data), "\n")
			for line := range lines {
				if command.Pattern.MatchString(line) {
					return 1, nil
				}
			}

			return 0, nil
		}, &stats)
		return total, stats
	case models.CONTENT, models.ALL:
		if command.WhereTarget == models.WHERE_NAME && command.WherePattern != nil {
			total := 0
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					stats.Skipped++
					continue
				}
				lines := strings.Split(string(data), "\n")
				total += len(lines)
				stats.Processed++
			}
			return total, stats
		}

		total := searcher.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			return countMatchingLines(file, command.Pattern)
		}, &stats)

		return total, stats
	default:
		total := searcher.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			return countMatchingLines(file, command.Pattern)
		}, &stats)

		return total, stats
	}
}

func (searcher *Searcher) Select(files []string, command models.Command) models.ExecutionStats {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if command.WhereTarget == models.WHERE_NAME && command.WherePattern != nil {
		files = searcher.filterFilesByName(files, command.WherePattern)
	}

	if command.SelectTarget == models.NAME && command.WhereTarget == models.WHERE_NAME {
		for _, file := range files {
			fmt.Printf("%s\n", searcher.utils.HighlightName(file, command.WherePattern))
		}

		stats.Processed = len(files)
		return stats
	}

	if command.WhereTarget == models.WHERE_NAME && (command.SelectTarget == models.ALL || command.SelectTarget == models.CONTENT) {
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				stats.Skipped++
				continue
			}

			lines := strings.Split(string(data), "\n")
			for i, line := range lines {
				switch command.SelectTarget {
				case models.CONTENT:
					fmt.Printf("%s\n", line)
				case models.ALL:
					fmt.Printf("%s:%d: %s\n", searcher.utils.HighlightName(file, command.WherePattern), i+1, line)
				}
			}
			stats.Processed++
		}
		return stats
	}

	searcher.parallelizer.ProcessFilesInParallelNoCount(files, func(file string) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		lines := strings.Split(string(data), "\n")
		matched := false
		for i, line := range lines {
			if command.Pattern.MatchString(line) {
				matched = true
				switch command.SelectTarget {
				case models.CONTENT:
					fmt.Printf("%s\n", searcher.utils.HighlightMatch(line, command.Pattern))
				case models.ALL:
					fmt.Printf("%s:%d: %s\n", file, i+1, searcher.utils.HighlightMatch(line, command.Pattern))
				}
			}
		}

		if matched && command.SelectTarget == models.NAME {
			fmt.Printf("%s\n", searcher.utils.HighlightName(file, command.Pattern))
		}
		return nil
	}, &stats)

	return stats
}
