package files

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
)

type Processor struct {
	defaultConfig *models.DefaultConfig
	utils         *services.Utils
}

func NewProcessor(defaultConfig *models.DefaultConfig, utils *services.Utils) *Processor {
	return &Processor{defaultConfig: defaultConfig, utils: utils}
}

// ProcessFile applies transformFunc to each line of the file. Files that fit in memory are read and rewritten on the
// same descriptor for speed, larger files are processed line-by-line into a temporary file and then atomically renamed
// over the original, avoiding OOM on multi-gigabyte logs.
func (processor *Processor) ProcessFile(
	filename string,
	transformFunc func(string) (string, bool, bool, error),
) (int, error) {
	if !processor.utils.IsPathInsideCwd(filename) {
		return 0, displayable_errors.NewInvalidPathError(filename)
	}

	if !processor.utils.CanWriteFile(filename) {
		return 0, displayable_errors.NewPermissionDeniedError(filename)
	}

	file, err := os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	if info.Size() <= int64(processor.defaultConfig.MaxFileSizeInMemory)*1024*1024 {
		return processor.processFileInMemory(file, info, transformFunc)
	}

	return processor.processFileStreaming(filename, file, info, transformFunc)
}

func (processor *Processor) processFileInMemory(
	file *os.File,
	info os.FileInfo,
	transformFunc func(string) (string, bool, bool, error),
) (int, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return 0, processor.mapFileError(info.Name(), err)
	}

	lines := strings.Split(string(data), "\n")
	newLines, count, err := processor.transformLines(lines, transformFunc)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	if err := file.Truncate(0); err != nil {
		return 0, processor.mapFileError(info.Name(), err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return 0, processor.mapFileError(info.Name(), err)
	}

	if _, err := file.WriteString(strings.Join(newLines, "\n")); err != nil {
		return 0, processor.mapFileError(info.Name(), err)
	}

	return count, nil
}

func (processor *Processor) processFileStreaming(
	filename string,
	sourceFile *os.File,
	sourceInfo os.FileInfo,
	transformFunc func(string) (string, bool, bool, error),
) (int, error) {
	tempFile, err := os.CreateTemp(filepath.Dir(filename), ".sqd_proc_*")
	if err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	tempName := tempFile.Name()
	cleanupTemp := func() {
		_ = tempFile.Close()
		_ = os.Remove(tempName)
	}

	if err := os.Chmod(tempName, sourceInfo.Mode().Perm()); err != nil {
		cleanupTemp()
		return 0, processor.mapFileError(filename, err)
	}

	reader := bufio.NewReaderSize(sourceFile, 1024*1024)
	writer := bufio.NewWriterSize(tempFile, 1024*1024)

	changedCount := 0
	for {
		rawLine, readErr := reader.ReadString('\n')
		if len(rawLine) > 0 {
			hasNewline := strings.HasSuffix(rawLine, "\n")
			line := strings.TrimSuffix(rawLine, "\n")

			newLine, changed, keep, transformErr := transformFunc(line)
			if transformErr != nil {
				cleanupTemp()
				return 0, transformErr
			}

			if changed {
				changedCount++
			}

			if !keep {
				continue
			}

			if _, writeErr := writer.WriteString(newLine); writeErr != nil {
				cleanupTemp()
				return 0, processor.mapFileError(filename, writeErr)
			}

			if hasNewline {
				if _, writeErr := writer.WriteString("\n"); writeErr != nil {
					cleanupTemp()
					return 0, processor.mapFileError(filename, writeErr)
				}
			}
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			cleanupTemp()
			return 0, processor.mapFileError(filename, readErr)
		}
	}

	if changedCount == 0 {
		cleanupTemp()
		return 0, nil
	}

	if err := writer.Flush(); err != nil {
		cleanupTemp()
		return 0, processor.mapFileError(filename, err)
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempName)
		return 0, processor.mapFileError(filename, err)
	}

	if err := os.Rename(tempName, filename); err != nil {
		_ = os.Remove(tempName)
		return 0, processor.mapFileError(filename, err)
	}

	return changedCount, nil
}

func (processor *Processor) transformLines(
	lines []string,
	transformFunc func(string) (string, bool, bool, error),
) ([]string, int, error) {
	count := 0
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		newLine, changed, keep, err := transformFunc(line)
		if err != nil {
			return nil, 0, err
		}

		if changed {
			count++
		}

		if !keep {
			continue
		}

		result = append(result, newLine)
	}

	return result, count, nil
}

func (processor *Processor) mapFileError(filename string, err error) error {
	if errors.Is(err, os.ErrPermission) {
		return displayable_errors.NewPermissionDeniedError(filename)
	}

	return err
}
