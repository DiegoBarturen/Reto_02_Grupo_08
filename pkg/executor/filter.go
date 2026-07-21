package executor

import (
	"reto_02_grupo_08/pkg/almacenamiento"
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
	esquema := f.hijo.Schema()
	for {
		fila, err := f.hijo.Next()
		if err != nil {
			return Row{}, err
		}

		if evaluarCondicion(fila, esquema, f.condicion) {
			return fila, nil
		}
	}
}

// Schema delega al hijo: un filtro no cambia el esquema de sus filas.
func (f *FilterOperator) Schema() almacenamiento.Esquema {
	return f.hijo.Schema()
}

func (f *FilterOperator) Close() error {
	return f.hijo.Close()
}

// evaluarCondicion evalúa una condición booleana sobre una fila.
// Retorna false ante cualquier error o valor no booleano (incluyendo NULL).
func evaluarCondicion(fila Row, esquema almacenamiento.Esquema, condicion parser.Expr) bool {
	resultado, err := Evaluar(fila, esquema, condicion)
	if err != nil {
		return false
	}
	if booleano, ok := resultado.(bool); ok {
		return booleano
	}
	return false
}
