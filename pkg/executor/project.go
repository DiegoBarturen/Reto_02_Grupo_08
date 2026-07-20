package executor

import (
	"reto_02_grupo_08/pkg/parser"
)

type ProjectOperator struct {
	hijo     Operator
	columnas []parser.SelectColumn
}

func NuevoProjectOperator(hijo Operator, columnas []parser.SelectColumn) *ProjectOperator {
	return &ProjectOperator{
		hijo:     hijo,
		columnas: columnas,
	}
}

func (p *ProjectOperator) Next() (Row, error) {
	fila, err := p.hijo.Next()
	if err != nil {
		return Row{}, err
	}

	if len(p.columnas) == 1 && p.columnas[0].IsAsterisk {
		return fila, nil
	}

	return fila, nil
}

func (p *ProjectOperator) Close() error {
	return p.hijo.Close()
}
