package executor

import (
	"errors"
	"fmt"
	"reto_02_grupo_08/pkg/parser"
)

func Evaluar(fila Row, expr parser.Expr) (any, error) {
	switch e := expr.(type) {
	case *parser.Literal:
		return e.Value, nil

	case *parser.Identifier:
		return nil, fmt.Errorf("lectura de la columna '%s' pendiente", e.Name)

	case *parser.BinaryExpr:
		izq, err := Evaluar(fila, e.Left)
		if err != nil {
			return nil, err
		}
		der, err := Evaluar(fila, e.Right)
		if err != nil {
			return nil, err
		}

		return resolverOperacion(izq, der, e.Operator)

	default:
		return nil, errors.New("tipo de expresión no soportada")
	}
}

func resolverOperacion(izq, der any, operador string) (bool, error) {
	switch operador {
	case "=":
		return izq == der, nil
	case "<>":
		return izq != der, nil
	case "AND":
		izqBool, ok1 := izq.(bool)
		derBool, ok2 := der.(bool)
		if !ok1 || !ok2 {
			return false, errors.New("AND requiere que ambos lados sean booleanos")
		}
		return izqBool && derBool, nil
	case "OR":
		izqBool, ok1 := izq.(bool)
		derBool, ok2 := der.(bool)
		if !ok1 || !ok2 {
			return false, errors.New("OR requiere que ambos lados sean booleanos")
		}
		return izqBool || derBool, nil
	}

	return false, fmt.Errorf("operador matemático '%s' pendiente de implementar", operador)
}
