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

func (searcher *Searcher) Count(files []string, pattern *regexp.Regexp, filterType models.Filter) (int, models.ExecutionStats) {
	stats := models.ExecutionStats{StartTime: time.Now()}

	total := searcher.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
		if filterType == models.Name {
			if pattern.MatchString(filepath.Base(file)) {
				return 1, nil
			}
			return 0, nil
		}

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

func (searcher *Searcher) Select(files []string, pattern *regexp.Regexp, filterType models.Filter, selectTarget models.Select) models.ExecutionStats {
	stats := models.ExecutionStats{StartTime: time.Now()}

	searcher.parallelizer.ProcessFilesInParallelNoCount(files, func(file string) error {
		if filterType == models.Name {
			if !pattern.MatchString(filepath.Base(file)) {
				return nil
			}

			if selectTarget == models.NAME {
				fmt.Printf("%s\n", file)
				return nil
			}

			if selectTarget == models.CONTENT {
				data, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				fmt.Printf("%s", string(data))
				return nil
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			fmt.Printf("%s:\n%s\n", file, string(data))
			return nil
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if pattern.MatchString(line) {
				if selectTarget == models.NAME {
					fmt.Printf("%s\n", file)
					return nil
				}
				if selectTarget == models.CONTENT {
					fmt.Printf("%s\n", line)
					continue
				}
				fmt.Printf("%s:%d: %s\n", file, i+1, line)
			}
		}

		return nil
	}, &stats)

	return stats
}
