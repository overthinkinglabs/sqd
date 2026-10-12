package files

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
)

type Processor struct {
	utils *services.Utils
}

func NewProcessor(utils *services.Utils) *Processor {
	return &Processor{utils: utils}
}

func (processor *Processor) ProcessFile(filename string, transformFunc func([]string) ([]string, int)) (int, error) {
	if !processor.utils.IsPathInsideCwd(filename) {
		return 0, displayable_errors.NewInvalidPathError(filename)
	}
	if !processor.utils.CanWriteFile(filename) {
		return 0, displayable_errors.NewPermissionDeniedError(filename)
	}

	// Open the file once and operate on the same descriptor to avoid TOCTOU races where the path could be swapped for a
	// symlink between check and write.
	file, err := os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		return 0, processor.mapFileError(filename, err)
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	if info.Size() > models.MAX_FILE_SIZE {
		return 0, displayable_errors.NewFileTooLargeError(filename, info.Size(), models.MAX_FILE_SIZE)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	lines := strings.Split(string(data), "\n")
	newLines, count := transformFunc(lines)

	if count == 0 {
		return 0, nil
	}

	if err := file.Truncate(0); err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	if _, err := file.WriteString(strings.Join(newLines, "\n")); err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	return count, nil
}

func (processor *Processor) mapFileError(filename string, err error) error {
	if errors.Is(err, os.ErrPermission) {
		return displayable_errors.NewPermissionDeniedError(filename)
	}

	return err
}
