package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
)

type Transactioner struct {
	utils *services.Utils
}

func NewTransactioner(utils *services.Utils) *Transactioner {
	return &Transactioner{utils: utils}
}

type fileBackup struct {
	original string
	backup   string
}

func (transactioner *Transactioner) Update(files []string,
	updateFunc func(string) (int, error), stats *models.ExecutionStats,
) (int, error) {
	return transactioner.execute(files, updateFunc, stats)
}

func (transactioner *Transactioner) Delete(files []string,
	deleteFunc func(string) (int, error), stats *models.ExecutionStats,
) (int, error) {
	return transactioner.execute(files, deleteFunc, stats)
}

func (transactioner *Transactioner) execute(files []string,
	mutate func(string) (int, error), stats *models.ExecutionStats,
) (int, error) {
	if err := transactioner.checkFilesBeforeTransaction(files); err != nil {
		return 0, err
	}

	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath, err := transactioner.createUniqueBackup(file)
		if err != nil {
			transactioner.rollbackFiles(backups)
			return 0, displayable_errors.NewTransactionFailedError(err.Error())
		}

		backups = append(backups, fileBackup{original: file, backup: backupPath})

		count, err := mutate(backupPath)
		if err != nil {
			transactioner.rollbackFiles(backups)
			return 0, displayable_errors.NewTransactionFailedError(err.Error())
		}

		if err := os.Rename(backupPath, file); err != nil {
			transactioner.rollbackFiles(backups)
			return 0, displayable_errors.NewTransactionFailedError(err.Error())
		}

		total += count
		stats.Processed++
	}

	return total, nil
}

// It renames the original file to a uniquely named backup in the same directory. Using os.CreateTemp avoids silently
// overwriting a pre-existing .sqd_backup file from a previous crashed run.
func (transactioner *Transactioner) createUniqueBackup(file string) (string, error) {
	dir := filepath.Dir(file)
	tempFile, createErr := os.CreateTemp(dir, ".sqd_backup_*")
	if createErr != nil {
		return "", createErr
	}

	tempName := tempFile.Name()
	if closeErr := tempFile.Close(); closeErr != nil {
		_ = os.Remove(tempName)
		return "", closeErr
	}
	if removeErr := os.Remove(tempName); removeErr != nil {
		return "", removeErr
	}

	if renameErr := os.Rename(file, tempName); renameErr != nil {
		return "", renameErr
	}

	return tempName, nil
}

func (transactioner *Transactioner) checkFilesBeforeTransaction(files []string) error {
	for _, file := range files {
		if !transactioner.utils.IsPathInsideCwd(file) {
			return displayable_errors.NewTransactionFailedError("invalid path " + file)
		}

		if !transactioner.utils.CanWriteFile(file) {
			return displayable_errors.NewTransactionFailedError("cannot write " + file)
		}
	}

	return nil
}

func (transactioner *Transactioner) rollbackFiles(backups []fileBackup) {
	for _, backup := range backups {
		if err := os.Rename(backup.backup, backup.original); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed for %s -> %s: %v\n", backup.backup, backup.original, err)
		}
	}
}
