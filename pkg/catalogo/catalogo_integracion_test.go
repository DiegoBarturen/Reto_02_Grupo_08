package catalogo

import (
	"os"
	"path/filepath"
	"testing"

	"reto_02_grupo_08/pkg/almacenamiento"
)

func TestCargarCSVRegistrarYConsultarCatalogo(t *testing.T) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(directorioTemporal, "estudiantes.csv")

	contenidoCSV := `id,nombre,promedio,activo
1,Ana,17.5,true
2,Carlos,14.8,false
3,Lucia,18.2,true
`

	err := os.WriteFile(rutaCSV, []byte(contenidoCSV), 0644)
	if err != nil {
		t.Fatalf("no se pudo crear el archivo CSV temporal: %v", err)
	}

	tabla, err := almacenamiento.CargarCSVComoTexto(
		"estudiantes",
		rutaCSV,
	)
	if err != nil {
		t.Fatalf("no se pudo cargar el archivo CSV: %v", err)
	}

	catalogo := Nuevo()

	err = catalogo.RegistrarTabla(tabla)
	if err != nil {
		t.Fatalf("no se pudo registrar la tabla: %v", err)
	}

	if !catalogo.ExisteTabla("estudiantes") {
		t.Fatal("la tabla estudiantes debería existir en el catálogo")
	}

	tablaObtenida, err := catalogo.ObtenerTabla("estudiantes")
	if err != nil {
		t.Fatalf("no se pudo obtener la tabla registrada: %v", err)
	}

	if tablaObtenida.Nombre != "estudiantes" {
		t.Errorf(
			"se esperaba el nombre %q y se obtuvo %q",
			"estudiantes",
			tablaObtenida.Nombre,
		)
	}

	if len(tablaObtenida.Filas) != 3 {
		t.Errorf(
			"se esperaban 3 filas y se obtuvieron %d",
			len(tablaObtenida.Filas),
		)
	}

	esquema, err := catalogo.ObtenerEsquema("estudiantes")
	if err != nil {
		t.Fatalf("no se pudo obtener el esquema: %v", err)
	}

	if len(esquema.Columnas) != 4 {
		t.Fatalf(
			"se esperaban 4 columnas y se obtuvieron %d",
			len(esquema.Columnas),
		)
	}

	nombresEsperados := []string{
		"id",
		"nombre",
		"promedio",
		"activo",
	}

	for indice, nombreEsperado := range nombresEsperados {
		columna := esquema.Columnas[indice]

		if columna.Nombre != nombreEsperado {
			t.Errorf(
				"columna %d: se esperaba %q y se obtuvo %q",
				indice+1,
				nombreEsperado,
				columna.Nombre,
			)
		}

		if columna.Tipo != almacenamiento.TipoTexto {
			t.Errorf(
				"columna %q: se esperaba el tipo %s y se obtuvo %s",
				columna.Nombre,
				almacenamiento.TipoTexto,
				columna.Tipo,
			)
		}
	}
}
