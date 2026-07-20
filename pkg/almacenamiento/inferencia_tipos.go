package almacenamiento

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func esValorNulo(valor string) bool {
	valorLimpio := strings.TrimSpace(valor)

	return valorLimpio == "" ||
		strings.EqualFold(valorLimpio, "NULL")
}

func inferirTipoValor(valor string) (TipoDato, bool) {
	valorLimpio := strings.TrimSpace(valor)

	if esValorNulo(valorLimpio) {
		return "", false
	}

	if strings.EqualFold(valorLimpio, "true") ||
		strings.EqualFold(valorLimpio, "false") {
		return TipoBooleano, true
	}

	if _, err := strconv.ParseInt(valorLimpio, 10, 64); err == nil {
		return TipoEntero, true
	}

	numeroDecimal, err := strconv.ParseFloat(valorLimpio, 64)
	if err == nil &&
		!math.IsNaN(numeroDecimal) &&
		!math.IsInf(numeroDecimal, 0) {
		return TipoDecimal, true
	}

	return TipoTexto, true
}

func combinarTipos(tipoActual, tipoNuevo TipoDato) TipoDato {
	if tipoActual == "" {
		return tipoNuevo
	}

	if tipoNuevo == "" {
		return tipoActual
	}

	if tipoActual == tipoNuevo {
		return tipoActual
	}

	esCombinacionNumerica :=
		(tipoActual == TipoEntero && tipoNuevo == TipoDecimal) ||
			(tipoActual == TipoDecimal && tipoNuevo == TipoEntero)

	if esCombinacionNumerica {
		return TipoDecimal
	}

	return TipoTexto
}

func inferirTiposColumnas(
	registros [][]string,
	cantidadColumnas int,
) ([]TipoDato, error) {
	if cantidadColumnas <= 0 {
		return nil, fmt.Errorf(
			"la cantidad de columnas debe ser mayor que cero",
		)
	}

	tipos := make([]TipoDato, cantidadColumnas)

	for indiceFila, registro := range registros {
		if len(registro) != cantidadColumnas {
			return nil, fmt.Errorf(
				"fila %d: se esperaban %d columnas, pero se encontraron %d",
				indiceFila+2,
				cantidadColumnas,
				len(registro),
			)
		}

		for indiceColumna, valorOriginal := range registro {
			tipoValor, tieneTipo := inferirTipoValor(valorOriginal)

			if !tieneTipo {
				continue
			}

			tipos[indiceColumna] = combinarTipos(
				tipos[indiceColumna],
				tipoValor,
			)
		}
	}

	for indice := range tipos {
		if tipos[indice] == "" {
			tipos[indice] = TipoTexto
		}
	}

	return tipos, nil
}
