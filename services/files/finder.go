package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/models/displayable_errors"
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

// Returns true for files that cannot be stat or opened, ensuring these paths
// are included in the results rather than silently skipped. This defers error
// handling to upper layers (e.g., dry run or transactional services) which can
// then report these issues to the user.
//
// Note: inaccessible files will be treated as text files and may cause
// errors during subsequent read operations.
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

	defer func() { _ = file.Close() }()

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

	// TODO: This is wrong, because it will treat any error during the walk as a failure of the entire operation, even if some files were successfully found.
	// TODO: We should instead collect these errors and report them to the user while still returning any successfully found files.
	if walkErr != nil {
		return nil, displayable_errors.NewWalkError(".", walkErr)
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
