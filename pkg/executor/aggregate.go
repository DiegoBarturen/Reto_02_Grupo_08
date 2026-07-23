package executor

import (
	"fmt"
	"io"
	"strings"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

// TipoAgregacion identifica la función de agregación a aplicar.
type TipoAgregacion string

const (
	AgregacionCount TipoAgregacion = "COUNT"
	AgregacionSum   TipoAgregacion = "SUM"
	AgregacionAvg   TipoAgregacion = "AVG"
	AgregacionMin   TipoAgregacion = "MIN"
	AgregacionMax   TipoAgregacion = "MAX"
)

// columnaAgregada describe una función de agregación dentro del SELECT.
type columnaAgregada struct {
	Tipo         TipoAgregacion
	IndiceCol    int    // índice en el esquema de entrada; -1 para COUNT(*)
	NombreSalida string // nombre con el que aparece en el esquema de salida
}

// AggregateOperator implementa GROUP BY con agregaciones en el modelo Volcano.
// Materializa todas las filas del hijo, las agrupa y emite una fila por grupo.
//
// Semántica NULL:
//   - COUNT(*) cuenta todas las filas (incluyendo NULLs).
//   - COUNT(col), SUM, AVG, MIN, MAX ignoran valores NULL.
//   - Si todos los valores de una columna son NULL, SUM/AVG/MIN/MAX retornan NULL.
type AggregateOperator struct {
	hijo          Operator
	indicesGrupo  []int // índices de las columnas GROUP BY en el esquema de entrada
	agregados     []columnaAgregada
	esquema       almacenamiento.Esquema // esquema de entrada
	esquemaSalida almacenamiento.Esquema
	resultados    []Row
	cursor        int
	cargado       bool
}

// NuevoAggregateOperator construye el operador de agregación.
// Recibe el operador hijo, los nombres de columnas del GROUP BY y las columnas
// del SELECT (que pueden mezclar columnas regulares y funciones de agregación).
func NuevoAggregateOperator(
	hijo Operator,
	grupoNombres []string,
	columnasSel []parser.SelectColumn,
) (*AggregateOperator, error) {
	esquema := hijo.Schema()

	// Resolver los índices de las columnas GROUP BY.
	indicesGrupo := make([]int, len(grupoNombres))
	for i, nombre := range grupoNombres {
		idx, err := indiceDeLaColumna(esquema, nombre)
		if err != nil {
			return nil, fmt.Errorf("GROUP BY: %w", err)
		}
		indicesGrupo[i] = idx
	}

	// Construir el esquema de salida y la lista de agregados.
	var colsSalida []almacenamiento.Columna
	var agregados []columnaAgregada

	// Las columnas GROUP BY aparecen primero en la salida.
	for i := range grupoNombres {
		idx := indicesGrupo[i]
		colsSalida = append(colsSalida, almacenamiento.Columna{
			Nombre: esquema.Columnas[idx].Nombre,
			Tipo:   esquema.Columnas[idx].Tipo,
		})
	}

	// Las funciones de agregación del SELECT aparecen después.
	for _, col := range columnasSel {
		if col.Agg == nil {
			continue
		}

		agg := col.Agg
		tipoAgg := TipoAgregacion(strings.ToUpper(agg.Name))
		indiceCol := -1

		if !agg.IsStar {
			idx, err := indiceDeLaColumna(esquema, agg.Column)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", agg.Name, err)
			}
			indiceCol = idx
		}

		// Determinar el tipo de dato de la columna de salida.
		var tipoSalida almacenamiento.TipoDato
		switch tipoAgg {
		case AgregacionCount:
			tipoSalida = almacenamiento.TipoEntero
		case AgregacionSum, AgregacionAvg:
			tipoSalida = almacenamiento.TipoDecimal
		case AgregacionMin, AgregacionMax:
			if indiceCol >= 0 {
				tipoSalida = esquema.Columnas[indiceCol].Tipo
			} else {
				tipoSalida = almacenamiento.TipoEntero
			}
		}

		nombreSalida := agg.String()
		colsSalida = append(colsSalida, almacenamiento.Columna{
			Nombre: nombreSalida,
			Tipo:   tipoSalida,
		})

		agregados = append(agregados, columnaAgregada{
			Tipo:         tipoAgg,
			IndiceCol:    indiceCol,
			NombreSalida: nombreSalida,
		})
	}

	return &AggregateOperator{
		hijo:          hijo,
		indicesGrupo:  indicesGrupo,
		agregados:     agregados,
		esquema:       esquema,
		esquemaSalida: almacenamiento.Esquema{Columnas: colsSalida},
	}, nil
}

