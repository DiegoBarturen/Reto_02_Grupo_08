package almacenamiento

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCargarCSVConInferencia(t *testing.T) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(
		directorioTemporal,
		"estudiantes.csv",
	)

	contenidoCSV := `id,nombre,promedio,activo,codigo,observacion
1,Ana,17.5,true,001,NULL
2,Carlos,15,false,002,
3,Lucia,18.25,TRUE,ABC,NULL
`

	err := os.WriteFile(
		rutaCSV,
		[]byte(contenidoCSV),
		0644,
	)

	if err != nil {
		t.Fatalf(
			"no se pudo crear el CSV temporal: %v",
			err,
		)
	}

	tabla, err := CargarCSVConInferencia(
		"estudiantes",
		rutaCSV,
	)

	if err != nil {
		t.Fatalf(
			"CargarCSVConInferencia devolvió un error inesperado: %v",
			err,
		)
	}

	if tabla.Nombre != "estudiantes" {
		t.Errorf(
			"se esperaba el nombre %q y se obtuvo %q",
			"estudiantes",
			tabla.Nombre,
		)
	}

	if len(tabla.Filas) != 3 {
		t.Fatalf(
			"se esperaban 3 filas y se obtuvieron %d",
			len(tabla.Filas),
		)
	}

	tiposEsperados := []TipoDato{
		TipoEntero,
		TipoTexto,
		TipoDecimal,
		TipoBooleano,
		TipoTexto,
		TipoTexto,
	}

	for indice, tipoEsperado := range tiposEsperados {
		tipoObtenido := tabla.Esquema.Columnas[indice].Tipo

		if tipoObtenido != tipoEsperado {
			t.Errorf(
				"columna %q: se esperaba el tipo %s y se obtuvo %s",
				tabla.Esquema.Columnas[indice].Nombre,
				tipoEsperado,
				tipoObtenido,
			)
		}
	}

	verificarTiposValoresPrimeraFila(t, tabla)
	verificarValoresNulos(t, tabla)
	verificarCodigoComoTexto(t, tabla)
}

func verificarTiposValoresPrimeraFila(
	t *testing.T,
	tabla *Tabla,
) {
	t.Helper()

	primeraFila := tabla.Filas[0]

	if _, correcto := primeraFila.Datos[0].Dato.(int64); !correcto {
		t.Errorf(
			"el valor de id debería almacenarse como int64, pero se obtuvo %T",
			primeraFila.Datos[0].Dato,
		)
	}

	if _, correcto := primeraFila.Datos[1].Dato.(string); !correcto {
		t.Errorf(
			"el nombre debería almacenarse como string, pero se obtuvo %T",
			primeraFila.Datos[1].Dato,
		)
	}

	if _, correcto := primeraFila.Datos[2].Dato.(float64); !correcto {
		t.Errorf(
			"el promedio debería almacenarse como float64, pero se obtuvo %T",
			primeraFila.Datos[2].Dato,
		)
	}

	if _, correcto := primeraFila.Datos[3].Dato.(bool); !correcto {
		t.Errorf(
			"el campo activo debería almacenarse como bool, pero se obtuvo %T",
			primeraFila.Datos[3].Dato,
		)
	}
}

func verificarValoresNulos(
	t *testing.T,
	tabla *Tabla,
) {
	t.Helper()

	observacionPrimeraFila := tabla.Filas[0].Datos[5]
	observacionSegundaFila := tabla.Filas[1].Datos[5]

	if !observacionPrimeraFila.EsNulo {
		t.Error(
			"el texto NULL debería convertirse en un valor nulo",
		)
	}

	if observacionPrimeraFila.Dato != nil {
		t.Error(
			"el dato de un valor nulo debería ser nil",
		)
	}

	if !observacionSegundaFila.EsNulo {
		t.Error(
			"un campo vacío debería convertirse en un valor nulo",
		)
	}
}

func verificarCodigoComoTexto(
	t *testing.T,
	tabla *Tabla,
) {
	t.Helper()

	codigoPrimeraFila := tabla.Filas[0].Datos[4]

	if codigoPrimeraFila.Tipo != TipoTexto {
		t.Errorf(
			"la columna código debería inferirse como TEXTO y se obtuvo %s",
			codigoPrimeraFila.Tipo,
		)
	}

	codigo, correcto := codigoPrimeraFila.Dato.(string)

	if !correcto {
		t.Fatalf(
			"el código debería almacenarse como string, pero se obtuvo %T",
			codigoPrimeraFila.Dato,
		)
	}

	if codigo != "001" {
		t.Errorf(
			"se esperaba conservar el código %q y se obtuvo %q",
			"001",
			codigo,
		)
	}
}

func TestCargarCSVConInferenciaFilaInvalida(t *testing.T) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(
		directorioTemporal,
		"datos_invalidos.csv",
	)

	contenidoCSV := `id,nombre,edad
1,Ana,20
2,Carlos
`

	err := os.WriteFile(
		rutaCSV,
		[]byte(contenidoCSV),
		0644,
	)

	if err != nil {
		t.Fatalf(
			"no se pudo crear el CSV temporal: %v",
			err,
		)
	}

	tabla, err := CargarCSVConInferencia(
		"estudiantes",
		rutaCSV,
	)

	if err == nil {
		t.Fatal(
			"se esperaba un error por cantidad incorrecta de columnas",
		)
	}

	if tabla != nil {
		t.Error(
			"se esperaba una tabla nula cuando ocurre un error",
		)
	}
}

func TestCargarCSVConInferenciaSoloEncabezado(
	t *testing.T,
) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(
		directorioTemporal,
		"tabla_vacia.csv",
	)

	contenidoCSV := `id,nombre,activo
`

	err := os.WriteFile(
		rutaCSV,
		[]byte(contenidoCSV),
		0644,
	)

	if err != nil {
		t.Fatalf(
			"no se pudo crear el CSV temporal: %v",
			err,
		)
	}

	tabla, err := CargarCSVConInferencia(
		"tabla_vacia",
		rutaCSV,
	)

	if err != nil {
		t.Fatalf(
			"se produjo un error inesperado: %v",
			err,
		)
	}

	if len(tabla.Filas) != 0 {
		t.Errorf(
			"se esperaban 0 filas y se obtuvieron %d",
			len(tabla.Filas),
		)
	}

	for _, columna := range tabla.Esquema.Columnas {
		if columna.Tipo != TipoTexto {
			t.Errorf(
				"una columna sin datos debería usar TEXTO y se obtuvo %s",
				columna.Tipo,
			)
		}
	}
}
