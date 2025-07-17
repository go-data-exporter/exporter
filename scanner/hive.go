// Package scanner provides implementations of the Rows interface for various data sources.
// This file defines a scanner for Apache Hive using the gohive library.
package scanner

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-data-exporter/gohive"
)

// hiveRowsScanner implements the Rows interface for Apache Hive,
// using a gohive.Cursor to read tabular data row by row.
type hiveRowsScanner struct {
	cursor         *gohive.Cursor
	ctx            context.Context
	columns        []Column
	currentRow     []any
	currentRowPtrs []any
}

// FromHiveCursor wraps a gohive.Cursor and returns a Rows-compatible scanner.
// The context is used for cancellation and timeout control.
func FromHiveCursor(cursor *gohive.Cursor, ctx context.Context) Rows {
	return &hiveRowsScanner{cursor: cursor, ctx: ctx}
}

// Next advances the cursor to the next row, returning true if another row is available.
func (h *hiveRowsScanner) Next() bool {
	return h.cursor.HasMore(h.ctx)
}

// ScanRow reads the current row of data from the Hive cursor.
// It returns the row as a slice of values.
func (h *hiveRowsScanner) ScanRow() ([]any, error) {
	if h.currentRow == nil {
		h.currentRow = make([]any, len(h.columns))
	}
	if h.currentRowPtrs == nil {
		h.currentRowPtrs = make([]any, len(h.columns))
	}
	for i := range len(h.columns) {
		h.currentRowPtrs[i] = &h.currentRow[i]
	}

	h.currentRow = h.cursor.RowSlice(h.ctx)
	if h.cursor.Err != nil {
		return nil, h.cursor.Err
	}
	return h.currentRow, nil
}

// Columns retrieves metadata about the result set's columns from the Hive cursor.
func (h *hiveRowsScanner) Columns() ([]Column, error) {
	if h.columns != nil {
		return h.columns, nil
	}
	cc := h.cursor.Description()
	for i, c := range cc {
		if len(c) == 0 {
			continue
		}
		var col hiveColumn
		if len(c) == 1 {
			col.name = c[0]
		} else if len(c) == 2 {
			col.name = c[0]
			col.hiveType = c[1]
		}
		_, colName, ok := strings.Cut(col.name, ".")
		if ok {
			col.name = colName
		}
		col.hiveType = strings.TrimSuffix(col.hiveType, "_TYPE")
		col.index = i
		h.columns = append(h.columns, &col)
	}
	return h.columns, nil
}

// Driver returns the name of the data source, which is "gohive" in this case.
func (h *hiveRowsScanner) Driver() string {
	return "gohive"
}

// Err returns any error encountered while iterating rows.
func (h *hiveRowsScanner) Err() error {
	return h.cursor.Error()
}

// hiveColumn represents metadata about a Hive column.
type hiveColumn struct {
	index    int
	name     string
	hiveType string
}

// Index returns the zero-based column index.
func (c *hiveColumn) Index() int {
	return c.index
}

// Name returns the column name.
func (c *hiveColumn) Name() string {
	return c.name
}

// Length returns 0 and false, as Hive columns do not report length.
func (c *hiveColumn) Length() (length int64, ok bool) {
	return 0, false
}

// DecimalSize returns 0 and false, as decimal precision is not known.
func (c *hiveColumn) DecimalSize() (precision, scale int64, ok bool) {
	return 0, 0, false
}

// ScanType returns nil, as Hive columns do not expose Go types directly.
func (c *hiveColumn) ScanType() reflect.Type {
	return nil
}

// Nullable returns false and false, as Hive does not expose nullability metadata.
func (c *hiveColumn) Nullable() (nullable, ok bool) {
	return false, false
}

// DatabaseTypeName returns the Hive-specific type name for the column.
func (c *hiveColumn) DatabaseTypeName() string {
	return c.hiveType
}
