package dry_mode

import (
	"errors"
	"os"
	"strings"

	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
)

type FileReader struct {
	utils *services.Utils
}

func NewFileReader(utils *services.Utils) *FileReader {
	return &FileReader{utils: utils}
}

// It checks path safety and reads the file for dry run preview. Unlike Processor.ProcessFile, it does NOT require write
// permissions dry mode is read-only and should work on read-only files.
func (fileReader *FileReader) ValidateAndReadFile(file string) ([]string, error) {
	if !fileReader.utils.IsPathInsideCwd(file) {
		return nil, displayable_errors.NewInvalidPathError(file)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fileReader.mapReadError(file, err)
	}

	return strings.Split(string(data), "\n"), nil
}

func (fileReader *FileReader) mapReadError(file string, err error) error {
	if errors.Is(err, os.ErrPermission) {
		return displayable_errors.NewPermissionDeniedError(file)
	}

	return err
}
