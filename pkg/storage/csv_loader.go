package storage

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadCSVAsText(tableName, filePath string) (*Table, error) {
	tableName = strings.TrimSpace(tableName)

	if tableName == "" {
		return nil, errors.New("el nombre de la tabla no puede estar vacío")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el archivo CSV %q: %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, errors.New("el archivo CSV está vacío")
	}

	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el encabezado del CSV: %w", err)
	}

	if len(header) == 0 {
		return nil, errors.New("el archivo CSV no contiene columnas")
	}

	header[0] = strings.TrimPrefix(header[0], "\uFEFF")

	columns, err := buildTextColumns(header)
	if err != nil {
		return nil, err
	}

	rows := make([]Row, 0)
	recordNumber := 1

	for {
		record, readErr := reader.Read()

		if errors.Is(readErr, io.EOF) {
			break
		}

		recordNumber++

		if readErr != nil {
			return nil, fmt.Errorf(
				"no se pudo leer la fila %d: %w",
				recordNumber,
				readErr,
			)
		}

		if len(record) != len(columns) {
			return nil, fmt.Errorf(
				"fila %d: se esperaban %d columnas, pero se encontraron %d",
				recordNumber,
				len(columns),
				len(record),
			)
		}

		values := make([]Value, len(record))

		for columnIndex, rawValue := range record {
			values[columnIndex] = Value{
				Type:   TypeText,
				Data:   rawValue,
				IsNull: false,
			}
		}

		rows = append(rows, Row{
			Data: values,
		})
	}

	return &Table{
		Name: tableName,
		Schema: Schema{
			Columns: columns,
		},
		Rows: rows,
	}, nil
}

func buildTextColumns(header []string) ([]Column, error) {
	columns := make([]Column, len(header))
	seenNames := make(map[string]struct{})

	for index, rawName := range header {
		columnName := strings.TrimSpace(rawName)

		if columnName == "" {
			return nil, fmt.Errorf(
				"la columna %d tiene un nombre vacío",
				index+1,
			)
		}

		normalizedName := strings.ToLower(columnName)

		if _, exists := seenNames[normalizedName]; exists {
			return nil, fmt.Errorf(
				"el nombre de columna %q está duplicado",
				columnName,
			)
		}

		seenNames[normalizedName] = struct{}{}

		columns[index] = Column{
			Name: columnName,
			Type: TypeText,
		}
	}

	return columns, nil
}
