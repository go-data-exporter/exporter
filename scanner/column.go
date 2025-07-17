// Package scanner defines interfaces and types for column and row metadata.
// This file provides a generic Column interface and a mock implementation.
package scanner

import "reflect"

// Column describes metadata about a single column in a tabular data source.
type Column interface {
	// Index returns the zero-based index of the column.
	Index() int

	// Name returns the name of the column.
	Name() string

	// Length returns the column length (if known). If unknown, ok will be false.
	Length() (length int64, ok bool)

	// DecimalSize returns precision and scale for decimal types. If not applicable, ok will be false.
	DecimalSize() (precision, scale int64, ok bool)

	// ScanType returns the Go type of the column values.
	ScanType() reflect.Type

	// Nullable indicates whether the column may contain NULL values.
	Nullable() (nullable, ok bool)

	// DatabaseTypeName returns the database-specific type name of the column.
	DatabaseTypeName() string
}

// mockColumn is a minimal implementation of the Column interface,
// typically used for in-memory or testing purposes.
type mockColumn struct {
	index  int    // Column index
	name   string // Column name
	goType string // Type name used for DatabaseTypeName
}

// Index returns the column index.
func (c *mockColumn) Index() int {
	return c.index
}

// Name returns the column name.
func (c *mockColumn) Name() string {
	return c.name
}

// Length returns 0 and false, indicating unknown length.
func (c *mockColumn) Length() (length int64, ok bool) {
	return 0, false
}

// DecimalSize returns 0 and false, as mockColumn does not provide decimal metadata.
func (c *mockColumn) DecimalSize() (precision, scale int64, ok bool) {
	return 0, 0, false
}

// ScanType returns nil since type information is not provided in mockColumn.
func (c *mockColumn) ScanType() reflect.Type {
	return nil
}

// Nullable returns false and false, indicating nullability is unknown.
func (c *mockColumn) Nullable() (nullable, ok bool) {
	return false, false
}

// DatabaseTypeName returns the string representation of the column's Go type.
func (c *mockColumn) DatabaseTypeName() string {
	return c.goType
}
