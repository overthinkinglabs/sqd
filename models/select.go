package models

type Select string

const (
	ALL     Select = "*"
	NAME    Select = "name"
	CONTENT Select = "content"
)
