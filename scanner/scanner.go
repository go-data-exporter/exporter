// Package scanner defines interfaces and metadata used to abstract over
// tabular data sources. It allows codecs to work with various row-based
// data providers in a consistent way.
package scanner

// Rows represents an abstract data source that provides tabular data
// one row at a time. It is similar in spirit to sql.Rows but is generalized.
type Rows interface {
	// Next prepares the next row for reading. It returns false when no more rows are available.
	Next() bool

	// ScanRow returns the current row's values as a slice of interface{}.
	ScanRow() ([]any, error)

	// Columns returns metadata about the columns in the result set.
	Columns() ([]Column, error)

	// Driver returns the name of the underlying driver or data source.
	Driver() string

	// Err returns the error, if any, that was encountered during iteration.
	Err() error
}

// Metadata provides contextual information about a particular cell value,
// including its column definition, row number, and originating driver.
type Metadata struct {
	RowID  int    // The row number (starting from 1).
	Driver string // The name of the driver or data source.
	Column Column // Metadata about the column.
}
