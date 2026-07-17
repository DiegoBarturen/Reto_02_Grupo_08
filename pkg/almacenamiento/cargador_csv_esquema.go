package almacenamiento

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func CargarCSVConEsquema(
	nombreTabla string,
	rutaArchivo string,
	esquema Esquema,
) (*Tabla, error) {
	nombreTabla = strings.TrimSpace(nombreTabla)

	if nombreTabla == "" {
		return nil, errors.New(
			"el nombre de la tabla no puede estar vacío",
		)
	}

	columnas, err := validarYCopiarEsquema(esquema)
	if err != nil {
		return nil, err
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

	if err := validarEncabezadoConEsquema(
		encabezado,
		columnas,
	); err != nil {
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
			valorConvertido, errorConversion := convertirValor(
				valorOriginal,
				columnas[indiceColumna].Tipo,
			)

			if errorConversion != nil {
				return nil, fmt.Errorf(
					"fila %d, columna %q: %w",
					numeroFila,
					columnas[indiceColumna].Nombre,
					errorConversion,
				)
			}

			valores[indiceColumna] = valorConvertido
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

func validarYCopiarEsquema(
	esquema Esquema,
) ([]Columna, error) {
	if len(esquema.Columnas) == 0 {
		return nil, errors.New(
			"el esquema debe contener al menos una columna",
		)
	}

	columnas := make([]Columna, len(esquema.Columnas))
	nombresRegistrados := make(map[string]struct{})

	for indice, columna := range esquema.Columnas {
		nombreColumna := strings.TrimSpace(columna.Nombre)

		if nombreColumna == "" {
			return nil, fmt.Errorf(
				"la columna %d del esquema tiene un nombre vacío",
				indice+1,
			)
		}

		nombreNormalizado := strings.ToLower(nombreColumna)

		if _, existe := nombresRegistrados[nombreNormalizado]; existe {
			return nil, fmt.Errorf(
				"el nombre de columna %q está duplicado en el esquema",
				nombreColumna,
			)
		}

		if !esTipoSoportado(columna.Tipo) {
			return nil, fmt.Errorf(
				"la columna %q utiliza un tipo no soportado: %q",
				nombreColumna,
				columna.Tipo,
			)
		}

		nombresRegistrados[nombreNormalizado] = struct{}{}

		columnas[indice] = Columna{
			Nombre: nombreColumna,
			Tipo:   columna.Tipo,
		}
	}

	return columnas, nil
}
func esTipoSoportado(tipo TipoDato) bool {
	switch tipo {
	case TipoEntero,
		TipoDecimal,
		TipoTexto,
		TipoBooleano:
		return true

	default:
		return false
	}
}

func validarEncabezadoConEsquema(
	encabezado []string,
	columnas []Columna,
) error {
	if len(encabezado) != len(columnas) {
		return fmt.Errorf(
			"el encabezado del CSV contiene %d columnas, pero el esquema declara %d",
			len(encabezado),
			len(columnas),
		)
	}

	for indice, nombreEncabezado := range encabezado {
		nombreCSV := strings.TrimSpace(nombreEncabezado)
		nombreEsquema := columnas[indice].Nombre

		if !strings.EqualFold(nombreCSV, nombreEsquema) {
			return fmt.Errorf(
				"columna %d: el CSV contiene %q, pero el esquema declara %q",
				indice+1,
				nombreCSV,
				nombreEsquema,
			)
		}
	}

	return nil
}
