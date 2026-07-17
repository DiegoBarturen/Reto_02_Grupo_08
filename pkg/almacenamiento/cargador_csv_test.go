package almacenamiento

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCargarCSVComoTexto(t *testing.T) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(directorioTemporal, "estudiantes.csv")

	contenidoCSV := `id,nombre,promedio,activo
1,Ana,17.5,true
2,Carlos,14.8,false
3,Lucia,18.2,true
`

	err := os.WriteFile(rutaCSV, []byte(contenidoCSV), 0644)
	if err != nil {
		t.Fatalf("no se pudo crear el CSV temporal: %v", err)
	}

	tabla, err := CargarCSVComoTexto("estudiantes", rutaCSV)
	if err != nil {
		t.Fatalf(
			"CargarCSVComoTexto devolvió un error inesperado: %v",
			err,
		)
	}

	if tabla.Nombre != "estudiantes" {
		t.Errorf(
			"nombre incorrecto: se esperaba %q y se obtuvo %q",
			"estudiantes",
			tabla.Nombre,
		)
	}

	if len(tabla.Esquema.Columnas) != 4 {
		t.Errorf(
			"cantidad de columnas incorrecta: se esperaban 4 y se obtuvieron %d",
			len(tabla.Esquema.Columnas),
		)
	}

	if len(tabla.Filas) != 3 {
		t.Errorf(
			"cantidad de filas incorrecta: se esperaban 3 y se obtuvieron %d",
			len(tabla.Filas),
		)
	}

	primerValor := tabla.Filas[0].Datos[0]

	if primerValor.Tipo != TipoTexto {
		t.Errorf(
			"tipo incorrecto: se esperaba %s y se obtuvo %s",
			TipoTexto,
			primerValor.Tipo,
		)
	}

	if primerValor.Dato != "1" {
		t.Errorf(
			"valor incorrecto: se esperaba %q y se obtuvo %v",
			"1",
			primerValor.Dato,
		)
	}
}

func TestCargarCSVComoTextoCSVInvalido(t *testing.T) {
	pruebas := []struct {
		nombre      string
		nombreTabla string
		contenido   string
	}{
		{
			nombre:      "nombre de tabla vacío",
			nombreTabla: "",
			contenido: `id,nombre
1,Ana
`,
		},
		{
			nombre:      "archivo vacío",
			nombreTabla: "estudiantes",
			contenido:   "",
		},
		{
			nombre:      "nombre de columna vacío",
			nombreTabla: "estudiantes",
			contenido: `id,
1,Ana
`,
		},
		{
			nombre:      "columnas duplicadas",
			nombreTabla: "estudiantes",
			contenido: `id,ID
1,2
`,
		},
		{
			nombre:      "fila con columnas faltantes",
			nombreTabla: "estudiantes",
			contenido: `id,nombre,edad
1,Ana,20
2,Carlos
`,
		},
		{
			nombre:      "fila con columnas adicionales",
			nombreTabla: "estudiantes",
			contenido: `id,nombre
1,Ana,20
`,
		},
	}

	for _, prueba := range pruebas {
		t.Run(prueba.nombre, func(t *testing.T) {
			directorioTemporal := t.TempDir()
			rutaCSV := filepath.Join(directorioTemporal, "datos.csv")

			err := os.WriteFile(
				rutaCSV,
				[]byte(prueba.contenido),
				0644,
			)

			if err != nil {
				t.Fatalf(
					"no se pudo crear el CSV temporal: %v",
					err,
				)
			}

			tabla, err := CargarCSVComoTexto(
				prueba.nombreTabla,
				rutaCSV,
			)

			if err == nil {
				t.Fatalf(
					"se esperaba un error, pero se obtuvo la tabla: %#v",
					tabla,
				)
			}

			if tabla != nil {
				t.Error(
					"se esperaba una tabla nil cuando ocurre un error",
				)
			}
		})
	}
}

func TestCargarCSVComoTextoArchivoInexistente(t *testing.T) {
	tabla, err := CargarCSVComoTexto(
		"estudiantes",
		"archivo_que_no_existe.csv",
	)

	if err == nil {
		t.Fatal(
			"se esperaba un error al cargar un archivo inexistente",
		)
	}

	if tabla != nil {
		t.Error("se esperaba una tabla nil")
	}
}
