# sqd | A SQL-like document editor

Traditional Unix tools (grep, sed, awk) are powerful but have inconsistent syntax and steep learning curves. `sqd` (pronounced like squad) provides a familiar SQL interface for common text operations.

## Getting Started

This project requires **Go version 1.25.4 or higher**. Make sure you have a compatible version installed. If needed, download the latest version from [https://go.dev/dl/](https://go.dev/dl/)

1. **Installation**: Installs sqd in the system

    ```bash
    go install github.com/albertoboccolini/sqd@latest
    ```

2. **Start using sqd**: See all the open todos in your markdown files

    ```bash
    sqd 'SELECT * FROM *.md WHERE content LIKE "%- [ ]%"'
    ```

## Useful Commands

Count all the LaTeX formulas in your notes

```bash
sqd 'SELECT COUNT(*) FROM *.md WHERE content LIKE "$$%"'
```

Refactor your markdown title hierarchy

```bash
sqd 'UPDATE *.md SET content="### " WHERE content LIKE "## %"'
```

Remove all DEBUG logs

```bash
sqd 'DELETE FROM *.log WHERE content LIKE "%DEBUG%"'
```

## Columns

You can reference the following columns in the `SELECT` clause:

* `name`: the file name.
* `content`: the content of each line.
* `*`: both file name and content.

Examples:

```bash
sqd 'SELECT COUNT(name) FROM *.md WHERE content LIKE "### %"' 
# Counts the number of files that contain at least one matching line.

sqd 'SELECT COUNT(content) FROM *.md WHERE content LIKE "### %"' 
# Counts the total number of matching lines across all files.
```

## Flags

* `-d`, `--dry-run`: Display the actions that would be performed without modifying any files.
* `-t`, `--transaction`: Apply changes atomically. If any operation fails, all changes are rolled back.

## The power of sqd

Let's suppose we have a file with multiple similar titles, but we only want to change specific ones. With sed or awk, we need complex regex or multiple commands. With sqd, we can target exact lines and batch multiple replacements in a single command.

```markdown
## Title 1 to be updated

## Title 1 not to be updated

## Title 1 TO be updated

## Title 2 to be updated

## Title 2 not to be updated

## Title 2 TO be updated
```

With only one sqd command

```bash
sqd 'UPDATE example.md 
SET content="## Title 1 UPDATED" WHERE content = "## Title 1 to be updated",
SET content="## Title 2 UPDATED" WHERE content = "## Title 2 TO be updated"'
```

You will obtain the following result

```markdown
## Title 1 UPDATED

## Title 1 not to be updated

## Title 1 TO be updated

## Title 2 to be updated

## Title 2 not to be updated

## Title 2 UPDATED
```
