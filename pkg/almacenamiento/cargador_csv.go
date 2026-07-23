package almacenamiento

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func CargarCSVComoTexto(nombreTabla, rutaArchivo string) (*Tabla, error) {
	nombreTabla = strings.TrimSpace(nombreTabla)

	if nombreTabla == "" {
		return nil, errors.New("el nombre de la tabla no puede estar vacío")
	}

	archivo, err := os.Open(rutaArchivo)
	if err != nil {
		return nil, fmt.Errorf(
			"no se pudo abrir el archivo CSV %q: %w",
			rutaArchivo,
			err,
		)
	}
	defer archivo.Close()

	lector := csv.NewReader(archivo)

	lector.FieldsPerRecord = -1
	lector.TrimLeadingSpace = true

	encabezado, err := lector.Read()

	if errors.Is(err, io.EOF) {
		return nil, errors.New("el archivo CSV está vacío")
	}

	if err != nil {
		return nil, fmt.Errorf(
			"no se pudo leer el encabezado del archivo CSV: %w",
			err,
		)
	}

	if len(encabezado) == 0 {
		return nil, errors.New("el archivo CSV no contiene columnas")
	}

	encabezado[0] = strings.TrimPrefix(encabezado[0], "\uFEFF")

	columnas, err := construirColumnasTexto(encabezado)
	if err != nil {
		return nil, err
	}

	filas := make([]Fila, 0)
	numeroFila := 1

	for {
		registro, errorLectura := lector.Read()

		if errors.Is(errorLectura, io.EOF) {
			break
		}

		numeroFila++

		if errorLectura != nil {
			return nil, fmt.Errorf(
				"no se pudo leer la fila %d: %w",
				numeroFila,
				errorLectura,
			)
		}

		if len(registro) != len(columnas) {
			return nil, fmt.Errorf(
				"fila %d: se esperaban %d columnas, pero se encontraron %d",
				numeroFila,
				len(columnas),
				len(registro),
			)
		}

		valores := make([]Valor, len(registro))

		for indiceColumna, valorOriginal := range registro {
			valores[indiceColumna] = Valor{
				Tipo:   TipoTexto,
				Dato:   valorOriginal,
				EsNulo: false,
			}
		}

		filas = append(filas, Fila{
			Datos: valores,
		})
	}

	return &Tabla{
		Nombre: nombreTabla,
		Esquema: Esquema{
			Columnas: columnas,
		},
		Filas: filas,
	}, nil
}

// construirColumnasTexto crea las columnas iniciales de un archivo CSV.
//
// También verifica que no existan nombres vacíos ni nombres duplicados.
func construirColumnasTexto(encabezado []string) ([]Columna, error) {
	columnas := make([]Columna, len(encabezado))
	nombresRegistrados := make(map[string]struct{})

	for indice, nombreOriginal := range encabezado {
		nombreColumna := strings.TrimSpace(nombreOriginal)

		if nombreColumna == "" {
			return nil, fmt.Errorf(
				"la columna %d tiene un nombre vacío",
				indice+1,
			)
		}

		nombreNormalizado := strings.ToLower(nombreColumna)

		if _, existe := nombresRegistrados[nombreNormalizado]; existe {
			return nil, fmt.Errorf(
				"el nombre de columna %q está duplicado",
				nombreColumna,
			)
		}

		nombresRegistrados[nombreNormalizado] = struct{}{}

		columnas[indice] = Columna{
			Nombre: nombreColumna,
			Tipo:   TipoTexto,
		}
	}

	return columnas, nil
}
