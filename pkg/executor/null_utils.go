package executor

import (
	"fmt"
	"strings"

	"reto_02_grupo_08/pkg/almacenamiento"
)

// indiceDeLaColumna busca el índice de una columna por nombre (insensible a mayúsculas).
// Acepta tanto nombres simples ("id") como calificados ("empleados.id").
// Retorna error si la columna no existe en el esquema.
func indiceDeLaColumna(esquema almacenamiento.Esquema, nombre string) (int, error) {
	lower := strings.ToLower(nombre)

	for i, col := range esquema.Columnas {
		if strings.ToLower(col.Nombre) == lower {
			return i, nil
		}
	}

	if strings.Contains(lower, ".") {
		parte := lower[strings.LastIndex(lower, ".")+1:]
		for i, col := range esquema.Columnas {
			if strings.ToLower(col.Nombre) == parte {
				return i, nil
			}
			if strings.HasSuffix(strings.ToLower(col.Nombre), "."+parte) {
				return i, nil
			}
		}
	}

	return -1, fmt.Errorf("la columna %q no existe en el esquema", nombre)
}

// CompararValores compara dos Valores tipados siguiendo la semántica SQL:
//   - Si alguno es NULL, retorna (0, false) — no son comparables.
//   - Compara numéricamente si ambos son numéricos (int64 o float64).
//   - Compara como texto en cualquier otro caso.
//
// El resultado es: negativo si a < b, cero si a == b, positivo si a > b.
func CompararValores(a, b almacenamiento.Valor) (int, bool) {
	if a.EsNulo || b.EsNulo {
		return 0, false
	}

	af, aEsNum := anyToFloat64(a.Dato)
	bf, bEsNum := anyToFloat64(b.Dato)

	if aEsNum && bEsNum {
		switch {
		case af < bf:
			return -1, true
		case af > bf:
			return 1, true
		default:
			return 0, true
		}
	}

	as := valorComoTexto(a)
	bs := valorComoTexto(b)

	switch {
	case as < bs:
		return -1, true
	case as > bs:
		return 1, true
	default:
		return 0, true
	}
}

// anyToFloat64 convierte valores numéricos nativos a float64.
// Retorna (valor, true) si la conversión es posible, (0, false) en caso contrario.
func anyToFloat64(v any) (float64, bool) {
	switch d := v.(type) {
	case int64:
		return float64(d), true
	case float64:
		return d, true
	case int:
		return float64(d), true
	}
	return 0, false
}

// valorComoTexto convierte un Valor a su representación textual.
func valorComoTexto(v almacenamiento.Valor) string {
	if v.EsNulo {
		return "NULL"
	}
	if s, ok := v.Dato.(string); ok {
		return s
	}
	return fmt.Sprint(v.Dato)
}
