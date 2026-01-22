package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

func (dispatcher *Dispatcher) filterFilesByName(files []string, pattern *regexp.Regexp) []string {
	var filtered []string
	for _, file := range files {
		if pattern.MatchString(filepath.Base(file)) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func (dispatcher *Dispatcher) filterFilesByContent(files []string, pattern *regexp.Regexp) []string {
	var filtered []string
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if pattern.Match(data) {
			filtered = append(filtered, file)
		}
	}
	return filtered
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
		total, stats := dispatcher.searcher.Count(files, command.Pattern, command.WhereTarget)
		fmt.Printf("%d lines matched\n", total)
		dispatcher.utils.PrintStats(stats)
		return
	}

	if command.Action == models.SELECT {
		stats := dispatcher.searcher.Select(files, command.Pattern, command.WhereTarget, command.SelectTarget)
		dispatcher.utils.PrintStats(stats)
		return
	}

	if command.Action == models.UPDATE {
		dispatcher.executeUpdate(command, files, useTransaction, dryRun, &stats)
		return
	}

	if command.Action == models.DELETE {
		dispatcher.executeDelete(command, files, useTransaction, dryRun, &stats)
	}
}

func (dispatcher *Dispatcher) executeUpdate(command models.Command, files []string, useTransaction bool, dryRun bool, stats *models.ExecutionStats) {
	if dryRun {
		isValid := dispatcher.dryRunner.Validate(command, files, stats, useTransaction)
		status := "fail"
		if isValid {
			status = "pass"
		}
		fmt.Printf("Dry run: %s\n", status)
		return
	}

	if command.IsBatch {
		total := 0
		for _, file := range files {
			count, err := dispatcher.updater.Batch(file, command.Replacements)
			if err != nil {
				dispatcher.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}
		dispatcher.utils.PrintUpdateMessage(total)
		dispatcher.utils.PrintStats(*stats)
		return
	}

	if command.SetTarget == models.Name {
		if command.WhereTarget == models.Name {
			files = dispatcher.filterFilesByName(files, command.Pattern)
		}
		if command.WhereTarget == models.Content {
			files = dispatcher.filterFilesByContent(files, command.Pattern)
		}

		if len(files) == 0 {
			fmt.Println("No matching files found")
			return
		}

		if len(files) > 1 {
			fmt.Fprintf(os.Stderr, "Error: Cannot rename %d files to the same name '%s'\n", len(files), command.Replace)
			return
		}

		total := 0
		for _, file := range files {
			dir := filepath.Dir(file)
			newPath := filepath.Join(dir, command.Replace)

			if _, err := os.Stat(newPath); err == nil {
				fmt.Fprintf(os.Stderr, "Error: File '%s' already exists\n", command.Replace)
				return
			}

			err := os.Rename(file, newPath)
			if err != nil {
				dispatcher.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total++
			stats.Processed++
		}
		fmt.Printf("%d files renamed\n", total)
		dispatcher.utils.PrintStats(*stats)
		return
	}

	if useTransaction {
		updateFunc := func(file string) (int, error) {
			return dispatcher.updater.Single(file, command.Pattern, command.Replace)
		}

		total := dispatcher.transactioner.Update(files, updateFunc, stats)
		dispatcher.utils.PrintUpdateMessage(total)
		dispatcher.utils.PrintStats(*stats)
		return
	}

	total := 0
	for _, file := range files {
		count, err := dispatcher.updater.Single(file, command.Pattern, command.Replace)
		if err != nil {
			dispatcher.utils.PrintProcessingErrorMessage(file, err)
			stats.Skipped++
			continue
		}
		total += count
		stats.Processed++
	}

	dispatcher.utils.PrintUpdateMessage(total)
	dispatcher.utils.PrintStats(*stats)
}

func (dispatcher *Dispatcher) executeDelete(command models.Command, files []string, useTransaction bool, dryRun bool, stats *models.ExecutionStats) {
	if dryRun {
		isValid := dispatcher.dryRunner.Validate(command, files, stats, useTransaction)
		status := "fail"
		if isValid {
			status = "pass"
		}
		fmt.Printf("Dry run: %s\n", status)
		return
	}

	if command.IsBatch {
		total := 0
		for _, file := range files {
			count, err := dispatcher.deleter.Batch(file, command.Deletions)
			if err != nil {
				dispatcher.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}
		fmt.Printf("Deleted: %d lines\n", total)
		dispatcher.utils.PrintStats(*stats)
		return
	}

	if command.WhereTarget == models.Name {
		files = dispatcher.filterFilesByName(files, command.Pattern)
		if len(files) == 0 {
			fmt.Println("No matching files found")
			return
		}

		total := 0
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				dispatcher.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total++
			stats.Processed++
		}
		fmt.Printf("Deleted: %d files\n", total)
		dispatcher.utils.PrintStats(*stats)
		return
	}

	if useTransaction {
		deleteFunc := func(file string) (int, error) {
			return dispatcher.deleter.Single(file, command.Pattern)
		}

		total := dispatcher.transactioner.Delete(files, deleteFunc, stats)
		fmt.Printf("Deleted: %d lines\n", total)
		dispatcher.utils.PrintStats(*stats)
		return
	}

	total := 0
	for _, file := range files {
		count, err := dispatcher.deleter.Single(file, command.Pattern)
		if err != nil {
			dispatcher.utils.PrintProcessingErrorMessage(file, err)
			stats.Skipped++
			continue
		}
		total += count
		stats.Processed++
	}

	fmt.Printf("Deleted: %d lines\n", total)
	dispatcher.utils.PrintStats(*stats)
}
