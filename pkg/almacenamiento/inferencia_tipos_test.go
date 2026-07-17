package almacenamiento

import (
	"reflect"
	"testing"
)

func TestInferirTipoValor(t *testing.T) {
	pruebas := []struct {
		nombre       string
		valor        string
		tipoEsperado TipoDato
		tieneTipo    bool
	}{
		{
			nombre:       "entero positivo",
			valor:        "25",
			tipoEsperado: TipoEntero,
			tieneTipo:    true,
		},
		{
			nombre:       "entero negativo",
			valor:        "-8",
			tipoEsperado: TipoEntero,
			tieneTipo:    true,
		},
		{
			nombre:       "decimal",
			valor:        "17.5",
			tipoEsperado: TipoDecimal,
			tieneTipo:    true,
		},
		{
			nombre:       "booleano verdadero",
			valor:        "true",
			tipoEsperado: TipoBooleano,
			tieneTipo:    true,
		},
		{
			nombre:       "booleano falso en mayúsculas",
			valor:        "FALSE",
			tipoEsperado: TipoBooleano,
			tieneTipo:    true,
		},
		{
			nombre:       "texto",
			valor:        "Ana",
			tipoEsperado: TipoTexto,
			tieneTipo:    true,
		},
		{
			nombre:       "campo vacío",
			valor:        "",
			tipoEsperado: "",
			tieneTipo:    false,
		},
		{
			nombre:       "valor NULL",
			valor:        "NULL",
			tipoEsperado: "",
			tieneTipo:    false,
		},
	}

	for _, prueba := range pruebas {
		t.Run(prueba.nombre, func(t *testing.T) {
			tipoObtenido, tieneTipoObtenido :=
				inferirTipoValor(prueba.valor)

			if tipoObtenido != prueba.tipoEsperado {
				t.Errorf(
					"se esperaba el tipo %q y se obtuvo %q",
					prueba.tipoEsperado,
					tipoObtenido,
				)
			}

			if tieneTipoObtenido != prueba.tieneTipo {
				t.Errorf(
					"se esperaba tieneTipo=%v y se obtuvo %v",
					prueba.tieneTipo,
					tieneTipoObtenido,
				)
			}
		})
	}
}

func TestCombinarTipos(t *testing.T) {
	pruebas := []struct {
		nombre       string
		tipoActual   TipoDato
		tipoNuevo    TipoDato
		tipoEsperado TipoDato
	}{
		{
			nombre:       "dos enteros",
			tipoActual:   TipoEntero,
			tipoNuevo:    TipoEntero,
			tipoEsperado: TipoEntero,
		},
		{
			nombre:       "entero y decimal",
			tipoActual:   TipoEntero,
			tipoNuevo:    TipoDecimal,
			tipoEsperado: TipoDecimal,
		},
		{
			nombre:       "decimal y entero",
			tipoActual:   TipoDecimal,
			tipoNuevo:    TipoEntero,
			tipoEsperado: TipoDecimal,
		},
		{
			nombre:       "booleanos",
			tipoActual:   TipoBooleano,
			tipoNuevo:    TipoBooleano,
			tipoEsperado: TipoBooleano,
		},
		{
			nombre:       "entero y texto",
			tipoActual:   TipoEntero,
			tipoNuevo:    TipoTexto,
			tipoEsperado: TipoTexto,
		},
		{
			nombre:       "booleano y entero",
			tipoActual:   TipoBooleano,
			tipoNuevo:    TipoEntero,
			tipoEsperado: TipoTexto,
		},
	}

	for _, prueba := range pruebas {
		t.Run(prueba.nombre, func(t *testing.T) {
			tipoObtenido := combinarTipos(
				prueba.tipoActual,
				prueba.tipoNuevo,
			)

			if tipoObtenido != prueba.tipoEsperado {
				t.Errorf(
					"se esperaba %s y se obtuvo %s",
					prueba.tipoEsperado,
					tipoObtenido,
				)
			}
		})
	}
}

func TestInferirTiposColumnas(t *testing.T) {
	registros := [][]string{
		{"1", "Ana", "17.5", "true", ""},
		{"2", "Luis", "15", "false", "NULL"},
		{"3", "Lucia", "18.25", "true", ""},
	}

	tiposEsperados := []TipoDato{
		TipoEntero,
		TipoTexto,
		TipoDecimal,
		TipoBooleano,
		TipoTexto,
	}

	tiposObtenidos, err := inferirTiposColumnas(
		registros,
		5,
	)
	if err != nil {
		t.Fatalf(
			"inferirTiposColumnas devolvió un error inesperado: %v",
			err,
		)
	}

	if !reflect.DeepEqual(tiposObtenidos, tiposEsperados) {
		t.Errorf(
			"se esperaba %v y se obtuvo %v",
			tiposEsperados,
			tiposObtenidos,
		)
	}
}

func TestInferirTiposColumnasFilaInvalida(t *testing.T) {
	registros := [][]string{
		{"1", "Ana"},
		{"2"},
	}

	_, err := inferirTiposColumnas(registros, 2)

	if err == nil {
		t.Fatal(
			"se esperaba un error por cantidad incorrecta de columnas",
		)
	}
}
