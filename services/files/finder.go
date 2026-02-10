package files

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/albertoboccolini/sqd/models"
)

type Finder struct {
	maxTextFileSize int64
	bufferSize      int
}

func NewFinder() *Finder {
	return &Finder{
		maxTextFileSize: 100 * 1024 * 1024,
		bufferSize:      8000,
	}
}

// If the file cannot be stat'ed or opened, the function returns true so that
// callers like FindFiles do not silently skip those paths.
func (finder *Finder) IsTextFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}

	if info.Size() > finder.maxTextFileSize {
		return false
	}

	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buf := make([]byte, finder.bufferSize)
	n, _ := file.Read(buf)

	for _, b := range buf[:n] {
		if b == 0 {
			return false
		}

		if b < 9 {
			return false
		}
	}

	return true
}

func (finder *Finder) FindFiles(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "*") {
		return []string{pattern}, nil
	}

	matchingPaths := []string{}
	walkErr := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			matchingPaths = append(matchingPaths, path)
		}

		return nil
	})

	if walkErr != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", walkErr)
	}

	var (
		files     []string
		mutex     sync.Mutex
		waitGroup sync.WaitGroup
		sem       = make(chan struct{}, models.MAX_CONCURRENT_GOROUTINES)
	)

	for _, path := range matchingPaths {
		waitGroup.Add(1)
		sem <- struct{}{}

		go func(p string) {
			defer waitGroup.Done()
			defer func() { <-sem }()

			if finder.IsTextFile(p) {
				mutex.Lock()
				files = append(files, p)
				mutex.Unlock()
			}
		}(path)
	}

	waitGroup.Wait()
	return files, nil
}
