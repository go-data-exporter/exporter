// Package scanner provides implementations of the Rows interface for various data sources.
// This file defines a scanner for database/sql-compatible rows.
package scanner

import "database/sql"

// sqlRowsScanner wraps a *sql.Rows and implements the Rows interface,
// allowing codecs to consume SQL data in a generic way.
type sqlRowsScanner struct {
	*sql.Rows

	driver         string
	columns        []Column
	currentRow     []any
	currentRowPtrs []any
}

// FromSQL creates a Rows-compatible wrapper around a *sql.Rows object.
// The driver name is required for metadata and contextual information.
func FromSQL(rows *sql.Rows, driver string) Rows {
	return &sqlRowsScanner{Rows: rows, driver: driver}
}

// sqlColumn implements the Column interface using *sql.ColumnType
// provided by the standard database/sql package.
type sqlColumn struct {
	*sql.ColumnType
	index int
}

// Index returns the column's index in the result set.
func (c *sqlColumn) Index() int {
	return c.index
}

// Columns returns column metadata for the SQL result set.
// Uses database/sql ColumnTypes to provide column information.
func (s *sqlRowsScanner) Columns() ([]Column, error) {
	if s.columns != nil {
		return s.columns, nil
	}
	cc, err := s.Rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	for i, c := range cc {
		s.columns = append(s.columns, &sqlColumn{
			ColumnType: c,
			index:      i,
		})
	}
	return s.columns, nil
}

// ScanRow reads and returns the next row from the SQL result set.
// It uses pointer indirection to fill a []any with values.
func (s *sqlRowsScanner) ScanRow() ([]any, error) {
	if s.columns == nil {
		var err error
		s.columns, err = s.Columns()
		if err != nil {
			return nil, err
		}
	}
	if s.currentRow == nil {
		s.currentRow = make([]any, len(s.columns))
	}
	if s.currentRowPtrs == nil {
		s.currentRowPtrs = make([]any, len(s.columns))
	}
	for i := range len(s.columns) {
		s.currentRowPtrs[i] = &s.currentRow[i]
	}
	if err := s.Rows.Scan(s.currentRowPtrs...); err != nil {
		return nil, err
	}
	return s.currentRow, nil
}

// Driver returns the name of the SQL driver used.
func (s *sqlRowsScanner) Driver() string {
	return s.driver
}
