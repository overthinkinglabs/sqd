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

func (searcher *Searcher) Count(files []string, pattern *regexp.Regexp) (int, models.ExecutionStats) {
	stats := models.ExecutionStats{StartTime: time.Now()}

	total := searcher.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
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
	}, &stats)

	return total, stats
}

func (searcher *Searcher) Select(files []string, pattern *regexp.Regexp) models.ExecutionStats {
	stats := models.ExecutionStats{StartTime: time.Now()}

	searcher.parallelizer.ProcessFilesInParallelNoCount(files, func(file string) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if pattern.MatchString(line) {
				fmt.Printf("%s:%d: %s\n", file, i+1, line)
			}
		}

		return nil
	}, &stats)

	return stats
}
