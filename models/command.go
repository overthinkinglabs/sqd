package models

import (
	"regexp"
)

type Command struct {
	Action       TokenType
	File         string
	Pattern      *regexp.Regexp
	Replace      string
	Replacements []Replacement
	Deletions    []Deletion
	IsBatch      bool
	SelectTarget TokenType
	WhereTarget  TokenType
	WherePattern *regexp.Regexp
	OrderBy      []OrderBy
}

type OrderBy struct {
	Column    TokenType
	Direction TokenType
}

type Replacement struct {
	Pattern *regexp.Regexp
	Replace string
}

type Deletion struct {
	Pattern *regexp.Regexp
}
