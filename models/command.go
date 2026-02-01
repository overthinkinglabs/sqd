package models

import (
	"path/filepath"
	"regexp"
)

type Command struct {
	Action          Action
	File            string
	Pattern         *regexp.Regexp
	Replace         string
	MatchExact      bool
	Replacements    []Replacement
	Deletions       []Deletion
	IsBatch         bool
	SelectTarget    Select
	WhereTarget     WhereTarget
	WherePattern    *regexp.Regexp
	WhereConditions []WhereCondition
	WhereOperation  WhereOperation
}

type Replacement struct {
	Pattern    *regexp.Regexp
	Replace    string
	MatchExact bool
}

type Deletion struct {
	Pattern    *regexp.Regexp
	MatchExact bool
}

func (whereCondition *WhereCondition) matches(line string, fileName string) bool {
	if whereCondition.Target == WHERE_NAME {
		return whereCondition.Negate != whereCondition.Pattern.MatchString(filepath.Base(fileName))
	}

	return whereCondition.Negate != whereCondition.Pattern.MatchString(line)
}

func (command *Command) MatchesLine(line string, fileName string) bool {
	if len(command.WhereConditions) == 0 {
		if command.Pattern != nil {
			return command.Pattern.MatchString(line)
		}

		return true
	}

	if command.WhereOperation == WHERE_AND {
		for _, whereCondition := range command.WhereConditions {
			if !whereCondition.matches(line, fileName) {
				return false
			}
		}

		return true
	}

	for _, whereCondition := range command.WhereConditions {
		if whereCondition.matches(line, fileName) {
			return true
		}
	}

	return false
}
