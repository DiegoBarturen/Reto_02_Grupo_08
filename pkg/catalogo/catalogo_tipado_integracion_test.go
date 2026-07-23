package catalogo

import (
	"os"
	"path/filepath"
	"testing"

	"reto_02_grupo_08/pkg/almacenamiento"
)

func TestRegistrarCSVConTiposInferidos(t *testing.T) {
	directorioTemporal := t.TempDir()
	rutaCSV := filepath.Join(directorioTemporal, "estudiantes.csv")

	contenidoCSV := `id,nombre,promedio,activo
1,Ana,17.5,true
2,Carlos,15,false
3,Lucia,18.25,true
`

	err := os.WriteFile(rutaCSV, []byte(contenidoCSV), 0644)
	if err != nil {
		t.Fatalf("no se pudo crear el CSV temporal: %v", err)
	}

	tabla, err := almacenamiento.CargarCSVConInferencia(
		"estudiantes",
		rutaCSV,
	)
	if err != nil {
		t.Fatalf("no se pudo cargar el CSV: %v", err)
	}

	catalogo := Nuevo()

	err = catalogo.RegistrarTabla(tabla)
	if err != nil {
		t.Fatalf("no se pudo registrar la tabla: %v", err)
	}

	esquema, err := catalogo.ObtenerEsquema("estudiantes")
	if err != nil {
		t.Fatalf("no se pudo consultar el esquema: %v", err)
	}

	tiposEsperados := []almacenamiento.TipoDato{
		almacenamiento.TipoEntero,
		almacenamiento.TipoTexto,
		almacenamiento.TipoDecimal,
		almacenamiento.TipoBooleano,
	}

	if len(esquema.Columnas) != len(tiposEsperados) {
		t.Fatalf(
			"se esperaban %d columnas y se obtuvieron %d",
			len(tiposEsperados),
			len(esquema.Columnas),
		)
	}

	for indice, tipoEsperado := range tiposEsperados {
		tipoObtenido := esquema.Columnas[indice].Tipo

		if tipoObtenido != tipoEsperado {
			t.Errorf(
				"columna %q: se esperaba %s y se obtuvo %s",
				esquema.Columnas[indice].Nombre,
				tipoEsperado,
				tipoObtenido,
			)
		}
	}
}