func (a *AggregateOperator) Next() (Row, error) {
	if !a.cargado {
		if err := a.materializar(); err != nil {
			return Row{}, err
		}
		a.cargado = true
	}

	if a.cursor >= len(a.resultados) {
		return Row{}, io.EOF
	}

	fila := a.resultados[a.cursor]
	a.cursor++
	return fila, nil
}

// acumuladorGrupo mantiene el estado de agregación para un grupo.
type acumuladorGrupo struct {
	filaRepresentativa Row   // fila de la que se extraen los valores GROUP BY
	countTotal         int64 // para COUNT(*)
	// Por cada agregado: estado separado
	countNoNulo []int64   // para COUNT(col)
	sumas       []float64 // para SUM y AVG (numerador)
	hasSuma     []bool    // ¿se acumuló al menos un valor no nulo?
	minVals     []almacenamiento.Valor
	hasMin      []bool
	maxVals     []almacenamiento.Valor
	hasMax      []bool
}

// materializar lee todas las filas del hijo, las agrupa y calcula los agregados.
func (a *AggregateOperator) materializar() error {
	grupos := make(map[string]*acumuladorGrupo)
	ordenClaves := make([]string, 0) // preservar orden de inserción

	for {
		fila, err := a.hijo.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		clave := a.claveGrupo(fila)

		if _, existe := grupos[clave]; !existe {
			g := &acumuladorGrupo{
				filaRepresentativa: fila,
				countNoNulo:        make([]int64, len(a.agregados)),
				sumas:              make([]float64, len(a.agregados)),
				hasSuma:            make([]bool, len(a.agregados)),
				minVals:            make([]almacenamiento.Valor, len(a.agregados)),
				hasMin:             make([]bool, len(a.agregados)),
				maxVals:            make([]almacenamiento.Valor, len(a.agregados)),
				hasMax:             make([]bool, len(a.agregados)),
			}
			grupos[clave] = g
			ordenClaves = append(ordenClaves, clave)
		}

		g := grupos[clave]
		g.countTotal++

		for i, agg := range a.agregados {
			switch agg.Tipo {
			case AgregacionCount:
				if agg.IndiceCol == -1 {
					// COUNT(*): ya se cuenta en countTotal
				} else if !fila.Datos[agg.IndiceCol].EsNulo {
					g.countNoNulo[i]++
				}

			case AgregacionSum:
				if agg.IndiceCol >= 0 && !fila.Datos[agg.IndiceCol].EsNulo {
					if f, ok := anyToFloat64(fila.Datos[agg.IndiceCol].Dato); ok {
						g.sumas[i] += f
						g.hasSuma[i] = true
					}
				}

			case AgregacionAvg:
				if agg.IndiceCol >= 0 && !fila.Datos[agg.IndiceCol].EsNulo {
					if f, ok := anyToFloat64(fila.Datos[agg.IndiceCol].Dato); ok {
						g.sumas[i] += f
						g.countNoNulo[i]++
						g.hasSuma[i] = true
					}
				}

			case AgregacionMin:
				if agg.IndiceCol >= 0 && !fila.Datos[agg.IndiceCol].EsNulo {
					v := fila.Datos[agg.IndiceCol]
					if !g.hasMin[i] {
						g.minVals[i] = v
						g.hasMin[i] = true
					} else if cmp, ok := CompararValores(v, g.minVals[i]); ok && cmp < 0 {
						g.minVals[i] = v
					}
				}

			case AgregacionMax:
				if agg.IndiceCol >= 0 && !fila.Datos[agg.IndiceCol].EsNulo {
					v := fila.Datos[agg.IndiceCol]
					if !g.hasMax[i] {
						g.maxVals[i] = v
						g.hasMax[i] = true
					} else if cmp, ok := CompararValores(v, g.maxVals[i]); ok && cmp > 0 {
						g.maxVals[i] = v
					}
				}
			}
		}
	}

	// Caso especial: sin GROUP BY y sin filas → emitir una fila con COUNT=0 / NULLs.
	if len(grupos) == 0 && len(a.indicesGrupo) == 0 {
		valores := make([]almacenamiento.Valor, len(a.agregados))
		for i, agg := range a.agregados {
			if agg.Tipo == AgregacionCount {
				valores[i] = almacenamiento.Valor{Tipo: almacenamiento.TipoEntero, Dato: int64(0)}
			} else {
				valores[i] = almacenamiento.Valor{
					Tipo:   a.esquemaSalida.Columnas[i].Tipo,
					EsNulo: true,
				}
			}
		}
		a.resultados = []Row{{Datos: valores}}
		return nil
	}

	// Construir las filas de resultado.
	a.resultados = make([]Row, 0, len(grupos))

	for _, clave := range ordenClaves {
		g := grupos[clave]
		numGrupoCols := len(a.indicesGrupo)
		numTotal := numGrupoCols + len(a.agregados)
		valores := make([]almacenamiento.Valor, numTotal)

		// Valores de las columnas GROUP BY.
		for i, idx := range a.indicesGrupo {
			valores[i] = g.filaRepresentativa.Datos[idx]
		}

		// Valores de las funciones de agregación.
		for i, agg := range a.agregados {
			colIdx := numGrupoCols + i

			switch agg.Tipo {
			case AgregacionCount:
				cnt := g.countTotal // COUNT(*)
				if agg.IndiceCol >= 0 {
					cnt = g.countNoNulo[i] // COUNT(col)
				}
				valores[colIdx] = almacenamiento.Valor{
					Tipo: almacenamiento.TipoEntero,
					Dato: cnt,
				}

			case AgregacionSum:
				if g.hasSuma[i] {
					valores[colIdx] = almacenamiento.Valor{
						Tipo: almacenamiento.TipoDecimal,
						Dato: g.sumas[i],
					}
				} else {
					valores[colIdx] = almacenamiento.Valor{
						Tipo:   almacenamiento.TipoDecimal,
						EsNulo: true,
					}
				}

			case AgregacionAvg:
				if g.hasSuma[i] && g.countNoNulo[i] > 0 {
					valores[colIdx] = almacenamiento.Valor{
						Tipo: almacenamiento.TipoDecimal,
						Dato: g.sumas[i] / float64(g.countNoNulo[i]),
					}
				} else {
					valores[colIdx] = almacenamiento.Valor{
						Tipo:   almacenamiento.TipoDecimal,
						EsNulo: true,
					}
				}

			case AgregacionMin:
				if g.hasMin[i] {
					valores[colIdx] = g.minVals[i]
				} else {
					valores[colIdx] = almacenamiento.Valor{
						Tipo:   a.esquemaSalida.Columnas[colIdx].Tipo,
						EsNulo: true,
					}
				}

			case AgregacionMax:
				if g.hasMax[i] {
					valores[colIdx] = g.maxVals[i]
				} else {
					valores[colIdx] = almacenamiento.Valor{
						Tipo:   a.esquemaSalida.Columnas[colIdx].Tipo,
						EsNulo: true,
					}
				}
			}
		}

		a.resultados = append(a.resultados, Row{Datos: valores})
	}

	return nil
}

// claveGrupo genera una clave de cadena única para el grupo al que pertenece una fila.
// Usa el separador \x01 entre valores y \x00NULL\x00 para representar NULL.
func (a *AggregateOperator) claveGrupo(fila Row) string {
	if len(a.indicesGrupo) == 0 {
		return "" // todas las filas van al mismo grupo
	}

	partes := make([]string, len(a.indicesGrupo))
	for i, idx := range a.indicesGrupo {
		v := fila.Datos[idx]
		if v.EsNulo {
			partes[i] = "\x00NULL\x00"
		} else {
			partes[i] = fmt.Sprint(v.Dato)
		}
	}
	return strings.Join(partes, "\x01")
}

// Schema devuelve el esquema de salida del operador.
func (a *AggregateOperator) Schema() almacenamiento.Esquema {
	return a.esquemaSalida
}

func (a *AggregateOperator) Close() error {
	a.resultados = nil
	a.cursor = 0
	a.cargado = false
	return a.hijo.Close()
}
