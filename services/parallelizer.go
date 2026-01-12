package services

import (
	"sync"

	"github.com/albertoboccolini/sqd/models"
)

type Parallelizer struct {
	fileOperator *FileOperator
}

func NewParallelizer(fileOperator *FileOperator) *Parallelizer {
	return &Parallelizer{
		fileOperator: fileOperator,
	}
}

func (parallelizer *Parallelizer) processFilesInParallel(
	files []string,
	processor func(string) (int, error),
	stats *models.ExecutionStats,
) int {
	var (
		totalCount   int
		mutex        sync.Mutex
		waitingGroup sync.WaitGroup
		sem          = make(chan struct{}, models.MAX_CONCURRENT_GOROUTINES)
	)

	for _, file := range files {
		waitingGroup.Add(1)
		sem <- struct{}{}

		go func(file string) {
			defer waitingGroup.Done()
			defer func() { <-sem }()

			count, err := processor(file)

			mutex.Lock()
			if err != nil {
				parallelizer.fileOperator.utils.printProcessingErrorMessage(file, err)
				stats.Skipped++
			} else {
				totalCount += count
				stats.Processed++
			}

			mutex.Unlock()
		}(file)
	}

	waitingGroup.Wait()
	return totalCount
}

func (parallelizer *Parallelizer) processFilesInParallelNoCount(
	files []string,
	processor func(string) error,
	stats *models.ExecutionStats,
) {
	var (
		mutex        sync.Mutex
		waitingGroup sync.WaitGroup
		sem          = make(chan struct{}, models.MAX_CONCURRENT_GOROUTINES)
	)

	for _, file := range files {
		waitingGroup.Add(1)
		sem <- struct{}{}

		go func(file string) {
			defer waitingGroup.Done()
			defer func() { <-sem }()

			err := processor(file)

			mutex.Lock()
			if err != nil {
				parallelizer.fileOperator.utils.printProcessingErrorMessage(file, err)
				stats.Skipped++
			} else {
				stats.Processed++
			}

			mutex.Unlock()
		}(file)
	}

	waitingGroup.Wait()
}
