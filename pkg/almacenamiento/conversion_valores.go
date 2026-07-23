package almacenamiento

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func convertirValor(valorOriginal string, tipo TipoDato) (Valor, error) {
	valorLimpio := strings.TrimSpace(valorOriginal)

	if esValorNulo(valorLimpio) {
		return Valor{
			Tipo:   tipo,
			Dato:   nil,
			EsNulo: true,
		}, nil
	}

	switch tipo {
	case TipoEntero:
		numero, err := strconv.ParseInt(valorLimpio, 10, 64)
		if err != nil {
			return Valor{}, fmt.Errorf(
				"el valor %q no es un entero válido",
				valorOriginal,
			)
		}

		return Valor{
			Tipo:   TipoEntero,
			Dato:   numero,
			EsNulo: false,
		}, nil

	case TipoDecimal:
		numero, err := strconv.ParseFloat(valorLimpio, 64)
		if err != nil ||
			math.IsNaN(numero) ||
			math.IsInf(numero, 0) {
			return Valor{}, fmt.Errorf(
				"el valor %q no es un decimal válido",
				valorOriginal,
			)
		}

		return Valor{
			Tipo:   TipoDecimal,
			Dato:   numero,
			EsNulo: false,
		}, nil

	case TipoBooleano:
		if strings.EqualFold(valorLimpio, "true") {
			return Valor{
				Tipo:   TipoBooleano,
				Dato:   true,
				EsNulo: false,
			}, nil
		}

		if strings.EqualFold(valorLimpio, "false") {
			return Valor{
				Tipo:   TipoBooleano,
				Dato:   false,
				EsNulo: false,
			}, nil
		}

		return Valor{}, fmt.Errorf(
			"el valor %q no es un booleano válido; se esperaba true o false",
			valorOriginal,
		)

	case TipoTexto:
		return Valor{
			Tipo:   TipoTexto,
			Dato:   valorOriginal,
			EsNulo: false,
		}, nil

	default:
		return Valor{}, fmt.Errorf(
			"tipo de dato no soportado: %q",
			tipo,
		)
	}
}
