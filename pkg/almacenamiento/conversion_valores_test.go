package almacenamiento

import (
	"reflect"
	"testing"
)

func TestConvertirValor(t *testing.T) {
	pruebas := []struct {
		nombre        string
		valorOriginal string
		tipo          TipoDato
		valorEsperado Valor
	}{
		{
			nombre:        "entero positivo",
			valorOriginal: "25",
			tipo:          TipoEntero,
			valorEsperado: Valor{
				Tipo:   TipoEntero,
				Dato:   int64(25),
				EsNulo: false,
			},
		},
		{
			nombre:        "entero negativo",
			valorOriginal: "-10",
			tipo:          TipoEntero,
			valorEsperado: Valor{
				Tipo:   TipoEntero,
				Dato:   int64(-10),
				EsNulo: false,
			},
		},
		{
			nombre:        "decimal",
			valorOriginal: "17.5",
			tipo:          TipoDecimal,
			valorEsperado: Valor{
				Tipo:   TipoDecimal,
				Dato:   float64(17.5),
				EsNulo: false,
			},
		},
		{
			nombre:        "entero convertido como decimal",
			valorOriginal: "15",
			tipo:          TipoDecimal,
			valorEsperado: Valor{
				Tipo:   TipoDecimal,
				Dato:   float64(15),
				EsNulo: false,
			},
		},
		{
			nombre:        "booleano verdadero",
			valorOriginal: "true",
			tipo:          TipoBooleano,
			valorEsperado: Valor{
				Tipo:   TipoBooleano,
				Dato:   true,
				EsNulo: false,
			},
		},
		{
			nombre:        "booleano falso en mayúsculas",
			valorOriginal: "FALSE",
			tipo:          TipoBooleano,
			valorEsperado: Valor{
				Tipo:   TipoBooleano,
				Dato:   false,
				EsNulo: false,
			},
		},
		{
			nombre:        "texto",
			valorOriginal: "Ana",
			tipo:          TipoTexto,
			valorEsperado: Valor{
				Tipo:   TipoTexto,
				Dato:   "Ana",
				EsNulo: false,
			},
		},
		{
			nombre:        "texto con ceros iniciales",
			valorOriginal: "00125",
			tipo:          TipoTexto,
			valorEsperado: Valor{
				Tipo:   TipoTexto,
				Dato:   "00125",
				EsNulo: false,
			},
		},
		{
			nombre:        "campo vacío como nulo",
			valorOriginal: "",
			tipo:          TipoEntero,
			valorEsperado: Valor{
				Tipo:   TipoEntero,
				Dato:   nil,
				EsNulo: true,
			},
		},
		{
			nombre:        "NULL como nulo",
			valorOriginal: "NULL",
			tipo:          TipoDecimal,
			valorEsperado: Valor{
				Tipo:   TipoDecimal,
				Dato:   nil,
				EsNulo: true,
			},
		},
	}

	for _, prueba := range pruebas {
		t.Run(prueba.nombre, func(t *testing.T) {
			valorObtenido, err := convertirValor(
				prueba.valorOriginal,
				prueba.tipo,
			)

			if err != nil {
				t.Fatalf(
					"convertirValor devolvió un error inesperado: %v",
					err,
				)
			}

			if !reflect.DeepEqual(valorObtenido, prueba.valorEsperado) {
				t.Errorf(
					"se esperaba %#v y se obtuvo %#v",
					prueba.valorEsperado,
					valorObtenido,
				)
			}
		})
	}
}

func TestConvertirValorInvalido(t *testing.T) {
	pruebas := []struct {
		nombre        string
		valorOriginal string
		tipo          TipoDato
	}{
		{
			nombre:        "entero inválido",
			valorOriginal: "veinticinco",
			tipo:          TipoEntero,
		},
		{
			nombre:        "decimal inválido",
			valorOriginal: "diecisiete",
			tipo:          TipoDecimal,
		},
		{
			nombre:        "booleano inválido",
			valorOriginal: "si",
			tipo:          TipoBooleano,
		},
		{
			nombre:        "tipo no soportado",
			valorOriginal: "dato",
			tipo:          TipoDato("FECHA"),
		},
	}

	for _, prueba := range pruebas {
		t.Run(prueba.nombre, func(t *testing.T) {
			_, err := convertirValor(
				prueba.valorOriginal,
				prueba.tipo,
			)

			if err == nil {
				t.Fatal(
					"se esperaba un error de conversión",
				)
			}
		})
	}
}
