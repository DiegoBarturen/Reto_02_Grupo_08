package executor

type Row struct {
	Data []any
}

type Operator interface {
	Next() (Row, error)
	Close() error
}
