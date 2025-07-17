// Package scanner defines interfaces and implementations for reading tabular data.
// This file provides an in-memory implementation of Rows backed by a slice of rows.
package scanner

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

// sliceRowsScanner implements the Rows interface using a slice of slices.
// It is useful for testing or small in-memory data sources.
type sliceRowsScanner struct {
	rows    [][]any  // The raw data: each inner slice is a row.
	columns []Column // Derived column metadata.
	lastRow []any    // The last read row, cached after Next().
	cursor  int      // The index of the current row.
}

// FromData creates a new Rows scanner from a 2D slice of data.
// Each inner slice represents a row. Column metadata is inferred from the first row.
func FromData(rows [][]any) Rows {
	s := &sliceRowsScanner{rows: rows}
	s.columns, _ = s.Columns()
	return s
}

// Driver returns a string identifying the data source as an in-memory slice.
func (s *sliceRowsScanner) Driver() string {
	return "go-slice"
}

// Err always returns nil for sliceRowsScanner since errors are handled immediately.
func (s *sliceRowsScanner) Err() error {
	return nil
}

// Next prepares the next row for reading. Returns false when no more rows are available.
func (s *sliceRowsScanner) Next() bool {
	if s.cursor >= len(s.rows) {
		return false
	}
	s.lastRow = s.rows[s.cursor]
	return true
}

// ScanRow returns the current row's data.
// It must be called only after a successful call to Next().
func (s *sliceRowsScanner) ScanRow() ([]any, error) {
	if s.cursor >= len(s.rows) {
		return nil, io.EOF
	}
	if s.lastRow == nil {
		return nil, errors.New("go-data-exporter: scan called without calling Next")
	}
	if s.cursor != 0 {
		if len(s.lastRow) != len(s.columns) {
			return nil, fmt.Errorf("length of row %d != length of the first row: %d != %d", s.cursor+1, len(s.lastRow), len(s.columns))
		}
	}
	s.cursor++
	return s.lastRow, nil
}

// Columns returns the inferred column metadata, based on the first row.
// If no data is available, returns an empty slice.
func (s *sliceRowsScanner) Columns() ([]Column, error) {
	if s.columns != nil {
		return s.columns, nil
	}
	if len(s.rows) != 0 {
		for i, v := range s.rows[0] {
			c := &mockColumn{
				index: i,
				name:  fmt.Sprintf("column_%d", i),
			}
			if v == nil {
				c.goType = "nil"
			} else {
				c.goType = reflect.TypeOf(v).String()
			}
			s.columns = append(s.columns, c)
		}
	}
	return s.columns, nil
}
