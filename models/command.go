package models

import (
	"regexp"
)

type Command struct {
	Action       Action
	File         string
	Pattern      *regexp.Regexp
	Replace      string
	MatchExact   bool
	Replacements []Replacement
	Deletions    []Deletion
	IsBatch      bool
	SelectTarget Select
	WhereTarget  WhereTarget
	WherePattern *regexp.Regexp
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

type WhereTarget int

const (
	WHERE_CONTENT WhereTarget = iota
	WHERE_NAME
)
