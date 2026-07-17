package storage

type DataType string

const (
	TypeInteger DataType = "INTEGER"
	TypeDecimal DataType = "DECIMAL"
	TypeText    DataType = "TEXT"
	TypeBoolean DataType = "BOOLEAN"
)

type Value struct {
	Type   DataType
	Data   any
	IsNull bool
}

type Column struct {
	Name string
	Type DataType
}

type Schema struct {
	Columns []Column
}

type Row struct {
	Data []Value
}

type Table struct {
	Name   string
	Schema Schema
	Rows   []Row
}
