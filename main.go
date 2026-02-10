package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"bufio"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/models/displayable_errors"
	"github.com/albertoboccolini/sqd/services"
	"github.com/albertoboccolini/sqd/services/commands"
	"github.com/albertoboccolini/sqd/services/dry_mode"
	"github.com/albertoboccolini/sqd/services/files"
	"github.com/albertoboccolini/sqd/services/sql"
)

func splitQueries(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i := range data {
		if data[i] == ';' {
			return i + 1, data[:i], nil
		}
	}

	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func handleError(errorHandler *services.ErrorHandler, err error) {
	errorHandler.HandleError(err)
	os.Exit(1)
}

func executeQuery(query string, useTransaction, dryRun bool, showDetailedOutputInDryMode bool) error {
	validator := sql.NewValidator()
	if err := validator.Validate(query); err != nil {
		return err
	}

	extractor := sql.NewExtractor()
	batchParser := sql.NewBatchParser(extractor)
	commandBuilder := sql.NewCommandBuilder()
	parser := sql.NewParser(extractor, batchParser, commandBuilder)
	command, err := parser.Parse(query)
	if err != nil {
		return err
	}

	utils := services.NewUtils()
	finder := files.NewFinder()
	processor := files.NewProcessor(utils)
	parallelizer := files.NewParallelizer(utils)

	foundFiles, err := finder.FindFiles(command.File)
	if err != nil {
		return err
	}

	if len(foundFiles) == 0 {
		return displayable_errors.NewNoFilesFoundError(command.File)
	}

	dryModeFileReader := dry_mode.NewFileReader(utils)
	dryModeChangeProcessor := dry_mode.NewChangeProcessor(dryModeFileReader, utils)
	dryModeRunner := dry_mode.NewRunner(dryModeChangeProcessor, utils)

	transactioner := commands.NewTransactioner(utils)
	sorter := commands.NewSorter()
	searcher := commands.NewSearcher(parallelizer, sorter, utils)
	counter := commands.NewCounter(parallelizer, searcher)
	updater := commands.NewUpdater(processor, utils)
	deleter := commands.NewDeleter(processor, utils)
	dispatcher := commands.NewDispatcher(
		searcher,
		counter,
		updater,
		deleter,
		transactioner,
		dryModeRunner,
		utils,
		parallelizer,
	)

	return dispatcher.Execute(command, foundFiles, useTransaction, dryRun, showDetailedOutputInDryMode)
}

func executeQueriesFromFile(filePath string, useTransaction, dryRun bool, showDetailedOutputInDryMode bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return displayable_errors.NewFileReadError(filePath, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(splitQueries)

	for scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}

		fmt.Printf("%s\n", query)
		if err := executeQuery(query, useTransaction, dryRun, showDetailedOutputInDryMode); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return displayable_errors.NewFileReadError(filePath, err)
	}

	return nil
}

func handleDryModeCommand(args []string, errorHandler *services.ErrorHandler) {
	dryFlagSet := flag.NewFlagSet("dry", flag.ExitOnError)
	completeFlag := dryFlagSet.Bool("complete", false, "Show file names with modified lines")
	dryFlagSet.BoolVar(completeFlag, "c", false, "Show file names with modified lines")
	transactionFlag := dryFlagSet.Bool("transaction", false, "Enable transaction mode with rollback on failure")
	dryFlagSet.BoolVar(transactionFlag, "t", false, "Enable transaction mode with rollback on failure")
	queryFile := dryFlagSet.String("file", "", "Path to a file containing queries to execute")
	dryFlagSet.StringVar(queryFile, "f", "", "Path to a file containing queries to execute")
	dryFlagSet.Parse(args)

	if *queryFile != "" {
		if err := executeQueriesFromFile(*queryFile, *transactionFlag, true, *completeFlag); err != nil {
			handleError(errorHandler, err)
		}

		return
	}

	if len(dryFlagSet.Args()) == 0 {
		fmt.Println("Usage: sqd dry [flags] 'query'")
		fmt.Println("\nFlags:")
		fmt.Println("  -c, --complete    Show file names with modified lines")
		fmt.Println("  -t, --transaction Enable transaction mode with rollback on failure")
		fmt.Println("  -f, --file        Path to a file containing queries to execute")
		os.Exit(1)
	}

	query := strings.Join(dryFlagSet.Args(), " ")

	if err := executeQuery(query, *transactionFlag, true, *completeFlag); err != nil {
		handleError(errorHandler, err)
	}
}

func main() {
	errorHandler := services.NewErrorHandler()
	if len(os.Args) > 1 && os.Args[1] == "dry" {
		handleDryModeCommand(os.Args[2:], errorHandler)
		return
	}

	versionFlag := flag.Bool("version", false, "Print version information")
	flag.BoolVar(versionFlag, "v", false, "Print version information")
	transactionFlag := flag.Bool("transaction", false, "Enable transaction mode with rollback on failure")
	flag.BoolVar(transactionFlag, "t", false, "Enable transaction mode with rollback on failure")
	queryFile := flag.String("file", "", "Path to a file containing queries to execute")
	flag.StringVar(queryFile, "f", "", "Path to a file containing queries to execute")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("v%s\n", models.VERSION)
		os.Exit(0)
	}

	if *queryFile != "" {
		if err := executeQueriesFromFile(*queryFile, *transactionFlag, false, false); err != nil {
			handleError(errorHandler, err)
		}

		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println("sqd | A SQL-like document editor")
		fmt.Println("\nUsage: sqd [flags] 'query'")
		fmt.Println("       sqd dry [flags] 'query'")
		fmt.Println("\nStatements:")
		fmt.Println("  SELECT	Display matching lines")
		fmt.Println("  UPDATE	Replace content in matching lines")
		fmt.Println("  DELETE	Remove matching lines")
		fmt.Println("  COUNT		Count matching lines (using *, name, or content)")
		fmt.Println("\nClauses:")
		fmt.Println("  FROM		Specify the target file or file pattern")
		fmt.Println("  WHERE		Define matching criteria")
		fmt.Println("  SET		Define replacement content for UPDATE statements (only for content)")
		fmt.Println("  ORDER BY 	Sort matching lines (using name or content)")
		fmt.Println("\nOperators:")
		fmt.Println("  =		Exact match")
		fmt.Println("  !=		Negation of exact match")
		fmt.Println("  LIKE		Pattern match with wildcards (%)")
		fmt.Println("\nExamples:")
		fmt.Println("  sqd 'SELECT * | name | content FROM file.txt WHERE content LIKE pattern ORDER BY name | content ASC | DESC'")
		fmt.Println("  sqd dry 'UPDATE file.txt SET old TO new WHERE content = match, SET foo TO bar WHERE content = other'")
		fmt.Println("  sqd dry -c 'UPDATE file.txt SET old TO new WHERE content = match'")
		fmt.Println("  sqd -t 'DELETE FROM file.txt WHERE content = exact_match'")
		fmt.Println("  sqd -f path/to/file")
		fmt.Println("\nCommands:")
		fmt.Println("  dry               Show what would be done without making changes")
		fmt.Println("    -c, --complete  Show file names with modified lines")
		fmt.Println("\nFlags:")
		fmt.Println("  -f, --file        Path to a file containing queries to execute")
		fmt.Println("  -t, --transaction Enable transaction mode with rollback on failure")
		fmt.Println("  -v, --version     Show the version information")
		os.Exit(1)
	}

	sql := strings.Join(flag.Args(), " ")

	if err := executeQuery(sql, *transactionFlag, false, false); err != nil {
		handleError(errorHandler, err)
	}
}
