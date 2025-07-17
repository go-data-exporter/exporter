package scanner

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

type sliceRowsScanner struct {
	rows    [][]any
	columns []Column
	lastRow []any
	cursor  int
}

func FromData(rows [][]any) Rows {
	s := &sliceRowsScanner{rows: rows}
	s.columns, _ = s.Columns()
	return s
}

func (s *sliceRowsScanner) Driver() string { return "go-slice" }

func (s *sliceRowsScanner) Err() error { return nil }

func (s *sliceRowsScanner) Next() bool {
	if s.cursor >= len(s.rows) {
		return false
	}
	s.lastRow = s.rows[s.cursor]
	return true
}

func (s *sliceRowsScanner) ScanRow() ([]any, error) {
	if s.cursor >= len(s.rows) {
		return nil, io.EOF
	}
	if s.lastRow == nil {
		return nil, errors.New("tocsv: Scan called without calling Next")
	}
	if s.cursor != 0 {
		if len(s.lastRow) != len(s.columns) {
			return nil, fmt.Errorf("length of row %d != length of the first row: %d != %d", s.cursor+1, len(s.lastRow), len(s.columns))
		}
	}
	s.cursor++
	return s.lastRow, nil
}

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
