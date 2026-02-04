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

func executeQuery(query string, useTransaction, dryRun bool) {
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

	dryRunner := commands.NewDryRunner(utils)
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
		dryRunner,
		utils,
		parallelizer,
	)

	dispatcher.Execute(command, foundFiles, useTransaction, dryRun)
}

func executeQueriesFromFile(filePath string, useTransaction, dryRun bool) {
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
		executeQuery(query, useTransaction, dryRun)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error: Failed to read queries from file %s: %v\n", filePath, err)
		os.Exit(1)
	}
}

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	flag.BoolVar(versionFlag, "v", false, "Print version information")
	transactionFlag := flag.Bool("transaction", false, "Enable transaction mode with rollback on failure")
	flag.BoolVar(transactionFlag, "t", false, "Enable transaction mode with rollback on failure")
	dryRunFlag := flag.Bool("dry-run", false, "Show what would be done without making changes")
	flag.BoolVar(dryRunFlag, "d", false, "Show what would be done without making changes")
	queryFile := flag.String("file", "", "Path to a file containing queries to execute")
	flag.StringVar(queryFile, "f", "", "Path to a file containing queries to execute")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("v%s\n", models.VERSION)
		os.Exit(0)
	}

	if *queryFile != "" {
		executeQueriesFromFile(*queryFile, *transactionFlag, *dryRunFlag)
		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println("sqd | A SQL-like document editor")
		fmt.Println("\nUsage: sqd [flags] 'query'")
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
		fmt.Println("  LIKE		Pattern match with wildcards (%)")
		fmt.Println("\nExamples:")
		fmt.Println("  sqd 'SELECT * | name | content FROM file.txt WHERE content LIKE pattern ORDER BY name | content ASC | DESC'")
		fmt.Println("  sqd -d 'UPDATE file.txt SET old TO new WHERE content = match, SET foo TO bar WHERE content = other'")
		fmt.Println("  sqd -t 'DELETE FROM file.txt WHERE content = exact_match'")
		fmt.Println("  sqd -f path/to/file")
		fmt.Println("\nFlags:")
		fmt.Println("  -f, --file        Path to a file containing queries to execute")
		fmt.Println("  -d, --dry-run     Show what would be done without making changes")
		fmt.Println("  -t, --transaction Enable transaction mode with rollback on failure")
		fmt.Println("  -v, --version     Show the version information")
		os.Exit(1)
	}

	sql := strings.Join(flag.Args(), " ")
	executeQuery(sql, *transactionFlag, *dryRunFlag)
}
