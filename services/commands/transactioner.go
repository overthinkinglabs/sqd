package commands

import (
	"fmt"
	"os"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
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

func (transactioner *Transactioner) ExecuteUpdateTransaction(files []string,
	updateFunc func(string) (int, error), stats *models.ExecutionStats) int {

	transactioner.checkFilesBeforeTransaction(files)
	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath := file + ".sqd_backup"
		if err := os.Rename(file, backupPath); err != nil {
			transactioner.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return 0
		}
		backups = append(backups, fileBackup{original: file, backup: backupPath})

		count, err := updateFunc(backupPath)
		if err != nil {
			transactioner.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return 0
		}

		if err := os.Rename(backupPath, file); err != nil {
			transactioner.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return 0
		}

		total += count
		stats.Processed++
	}

	return total
}

func (transactioner *Transactioner) checkFilesBeforeTransaction(files []string) {
	for _, file := range files {
		if !transactioner.utils.IsPathInsideCwd(file) {
			fmt.Fprintf(os.Stderr, "Transaction failed: invalid path %s\n", file)
			os.Exit(1)
		}
		if !transactioner.utils.CanWriteFile(file) {
			fmt.Fprintf(os.Stderr, "Transaction failed: cannot write %s\n", file)
			os.Exit(1)
		}
	}
}

func (transactioner *Transactioner) rollbackFiles(backups []fileBackup) {
	for _, backup := range backups {
		if err := os.Rename(backup.backup, backup.original); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed for %s -> %s: %v\n", backup.backup, backup.original, err)
		}
	}
}

func (transactioner *Transactioner) ExecuteDeleteTransaction(files []string,
	deleteFunc func(string) (int, error), stats *models.ExecutionStats) int {
	transactioner.checkFilesBeforeTransaction(files)
	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath := file + ".sqd_backup"
		if err := os.Rename(file, backupPath); err != nil {
			transactioner.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return 0
		}
		backups = append(backups, fileBackup{original: file, backup: backupPath})

		count, err := deleteFunc(backupPath)
		if err != nil {
			transactioner.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return 0
		}

		if err := os.Rename(backupPath, file); err != nil {
			transactioner.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return 0
		}

		total += count
		stats.Processed++
	}

	return total
}
