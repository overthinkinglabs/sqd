package commands

import (
	"os"
	"strings"
	"time"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services/files"
)

type Counter struct {
	parallelizer *files.Parallelizer
	searcher     *Searcher
}

func NewCounter(parallelizer *files.Parallelizer, searcher *Searcher) *Counter {
	return &Counter{
		parallelizer: parallelizer,
		searcher:     searcher,
	}
}

func (counter *Counter) Count(files []string, command models.Command) (int, models.ExecutionStats) {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		files = counter.searcher.filterFilesByName(files, command.WherePattern, command.NegateFileName)
	}

	switch command.SelectTarget {
	case models.NAME:
		if command.WhereTarget == models.NAME && command.WherePattern != nil {
			stats.Processed = len(files)
			return len(files), stats
		}

		total := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			data, err := os.ReadFile(file)
			if err != nil {
				return 0, err
			}

			lines := strings.SplitSeq(string(data), "\n")
			for line := range lines {
				matches := command.Pattern.MatchString(line)
				if command.NegateContent {
					matches = !matches
				}

				if matches {
					return 1, nil
				}
			}

			return 0, nil
		}, &stats)

		return total, stats
	case models.CONTENT, models.ASTERISK:
		if command.WhereTarget == models.NAME && command.WherePattern != nil {
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

		total := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			return countMatchingLines(file, command.Pattern, command.NegateContent)
		}, &stats)

		return total, stats
	default:
		total := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			return countMatchingLines(file, command.Pattern, command.NegateContent)
		}, &stats)

		return total, stats
	}
}
