package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCSVAsText(t *testing.T) {
	tempDirectory := t.TempDir()
	csvPath := filepath.Join(tempDirectory, "estudiantes.csv")

	csvContent := `id,nombre,promedio,activo
1,Ana,17.5,true
2,Carlos,14.8,false
3,Lucia,18.2,true
`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("no se pudo crear el CSV temporal: %v", err)
	}

	table, err := LoadCSVAsText("estudiantes", csvPath)
	if err != nil {
		t.Fatalf("LoadCSVAsText devolvió un error inesperado: %v", err)
	}

	if table.Name != "estudiantes" {
		t.Errorf(
			"nombre de tabla incorrecto: se esperaba %q y se obtuvo %q",
			"estudiantes",
			table.Name,
		)
	}

	if len(table.Schema.Columns) != 4 {
		t.Errorf(
			"cantidad de columnas incorrecta: se esperaban 4 y se obtuvieron %d",
			len(table.Schema.Columns),
		)
	}

	if len(table.Rows) != 3 {
		t.Errorf(
			"cantidad de filas incorrecta: se esperaban 3 y se obtuvieron %d",
			len(table.Rows),
		)
	}

	firstValue := table.Rows[0].Data[0]

	if firstValue.Type != TypeText {
		t.Errorf(
			"tipo incorrecto: se esperaba %s y se obtuvo %s",
			TypeText,
			firstValue.Type,
		)
	}

	if firstValue.Data != "1" {
		t.Errorf(
			"valor incorrecto: se esperaba %q y se obtuvo %v",
			"1",
			firstValue.Data,
		)
	}
}

func TestLoadCSVAsTextInvalidCSV(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		content   string
	}{
		{
			name:      "nombre de tabla vacío",
			tableName: "",
			content: `id,nombre
1,Ana
`,
		},
		{
			name:      "archivo vacío",
			tableName: "estudiantes",
			content:   "",
		},
		{
			name:      "nombre de columna vacío",
			tableName: "estudiantes",
			content: `id,
1,Ana
`,
		},
		{
			name:      "columnas duplicadas",
			tableName: "estudiantes",
			content: `id,ID
1,2
`,
		},
		{
			name:      "fila con columnas faltantes",
			tableName: "estudiantes",
			content: `id,nombre,edad
1,Ana,20
2,Carlos
`,
		},
		{
			name:      "fila con columnas adicionales",
			tableName: "estudiantes",
			content: `id,nombre
1,Ana,20
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tempDirectory := t.TempDir()
			csvPath := filepath.Join(tempDirectory, "datos.csv")

			err := os.WriteFile(csvPath, []byte(test.content), 0644)
			if err != nil {
				t.Fatalf("no se pudo crear el CSV temporal: %v", err)
			}

			table, err := LoadCSVAsText(test.tableName, csvPath)

			if err == nil {
				t.Fatalf(
					"se esperaba un error, pero se obtuvo la tabla: %#v",
					table,
				)
			}

			if table != nil {
				t.Errorf(
					"se esperaba una tabla nil cuando ocurre un error",
				)
			}
		})
	}
}

func TestLoadCSVAsTextFileNotFound(t *testing.T) {
	table, err := LoadCSVAsText(
		"estudiantes",
		"archivo_que_no_existe.csv",
	)

	if err == nil {
		t.Fatal("se esperaba un error al cargar un archivo inexistente")
	}

	if table != nil {
		t.Error("se esperaba una tabla nil")
	}
}
