package almacenamiento

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func CargarCSVConInferencia(
	nombreTabla string,
	rutaArchivo string,
) (*Tabla, error) {
	nombreTabla = strings.TrimSpace(nombreTabla)

	if nombreTabla == "" {
		return nil, errors.New(
			"el nombre de la tabla no puede estar vacío",
		)
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
		return nil, errors.New(
			"el archivo CSV no contiene columnas",
		)
	}

	encabezado[0] = strings.TrimPrefix(
		encabezado[0],
		"\uFEFF",
	)

	columnas, err := construirColumnasTexto(encabezado)
	if err != nil {
		return nil, err
	}

	registros := make([][]string, 0)
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

		registros = append(registros, registro)
	}

	tipos, err := inferirTiposColumnas(
		registros,
		len(columnas),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"no se pudieron inferir los tipos del CSV: %w",
			err,
		)
	}

	for indice := range columnas {
		columnas[indice].Tipo = tipos[indice]
	}

	filas := make([]Fila, len(registros))

	for indiceFila, registro := range registros {
		valores := make([]Valor, len(registro))

		for indiceColumna, valorOriginal := range registro {
			valorConvertido, errorConversion := convertirValor(
				valorOriginal,
				tipos[indiceColumna],
			)

			if errorConversion != nil {
				return nil, fmt.Errorf(
					"fila %d, columna %q: %w",
					indiceFila+2,
					columnas[indiceColumna].Nombre,
					errorConversion,
				)
			}

			valores[indiceColumna] = valorConvertido
		}

		filas[indiceFila] = Fila{
			Datos: valores,
		}
	}

	return &Tabla{
		Nombre: nombreTabla,
		Esquema: Esquema{
			Columnas: columnas,
		},
		Filas: filas,
	}, nil
}
