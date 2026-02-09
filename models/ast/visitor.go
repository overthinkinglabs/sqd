package ast

import "github.com/albertoboccolini/sqd/models"

type Visitor interface {
	VisitSelect(statement *Select) (models.Command, error)
	VisitUpdate(statement *UpdateStatement) (models.Command, error)
	VisitDelete(statement *Delete) (models.Command, error)
}
