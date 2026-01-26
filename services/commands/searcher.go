package commands

import (
	"fmt"
	"os"
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

func (searcher *Searcher) Count(files []string, pattern *regexp.Regexp, selectTarget models.Select) (int, models.ExecutionStats) {
	stats := models.ExecutionStats{StartTime: time.Now()}
	switch selectTarget {
	case models.NAME:
		total := searcher.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			data, err := os.ReadFile(file)
			if err != nil {
				return 0, err
			}

			lines := strings.SplitSeq(string(data), "\n")
			for line := range lines {
				if pattern.MatchString(line) {
					return 1, nil
				}
			}

			return 0, nil
		}, &stats)
		return total, stats
	case models.CONTENT, models.ALL:
		total := searcher.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			return countMatchingLines(file, pattern)
		}, &stats)

		return total, stats
	default:
		total := searcher.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			return countMatchingLines(file, pattern)
		}, &stats)

		return total, stats
	}
}

func (searcher *Searcher) Select(files []string, pattern *regexp.Regexp, selectTarget models.Select) models.ExecutionStats {
	stats := models.ExecutionStats{StartTime: time.Now()}

	searcher.parallelizer.ProcessFilesInParallelNoCount(files, func(file string) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		lines := strings.Split(string(data), "\n")
		matched := false
		for i, line := range lines {
			if pattern.MatchString(line) {
				matched = true
				switch selectTarget {
				case models.CONTENT:
					fmt.Printf("%s\n", line)
				case models.ALL:
					fmt.Printf("%s:%d: %s\n", file, i+1, line)
				}
			}
		}

		if matched && selectTarget == models.NAME {
			fmt.Printf("%s\n", file)
		}
		return nil
	}, &stats)

	return stats
}
