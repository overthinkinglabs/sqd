package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/commands"
	"github.com/albertoboccolini/sqd/services/files"
)

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	flag.BoolVar(versionFlag, "v", false, "Print version information")
	transactionFlag := flag.Bool("transaction", false, "Enable transaction mode with rollback on failure")
	flag.BoolVar(transactionFlag, "t", false, "Enable transaction mode with rollback on failure")
	dryRunFlag := flag.Bool("dry-run", false, "Show what would be done without making changes")
	flag.BoolVar(dryRunFlag, "d", false, "Show what would be done without making changes")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("v%s\n", models.VERSION)
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Usage: sqd 'query'")
		fmt.Println("\nCommands:")
		fmt.Println("  SELECT - Display matching lines")
		fmt.Println("  UPDATE - Replace content in matching lines")
		fmt.Println("  DELETE - Remove matching lines")
		fmt.Println("\nExamples:")
		fmt.Println("  sqd 'SELECT * FROM file.txt WHERE content LIKE pattern'")
		fmt.Println("  sqd 'UPDATE file.txt SET old TO new WHERE content = match, SET foo TO bar WHERE content = other'")
		fmt.Println("  sqd 'DELETE FROM file.txt WHERE content = exact_match'")
		fmt.Println("\nFlags:")
		fmt.Println("  -d, --dry-run\t\tShow what would be done without making changes")
		fmt.Println("  -t, --transaction	Enable transaction mode with rollback on failure")
		fmt.Println("  -v, --version		Show the version information")
		os.Exit(1)
	}

	sql := strings.Join(flag.Args(), " ")

	sqlParser := services.NewSQLParser()
	if err := sqlParser.Validate(sql); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	command := sqlParser.Parse(sql)

	utils := services.NewUtils()
	finder := files.NewFinder()
	processor := files.NewProcessor(utils)
	parallelizer := files.NewParallelizer(utils)

	foundFiles := finder.FindFiles(command.File)
	if len(foundFiles) == 0 {
		fmt.Println("No files found")
		os.Exit(1)
	}

	dryRunner := commands.NewDryRunner(utils)
	transactioner := commands.NewTransactioner(utils)
	searcher := commands.NewSearcher(parallelizer, utils)
	updater := commands.NewUpdater(processor, utils)
	deleter := commands.NewDeleter(processor, utils)
	dispatcher := commands.NewDispatcher(
		searcher,
		updater,
		deleter,
		transactioner,
		dryRunner,
		utils,
		parallelizer,
	)

	dispatcher.Execute(command, foundFiles, *transactionFlag, *dryRunFlag)
}
