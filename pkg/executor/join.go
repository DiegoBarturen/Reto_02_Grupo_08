package executor

import (
	"fmt"
	"io"
	"strings"
	"time"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

// JoinOperator ejecuta un INNER JOIN ... ON con dos algoritmos soportados:
// - nested-loop
// - hash join
//
// La salida se materializa en memoria para permitir comparar tiempos y
// mantener el flujo del REPL sin cambiar la semántica del resto del motor.
type JoinOperator struct {
	leftRows        []Row
	rightRows       []Row
	leftSchema      almacenamiento.Esquema
	rightSchema     almacenamiento.Esquema
	condicion       parser.Expr
	algoritmo       string
	filasSalida     []Row
	outputCursor    int
	esquemaSalida   almacenamiento.Esquema
	hashIndex       map[string][]Row
	joinTimeMeasure time.Duration
}

func NuevoJoinOperator(hijo Operator, leftName string, rightTabla *almacenamiento.Tabla, condicion parser.Expr, algoritmo string) (*JoinOperator, error) {
	if rightTabla == nil {
		return nil, fmt.Errorf("la tabla derecha del JOIN no puede ser nil")
	}
	if algoritmo != "nested" && algoritmo != "hash" {
		return nil, fmt.Errorf("algoritmo de JOIN no soportado: %q", algoritmo)
	}

	leftRows, err := recolectarRows(hijo)
	if err != nil {
		return nil, err
	}

	rightRows := make([]Row, len(rightTabla.Filas))
	copy(rightRows, rightTabla.Filas)

	join := &JoinOperator{
		leftRows:      leftRows,
		rightRows:     rightRows,
		leftSchema:    hijo.Schema(),
		rightSchema:   rightTabla.Esquema,
		condicion:     condicion,
		algoritmo:     algoritmo,
		esquemaSalida: combinarEsquemasPrefijados(hijo.Schema(), leftName, rightTabla.Esquema, rightTabla.Nombre),
	}

	inicio := time.Now()
	if algoritmo == "hash" {
		join.filasSalida, err = join.ejecutarHashJoin()
	} else {
		join.filasSalida, err = join.ejecutarNestedLoopJoin()
	}
	join.joinTimeMeasure = time.Since(inicio)
	if err != nil {
		return nil, err
	}

	return join, nil
}

func (j *JoinOperator) Next() (Row, error) {
	if j.outputCursor >= len(j.filasSalida) {
		return Row{}, io.EOF
	}

	fila := j.filasSalida[j.outputCursor]
	j.outputCursor++
	return fila, nil
}

func (j *JoinOperator) Close() error {
	j.outputCursor = 0
	return nil
}

func (j *JoinOperator) Schema() almacenamiento.Esquema {
	return j.esquemaSalida
}

func (j *JoinOperator) Duracion() time.Duration {
	return j.joinTimeMeasure
}

func (j *JoinOperator) ejecutarNestedLoopJoin() ([]Row, error) {
	resultado := make([]Row, 0)
	for _, leftRow := range j.leftRows {
		for _, rightRow := range j.rightRows {
			filaCombinada := combinarFilas(leftRow, rightRow)
			if evaluarCondicionJoin(filaCombinada, j.esquemaSalida, j.condicion) {
				resultado = append(resultado, filaCombinada)
			}
		}
	}
	return resultado, nil
}

func (j *JoinOperator) ejecutarHashJoin() ([]Row, error) {
	binario, ok := j.condicion.(*parser.BinaryExpr)
	if !ok || binario.Operator != "=" {
		return nil, fmt.Errorf("el hash join solo soporta comparaciones de igualdad en ON")
	}

	j.hashIndex = make(map[string][]Row)
	for _, rightRow := range j.rightRows {
		valor, err := Evaluar(rightRow, j.rightSchema, binario.Right)
		if err != nil {
			return nil, err
		}
		clave := claveJoin(valor)
		j.hashIndex[clave] = append(j.hashIndex[clave], rightRow)
	}

	resultado := make([]Row, 0)
	for _, leftRow := range j.leftRows {
		valor, err := Evaluar(leftRow, j.leftSchema, binario.Left)
		if err != nil {
			return nil, err
		}
		clave := claveJoin(valor)
		for _, rightRow := range j.hashIndex[clave] {
			filaCombinada := combinarFilas(leftRow, rightRow)
			if evaluarCondicionJoin(filaCombinada, j.esquemaSalida, j.condicion) {
				resultado = append(resultado, filaCombinada)
			}
		}
	}

	return resultado, nil
}

func claveJoin(valor any) string {
	if valor == nil {
		return "NULL"
	}
	return strings.ToLower(fmt.Sprint(valor))
}

func combinarFilas(left, right Row) Row {
	datos := make([]almacenamiento.Valor, 0, len(left.Datos)+len(right.Datos))
	datos = append(datos, left.Datos...)
	datos = append(datos, right.Datos...)
	return Row{Datos: datos}
}

func combinarEsquemasPrefijados(leftEsquema almacenamiento.Esquema, leftName string, rightEsquema almacenamiento.Esquema, rightName string) almacenamiento.Esquema {
	columnas := make([]almacenamiento.Columna, 0, len(leftEsquema.Columnas)+len(rightEsquema.Columnas))
	for _, col := range leftEsquema.Columnas {
		columnas = append(columnas, almacenamiento.Columna{Nombre: fmt.Sprintf("%s.%s", leftName, col.Nombre), Tipo: col.Tipo})
	}
	for _, col := range rightEsquema.Columnas {
		columnas = append(columnas, almacenamiento.Columna{Nombre: fmt.Sprintf("%s.%s", rightName, col.Nombre), Tipo: col.Tipo})
	}
	return almacenamiento.Esquema{Columnas: columnas}
}

func evaluarCondicionJoin(fila Row, esquema almacenamiento.Esquema, condicion parser.Expr) bool {
	resultado, err := Evaluar(fila, esquema, condicion)
	if err != nil {
		return false
	}
	if valor, ok := resultado.(bool); ok {
		return valor
	}
	return false
}

func recolectarRows(op Operator) ([]Row, error) {
	filas := make([]Row, 0)
	for {
		fila, err := op.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		filas = append(filas, fila)
	}
	return filas, nil
}
