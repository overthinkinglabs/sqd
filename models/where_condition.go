package models

import "regexp"

type WhereCondition struct {
	Target  WhereTarget
	Pattern *regexp.Regexp
	Negate  bool
}
