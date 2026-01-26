package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/files"
)

type Dispatcher struct {
	searcher      *Searcher
	updater       *Updater
	deleter       *Deleter
	transactioner *Transactioner
	dryRunner     *DryRunner
	utils         *services.Utils
	parallelizer  *files.Parallelizer
}

func NewDispatcher(
	searcher *Searcher,
	updater *Updater,
	deleter *Deleter,
	transactioner *Transactioner,
	dryRunner *DryRunner,
	utils *services.Utils,
	parallelizer *files.Parallelizer,
) *Dispatcher {
	return &Dispatcher{
		searcher:      searcher,
		updater:       updater,
		deleter:       deleter,
		transactioner: transactioner,
		dryRunner:     dryRunner,
		utils:         utils,
		parallelizer:  parallelizer,
	}
}

func (dispatcher *Dispatcher) Execute(command models.Command, files []string, useTransaction bool, dryRun bool) {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if command.Pattern == nil && ((command.Action == models.SELECT ||
		command.Action == models.COUNT ||
		command.Action == models.UPDATE ||
		command.Action == models.DELETE) && !command.IsBatch) {
		fmt.Fprintf(os.Stderr, "Error: Invalid query pattern\n")
		return
	}

	if command.Action == models.UPDATE && !command.IsBatch && command.Replace == "" {
		fmt.Fprintf(os.Stderr, "Error: Invalid replacement value\n")
		return
	}

	if command.Action == models.COUNT {
		total, stats := dispatcher.searcher.Count(files, command.Pattern, command.SelectTarget)
		fmt.Printf("%d matches\n", total)
		dispatcher.utils.PrintStats(stats)
		return
	}

	if command.Action == models.SELECT {
		stats := dispatcher.searcher.Select(files, command.Pattern, command.SelectTarget)
		dispatcher.utils.PrintStats(stats)
		return
	}

	if command.Action == models.UPDATE {
		if dryRun {
			isValid := dispatcher.dryRunner.Validate(command, files, &stats, useTransaction)
			status := "fail"
			if isValid {
				status = "pass"
			}
			fmt.Printf("Dry run: %s\n", status)
			return
		}

		if useTransaction {
			var updateFunc func(string) (int, error)
			if command.IsBatch {
				updateFunc = func(file string) (int, error) {
					return dispatcher.updater.Batch(file, command.Replacements)
				}
			} else {
				updateFunc = func(file string) (int, error) {
					return dispatcher.updater.Single(file, command.Pattern, command.Replace)
				}
			}

			total := dispatcher.transactioner.Update(files, updateFunc, &stats)
			dispatcher.utils.PrintUpdateMessage(total)
			dispatcher.utils.PrintStats(stats)
			return
		}

		total := 0
		for _, file := range files {
			var count int
			var err error

			if command.IsBatch {
				count, err = dispatcher.updater.Batch(file, command.Replacements)
			} else {
				count, err = dispatcher.updater.Single(file, command.Pattern, command.Replace)
			}

			if err != nil {
				dispatcher.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		dispatcher.utils.PrintUpdateMessage(total)
		dispatcher.utils.PrintStats(stats)
		return
	}

	if command.Action == models.DELETE {
		if dryRun {
			isValid := dispatcher.dryRunner.Validate(command, files, &stats, useTransaction)
			status := "fail"
			if isValid {
				status = "pass"
			}
			fmt.Printf("Dry run: %s\n", status)
			return
		}

		if useTransaction {
			var deleteFunc func(string) (int, error)
			if command.IsBatch {
				deleteFunc = func(file string) (int, error) {
					return dispatcher.deleter.Batch(file, command.Deletions)
				}
			} else {
				deleteFunc = func(file string) (int, error) {
					return dispatcher.deleter.Single(file, command.Pattern)
				}
			}

			total := dispatcher.transactioner.Delete(files, deleteFunc, &stats)
			fmt.Printf("Deleted: %d lines\n", total)
			dispatcher.utils.PrintStats(stats)
			return
		}

		total := 0
		for _, file := range files {
			var count int
			var err error

			if command.IsBatch {
				count, err = dispatcher.deleter.Batch(file, command.Deletions)
			} else {
				count, err = dispatcher.deleter.Single(file, command.Pattern)
			}

			if err != nil {
				dispatcher.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		fmt.Printf("Deleted: %d lines\n", total)
		dispatcher.utils.PrintStats(stats)
	}
}
