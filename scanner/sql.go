package scanner

import "database/sql"

type sqlRowsScanner struct {
	*sql.Rows
	driver         string
	columns        []Column
	currentRow     []any
	currentRowPtrs []any
}

func FromSQL(rows *sql.Rows, driver string) Rows {
	return &sqlRowsScanner{Rows: rows, driver: driver}
}

func (s *sqlRowsScanner) Columns() ([]Column, error) {
	if s.columns != nil {
		return s.columns, nil
	}
	cc, err := s.Rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	for _, c := range cc {
		s.columns = append(s.columns, c)
	}
	return s.columns, nil
}

func (s *sqlRowsScanner) ScanRow() ([]any, error) {
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

func (s *sqlRowsScanner) Driver() string {
	return s.driver
}
