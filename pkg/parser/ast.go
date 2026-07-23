package parser

import (
	"fmt"
	"strings"
)

type Node interface {
	String() string
}

type Expr interface {
	Node
	exprNode()
}

// SelectStmt representa una consulta SELECT completa.
type SelectStmt struct {
	Columns []SelectColumn
	Table   string
	Join    *JoinClause
	Where   Expr
	GroupBy []string // nombres de columnas del GROUP BY
	OrderBy []OrderByItem
	Limit   *int64 // nil si no hay LIMIT
}

// JoinClause representa la cláusula INNER JOIN ... ON.
type JoinClause struct {
	Table string
	On    Expr
}

func (s *SelectStmt) String() string {
	if s == nil {
		return ""
	}

	columnas := make([]string, len(s.Columns))
	for indice, columna := range s.Columns {
		columnas[indice] = columna.String()
	}

	consulta := fmt.Sprintf(
		"SELECT %s FROM %s",
		strings.Join(columnas, ", "),
		s.Table,
	)

	if s.Join != nil {
		consulta += fmt.Sprintf(" INNER JOIN %s ON %s", s.Join.Table, s.Join.On.String())
	}

	if s.Where != nil {
		consulta += " WHERE " + s.Where.String()
	}

	if len(s.GroupBy) > 0 {
		consulta += " GROUP BY " + strings.Join(s.GroupBy, ", ")
	}

	if len(s.OrderBy) > 0 {
		items := make([]string, len(s.OrderBy))
		for i, o := range s.OrderBy {
			dir := "ASC"
			if o.Desc {
				dir = "DESC"
			}
			items[i] = o.Column + " " + dir
		}
		consulta += " ORDER BY " + strings.Join(items, ", ")
	}

	if s.Limit != nil {
		consulta += fmt.Sprintf(" LIMIT %d", *s.Limit)
	}

	return consulta
}

// OrderByItem representa un criterio de ordenamiento (columna + dirección).
type OrderByItem struct {
	Column string
	Desc   bool // true = DESC, false = ASC
}

// SelectColumn representa una columna en la cláusula SELECT.
// Puede ser *, un nombre de columna o una función de agregación.
type SelectColumn struct {
	Name       string
	IsAsterisk bool
	Agg        *AggFunc // no nil si es una función de agregación
}

func (c SelectColumn) String() string {
	if c.IsAsterisk {
		return "*"
	}
	if c.Agg != nil {
		return c.Agg.String()
	}
	return c.Name
}

// AggFunc representa una función de agregación: COUNT(*), SUM(col), AVG(col), MIN(col), MAX(col).
type AggFunc struct {
	Name   string // COUNT, SUM, AVG, MIN, MAX
	Column string // nombre de columna (vacío para COUNT(*))
	IsStar bool   // true para COUNT(*)
}

func (a *AggFunc) exprNode() {}

func (a *AggFunc) String() string {
	if a == nil {
		return ""
	}
	if a.IsStar {
		return fmt.Sprintf("%s(*)", a.Name)
	}
	return fmt.Sprintf("%s(%s)", a.Name, a.Column)
}

type Identifier struct {
	Name string
}

func (i *Identifier) exprNode() {}

func (i *Identifier) String() string {
	if i == nil {
		return ""
	}

	return i.Name
}

type Literal struct {
	Value any
	Raw   string
}

func (l *Literal) exprNode() {}

func (l *Literal) String() string {
	if l == nil {
		return ""
	}

	if l.Raw != "" {
		return l.Raw
	}

	return fmt.Sprint(l.Value)
}

type BinaryExpr struct {
	Left     Expr
	Operator string
	Right    Expr
}

func (b *BinaryExpr) exprNode() {}

func (b *BinaryExpr) String() string {
	if b == nil {
		return ""
	}

	return fmt.Sprintf(
		"(%s %s %s)",
		b.Left.String(),
		b.Operator,
		b.Right.String(),
	)
}
