package executor

import (
	"io"

	"reto_02_grupo_08/pkg/almacenamiento"
)

// LimitOperator implementa la cláusula LIMIT n del modelo Volcano.
// Después de emitir n filas, retorna io.EOF independientemente de si el hijo tiene más.
type LimitOperator struct {
	hijo   Operator
	limite int
	conteo int
}

// NuevoLimitOperator construye un LimitOperator con el número máximo de filas a emitir.
func NuevoLimitOperator(hijo Operator, limite int) *LimitOperator {
	return &LimitOperator{
		hijo:   hijo,
		limite: limite,
	}
}

func (l *LimitOperator) Next() (Row, error) {
	if l.conteo >= l.limite {
		return Row{}, io.EOF
	}

	fila, err := l.hijo.Next()
	if err != nil {
		return Row{}, err
	}

	l.conteo++
	return fila, nil
}

// Schema delega al hijo: LIMIT no cambia el esquema.
func (l *LimitOperator) Schema() almacenamiento.Esquema {
	return l.hijo.Schema()
}

func (l *LimitOperator) Close() error {
	l.conteo = 0
	return l.hijo.Close()
}
