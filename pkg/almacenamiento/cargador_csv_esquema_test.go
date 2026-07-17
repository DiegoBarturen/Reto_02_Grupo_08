package almacenamiento

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCargarCSVConEsquema(t *testing.T) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(
		directorioTemporal,
		"productos.csv",
	)

	contenidoCSV := `codigo,nombre,precio,activo
001,Teclado,89.90,true
002,Mouse,45,false
003,Monitor,650.50,true
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

	esquema := Esquema{
		Columnas: []Columna{
			{
				Nombre: "codigo",
				Tipo:   TipoTexto,
			},
			{
				Nombre: "nombre",
				Tipo:   TipoTexto,
			},
			{
				Nombre: "precio",
				Tipo:   TipoDecimal,
			},
			{
				Nombre: "activo",
				Tipo:   TipoBooleano,
			},
		},
	}

	tabla, err := CargarCSVConEsquema(
		"productos",
		rutaCSV,
		esquema,
	)

	if err != nil {
		t.Fatalf(
			"CargarCSVConEsquema devolvió un error inesperado: %v",
			err,
		)
	}

	if tabla.Nombre != "productos" {
		t.Errorf(
			"se esperaba el nombre %q y se obtuvo %q",
			"productos",
			tabla.Nombre,
		)
	}

	if len(tabla.Filas) != 3 {
		t.Fatalf(
			"se esperaban 3 filas y se obtuvieron %d",
			len(tabla.Filas),
		)
	}

	codigo := tabla.Filas[0].Datos[0]

	if codigo.Tipo != TipoTexto {
		t.Errorf(
			"se esperaba TEXTO y se obtuvo %s",
			codigo.Tipo,
		)
	}

	codigoTexto, correcto := codigo.Dato.(string)

	if !correcto {
		t.Fatalf(
			"el código debería almacenarse como string, pero se obtuvo %T",
			codigo.Dato,
		)
	}

	if codigoTexto != "001" {
		t.Errorf(
			"se esperaba conservar %q y se obtuvo %q",
			"001",
			codigoTexto,
		)
	}

	precio := tabla.Filas[1].Datos[2]

	precioDecimal, correcto := precio.Dato.(float64)

	if !correcto {
		t.Fatalf(
			"el precio debería almacenarse como float64, pero se obtuvo %T",
			precio.Dato,
		)
	}

	if precioDecimal != 45 {
		t.Errorf(
			"se esperaba el precio 45 y se obtuvo %v",
			precioDecimal,
		)
	}

	activo := tabla.Filas[0].Datos[3]

	if _, correcto := activo.Dato.(bool); !correcto {
		t.Errorf(
			"el campo activo debería almacenarse como bool, pero se obtuvo %T",
			activo.Dato,
		)
	}
}

func TestCargarCSVConEsquemaValorInvalido(t *testing.T) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(
		directorioTemporal,
		"productos_invalidos.csv",
	)

	contenidoCSV := `codigo,precio
001,89.90
002,gratis
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

	esquema := Esquema{
		Columnas: []Columna{
			{
				Nombre: "codigo",
				Tipo:   TipoTexto,
			},
			{
				Nombre: "precio",
				Tipo:   TipoDecimal,
			},
		},
	}

	tabla, err := CargarCSVConEsquema(
		"productos",
		rutaCSV,
		esquema,
	)

	if err == nil {
		t.Fatal(
			"se esperaba un error por decimal inválido",
		)
	}

	if tabla != nil {
		t.Error(
			"se esperaba una tabla nula cuando ocurre un error",
		)
	}
}

func TestCargarCSVConEsquemaEncabezadoIncompatible(
	t *testing.T,
) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(
		directorioTemporal,
		"encabezado_invalido.csv",
	)

	contenidoCSV := `codigo,descripcion
001,Teclado
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

	esquema := Esquema{
		Columnas: []Columna{
			{
				Nombre: "codigo",
				Tipo:   TipoTexto,
			},
			{
				Nombre: "nombre",
				Tipo:   TipoTexto,
			},
		},
	}

	_, err = CargarCSVConEsquema(
		"productos",
		rutaCSV,
		esquema,
	)

	if err == nil {
		t.Fatal(
			"se esperaba un error por encabezado incompatible",
		)
	}
}

func TestValidarEsquemaInvalido(t *testing.T) {
	pruebas := []struct {
		nombre  string
		esquema Esquema
	}{
		{
			nombre:  "esquema sin columnas",
			esquema: Esquema{},
		},
		{
			nombre: "columna con nombre vacío",
			esquema: Esquema{
				Columnas: []Columna{
					{
						Nombre: "",
						Tipo:   TipoTexto,
					},
				},
			},
		},
		{
			nombre: "columnas duplicadas",
			esquema: Esquema{
				Columnas: []Columna{
					{
						Nombre: "codigo",
						Tipo:   TipoTexto,
					},
					{
						Nombre: "CODIGO",
						Tipo:   TipoEntero,
					},
				},
			},
		},
		{
			nombre: "tipo no soportado",
			esquema: Esquema{
				Columnas: []Columna{
					{
						Nombre: "fecha",
						Tipo:   TipoDato("FECHA"),
					},
				},
			},
		},
	}

	for _, prueba := range pruebas {
		t.Run(prueba.nombre, func(t *testing.T) {
			_, err := validarYCopiarEsquema(
				prueba.esquema,
			)

			if err == nil {
				t.Fatal(
					"se esperaba un error de validación",
				)
			}
		})
	}
}

func TestCargarCSVConEsquemaValorNulo(t *testing.T) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(
		directorioTemporal,
		"personas.csv",
	)

	contenidoCSV := `id,nombre,edad
1,Ana,20
2,Carlos,NULL
3,Lucia,
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

	esquema := Esquema{
		Columnas: []Columna{
			{
				Nombre: "id",
				Tipo:   TipoEntero,
			},
			{
				Nombre: "nombre",
				Tipo:   TipoTexto,
			},
			{
				Nombre: "edad",
				Tipo:   TipoEntero,
			},
		},
	}

	tabla, err := CargarCSVConEsquema(
		"personas",
		rutaCSV,
		esquema,
	)

	if err != nil {
		t.Fatalf(
			"se produjo un error inesperado: %v",
			err,
		)
	}

	edadSegundaFila := tabla.Filas[1].Datos[2]
	edadTerceraFila := tabla.Filas[2].Datos[2]

	if !edadSegundaFila.EsNulo {
		t.Error(
			"NULL debería convertirse en un valor nulo",
		)
	}

	if !edadTerceraFila.EsNulo {
		t.Error(
			"el campo vacío debería convertirse en un valor nulo",
		)
	}

	if edadSegundaFila.Tipo != TipoEntero {
		t.Errorf(
			"el valor nulo debería conservar el tipo ENTERO y se obtuvo %s",
			edadSegundaFila.Tipo,
		)
	}
}
