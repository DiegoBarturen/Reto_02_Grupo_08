package executor

import (
	"reto_02_grupo_08/pkg/parser"
)

type FilterOperator struct {
	hijo      Operator
	condicion parser.Expr
}

func NuevoFilterOperator(hijo Operator, condicion parser.Expr) *FilterOperator {
	return &FilterOperator{
		hijo:      hijo,
		condicion: condicion,
	}
}

func (f *FilterOperator) Next() (Row, error) {
	for {
		fila, err := f.hijo.Next()
		if err != nil {
			return Row{}, err
		}

		pasaFiltro := evaluarCondicion(fila, f.condicion)

		if pasaFiltro {
			return fila, nil
		}

	}
}

func (f *FilterOperator) Close() error {
	return f.hijo.Close()
}

func evaluarCondicion(fila Row, condicion parser.Expr) bool {
	resultado, err := Evaluar(fila, condicion)
	if err != nil {
		return false
	}

	if booleano, ok := resultado.(bool); ok {
		return booleano
	}
	return false
}
