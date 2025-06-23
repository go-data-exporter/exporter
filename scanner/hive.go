package scanner

import (
	"context"
	"reflect"
	"strings"

	"github.com/beltran/gohive"
)

type hiveRowsScanner struct {
	cursor         *gohive.Cursor
	ctx            context.Context
	columns        []Column
	currentRow     []any
	currentRowPtrs []any
}

func FromHiveCursor(cursor *gohive.Cursor, ctx context.Context) Rows {
	return &hiveRowsScanner{cursor: cursor, ctx: ctx}
}

func (h *hiveRowsScanner) Next() bool {
	return h.cursor.HasMore(h.ctx)
}

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
	h.cursor.FetchOne(h.ctx, h.currentRowPtrs...)
	if h.cursor.Err != nil {
		return nil, h.cursor.Err
	}
	return h.currentRow, nil
}

func (h *hiveRowsScanner) Columns() ([]Column, error) {
	if h.columns != nil {
		return h.columns, nil
	}
	cc := h.cursor.Description()
	for _, c := range cc {
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
		h.columns = append(h.columns, &col)
	}
	return h.columns, nil
}

func (h *hiveRowsScanner) Driver() string {
	return "gohive"
}

func (h *hiveRowsScanner) Err() error {
	return h.cursor.Error()
}

type hiveColumn struct {
	name     string
	hiveType string
}

func (c *hiveColumn) Name() string {
	return c.name
}

func (c *hiveColumn) Length() (length int64, ok bool) {
	return 0, false
}

func (c *hiveColumn) DecimalSize() (precision, scale int64, ok bool) {
	return 0, 0, false
}

func (c *hiveColumn) ScanType() reflect.Type {
	return nil
}

func (c *hiveColumn) Nullable() (nullable, ok bool) {
	return false, false
}

func (c *hiveColumn) DatabaseTypeName() string {
	return c.hiveType
}
