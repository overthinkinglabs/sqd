package models

type WhereOperation int

const (
	WHERE_SINGLE WhereOperation = iota
	WHERE_AND
	WHERE_OR
)
