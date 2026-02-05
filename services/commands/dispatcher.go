package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/dry_mode"
	"github.com/albertoboccolini/sqd/services/files"
)

type Dispatcher struct {
	searcher      *Searcher
	counter       *Counter
	updater       *Updater
	deleter       *Deleter
	transactioner *Transactioner
	dryModeRunner *dry_mode.Runner
	utils         *services.Utils
	parallelizer  *files.Parallelizer
}

func NewDispatcher(
	searcher *Searcher,
	counter *Counter,
	updater *Updater,
	deleter *Deleter,
	transactioner *Transactioner,
	dryModeRunner *dry_mode.Runner,
	utils *services.Utils,
	parallelizer *files.Parallelizer,
) *Dispatcher {
	return &Dispatcher{
		searcher:      searcher,
		counter:       counter,
		updater:       updater,
		deleter:       deleter,
		transactioner: transactioner,
		dryModeRunner: dryModeRunner,
		utils:         utils,
		parallelizer:  parallelizer,
	}
}

func (dispatcher *Dispatcher) Execute(command models.Command, files []string, useTransaction bool, dryRun bool, showDetailedOutputInDryMode bool) {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if (command.Action == models.UPDATE || command.Action == models.DELETE) &&
		command.WhereTarget == models.NAME {
		fmt.Fprintf(os.Stderr, "Error: UPDATE and DELETE operations cannot filter by file name. Use WHERE content = ... instead\n")
		return
	}

	if command.Pattern == nil &&
		command.WherePattern == nil &&
		((command.Action == models.SELECT ||
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
		total, stats := dispatcher.counter.Count(files, command)
		fmt.Printf("%d matches\n", total)
		dispatcher.utils.PrintStats(stats)
		return
	}

	if command.Action == models.SELECT {
		stats := dispatcher.searcher.Select(files, command)
		dispatcher.utils.PrintStats(stats)
		return
	}

	if command.Action == models.UPDATE {
		if dryRun {
			isValid := dispatcher.dryModeRunner.Validate(command, files, &stats, useTransaction, showDetailedOutputInDryMode)
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
					return dispatcher.updater.Single(file, command.Pattern, command.NegateContent, command.Replace)
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
				count, err = dispatcher.updater.Single(file, command.Pattern, command.NegateContent, command.Replace)
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
			isValid := dispatcher.dryModeRunner.Validate(command, files, &stats, useTransaction, showDetailedOutputInDryMode)
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
					return dispatcher.deleter.Single(file, command.Pattern, command.NegateContent)
				}
			}

			total := dispatcher.transactioner.Delete(files, deleteFunc, &stats)
			dispatcher.utils.PrintDeleteMessage(total)
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
				count, err = dispatcher.deleter.Single(file, command.Pattern, command.NegateContent)
			}

			if err != nil {
				dispatcher.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}

			total += count
			stats.Processed++
		}

		dispatcher.utils.PrintDeleteMessage(total)
		dispatcher.utils.PrintStats(stats)
	}
}
