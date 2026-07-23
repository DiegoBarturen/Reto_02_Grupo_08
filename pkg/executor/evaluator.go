package executor

import (
	"errors"
	"fmt"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

// Evaluar evalúa una expresión del AST sobre una fila con su esquema.
// Usa el esquema para resolver los identificadores de columna a valores reales.
// Sigue la semántica SQL para NULL: cualquier comparación con NULL devuelve false.
func Evaluar(fila Row, esquema almacenamiento.Esquema, expr parser.Expr) (any, error) {
	switch e := expr.(type) {
	case *parser.Literal:
		return e.Value, nil

	case *parser.Identifier:
		idx, err := indiceDeLaColumna(esquema, e.Name)
		if err != nil {
			return nil, err
		}
		v := fila.Datos[idx]
		if v.EsNulo {
			return nil, nil // NULL se representa como nil
		}
		return v.Dato, nil

	case *parser.BinaryExpr:
		return evaluarBinario(fila, esquema, e)

	default:
		return nil, errors.New("tipo de expresión no soportada")
	}
}

func evaluarBinario(fila Row, esquema almacenamiento.Esquema, e *parser.BinaryExpr) (any, error) {
	// AND y OR usan evaluación en cortocircuito.
	switch e.Operator {
	case "AND":
		izq, err := Evaluar(fila, esquema, e.Left)
		if err != nil {
			return nil, err
		}
		izqBool, ok := izq.(bool)
		if !ok || !izqBool {
			return false, nil // NULL o false → false
		}
		der, err := Evaluar(fila, esquema, e.Right)
		if err != nil {
			return nil, err
		}
		derBool, ok := der.(bool)
		if !ok {
			return false, nil
		}
		return derBool, nil

	case "OR":
		izq, err := Evaluar(fila, esquema, e.Left)
		if err != nil {
			return nil, err
		}
		if izqBool, ok := izq.(bool); ok && izqBool {
			return true, nil
		}
		der, err := Evaluar(fila, esquema, e.Right)
		if err != nil {
			return nil, err
		}
		if derBool, ok := der.(bool); ok && derBool {
			return true, nil
		}
		return false, nil
	}

	// Para los demás operadores, evaluar ambos lados primero.
	izq, err := Evaluar(fila, esquema, e.Left)
	if err != nil {
		return nil, err
	}
	der, err := Evaluar(fila, esquema, e.Right)
	if err != nil {
		return nil, err
	}

	return resolverOperacion(izq, der, e.Operator)
}

// resolverOperacion aplica un operador binario a dos valores ya evaluados.
// NULL en cualquier operando siempre produce false (semántica SQL estándar).
func resolverOperacion(izq, der any, operador string) (bool, error) {
	if izq == nil || der == nil {
		return false, nil
	}

	switch operador {
	case "=":
		return compararIgualdad(izq, der), nil
	case "<>":
		return !compararIgualdad(izq, der), nil
	case "<", ">", "<=", ">=":
		return compararOrden(izq, der, operador)
	}

	return false, fmt.Errorf("operador no soportado: %q", operador)
}

// compararIgualdad compara dos valores considerando tipos numéricos.
func compararIgualdad(izq, der any) bool {
	af, aNum := anyToFloat64(izq)
	bf, bNum := anyToFloat64(der)
	if aNum && bNum {
		return af == bf
	}
	return fmt.Sprint(izq) == fmt.Sprint(der)
}

// compararOrden evalúa operadores de orden (<, >, <=, >=).
// Compara numéricamente si ambos lados son numéricos, textualmente en caso contrario.
func compararOrden(izq, der any, op string) (bool, error) {
	af, aNum := anyToFloat64(izq)
	bf, bNum := anyToFloat64(der)

	if aNum && bNum {
		switch op {
		case "<":
			return af < bf, nil
		case ">":
			return af > bf, nil
		case "<=":
			return af <= bf, nil
		case ">=":
			return af >= bf, nil
		}
	}

	as := fmt.Sprint(izq)
	bs := fmt.Sprint(der)

	switch op {
	case "<":
		return as < bs, nil
	case ">":
		return as > bs, nil
	case "<=":
		return as <= bs, nil
	case ">=":
		return as >= bs, nil
	}

	return false, fmt.Errorf("operador de orden no soportado: %q", op)
}
