package scanner

type Rows interface {
	Next() bool
	ScanRow() ([]any, error)
	Columns() ([]Column, error)
	Driver() string
	Err() error
}

type Metadata struct {
	RowID  int
	Driver string
	Column Column
}
