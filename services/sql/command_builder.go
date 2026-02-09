package sql

import (
	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/models/ast"
)

type CommandBuilder struct{}

func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{}
}

func (commandBuilder *CommandBuilder) VisitSelect(statement *ast.Select) (models.Command, error) {
	command := models.Command{
		Action:       models.SELECT,
		SelectTarget: statement.Target,
		File:         statement.Source,
		OrderBy:      statement.OrderBy,
		WhereTarget:  models.CONTENT,
	}

	if statement.IsCount {
		command.Action = models.COUNT
	}

	if statement.WhereClause != nil {
		command.WhereTarget = statement.WhereClause.Target

		if statement.WhereClause.Target == models.CONTENT {
			command.Pattern = statement.WhereClause.Pattern
			command.NegateContent = statement.WhereClause.Negate
		}

		if statement.WhereClause.Target == models.NAME {
			command.WherePattern = statement.WhereClause.Pattern
			command.NegateFileName = statement.WhereClause.Negate
		}
	}

	return command, nil
}

func (commandBuilder *CommandBuilder) VisitUpdate(statement *ast.UpdateStatement) (models.Command, error) {
	command := models.Command{
		Action:       models.UPDATE,
		File:         statement.Source,
		IsBatch:      statement.IsBatch,
		Replacements: statement.Replacements,
		WhereTarget:  models.CONTENT,
	}

	if statement.WhereClause != nil && !statement.IsBatch {
		command.WhereTarget = statement.WhereClause.Target

		if statement.WhereClause.Target == models.CONTENT {
			command.Pattern = statement.WhereClause.Pattern
			command.NegateContent = statement.WhereClause.Negate

			if len(statement.Replacements) > 0 {
				command.Replace = statement.Replacements[0].Replace
			}
		}

		if statement.WhereClause.Target == models.NAME {
			command.WherePattern = statement.WhereClause.Pattern
			command.NegateFileName = statement.WhereClause.Negate
		}
	}

	return command, nil
}

func (commandBuilder *CommandBuilder) VisitDelete(statement *ast.Delete) (models.Command, error) {
	command := models.Command{
		Action:      models.DELETE,
		File:        statement.Source,
		IsBatch:     statement.IsBatch,
		Deletions:   statement.Deletions,
		WhereTarget: models.CONTENT,
	}

	if statement.WhereClause != nil && !statement.IsBatch {
		command.WhereTarget = statement.WhereClause.Target

		if statement.WhereClause.Target == models.CONTENT {
			command.Pattern = statement.WhereClause.Pattern
			command.NegateContent = statement.WhereClause.Negate
		}

		if statement.WhereClause.Target == models.NAME {
			command.WherePattern = statement.WhereClause.Pattern
			command.NegateFileName = statement.WhereClause.Negate
		}
	}

	return command, nil
}
