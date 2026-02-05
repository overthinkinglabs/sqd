package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"bufio"

	"github.com/albertoboccolini/sqd/models"
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

func executeQuery(query string, useTransaction, dryRun bool, showFileNames bool) {
	validator := sql.NewValidator()
	if err := validator.Validate(query); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	extractor := sql.NewExtractor()
	batchParser := sql.NewBatchParser(extractor)
	parser := sql.NewParser(extractor, batchParser)
	command := parser.Parse(query)

	utils := services.NewUtils()
	finder := files.NewFinder()
	processor := files.NewProcessor(utils)
	parallelizer := files.NewParallelizer(utils)

	foundFiles := finder.FindFiles(command.File)
	if len(foundFiles) == 0 {
		fmt.Println("No files found")
		return
	}

	dryModeErrorHandler := dry_mode.NewErrorHandler()
	dryModeFileReader := dry_mode.NewFileReader(dryModeErrorHandler, utils)
	dryModeChangeDisplayer := dry_mode.NewChangeDisplayer(dryModeFileReader)
	dryModeChangeCounter := dry_mode.NewChangeCounter(dryModeFileReader)
	dryModeRunner := dry_mode.NewRunner(dryModeChangeDisplayer, dryModeChangeCounter, dryModeFileReader, dryModeErrorHandler, utils)

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

	dispatcher.Execute(command, foundFiles, useTransaction, dryRun, showFileNames)
}

func executeQueriesFromFile(filePath string, useTransaction, dryRun bool, showFileNames bool) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error: Unable to open file %s: %v\n", filePath, err)
		os.Exit(1)
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
		executeQuery(query, useTransaction, dryRun, showFileNames)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error: Failed to read queries from file %s: %v\n", filePath, err)
		os.Exit(1)
	}
}

func handleDryCommand(args []string) {
	dryFlagSet := flag.NewFlagSet("dry", flag.ExitOnError)
	completeFlag := dryFlagSet.Bool("complete", false, "Show file names with modified lines")
	dryFlagSet.BoolVar(completeFlag, "c", false, "Show file names with modified lines")
	transactionFlag := dryFlagSet.Bool("transaction", false, "Enable transaction mode with rollback on failure")
	dryFlagSet.BoolVar(transactionFlag, "t", false, "Enable transaction mode with rollback on failure")
	queryFile := dryFlagSet.String("file", "", "Path to a file containing queries to execute")
	dryFlagSet.StringVar(queryFile, "f", "", "Path to a file containing queries to execute")
	dryFlagSet.Parse(args)

	if *queryFile != "" {
		executeQueriesFromFile(*queryFile, *transactionFlag, true, *completeFlag)
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
	executeQuery(query, *transactionFlag, true, *completeFlag)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "dry" {
		handleDryCommand(os.Args[2:])
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
		executeQueriesFromFile(*queryFile, *transactionFlag, false, false)
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
	executeQuery(sql, *transactionFlag, false, false)
}
