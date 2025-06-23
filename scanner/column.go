package scanner

import "reflect"

type Column interface {
	Name() string
	Length() (length int64, ok bool)
	DecimalSize() (precision, scale int64, ok bool)
	ScanType() reflect.Type
	Nullable() (nullable, ok bool)
	DatabaseTypeName() string
}

type mockColumn struct {
	name   string
	goType string
}

func (c *mockColumn) Name() string {
	return c.name
}

func (c *mockColumn) Length() (length int64, ok bool) {
	return 0, false
}

func (c *mockColumn) DecimalSize() (precision, scale int64, ok bool) {
	return 0, 0, false
}

func (c *mockColumn) ScanType() reflect.Type {
	return nil
}

func (c *mockColumn) Nullable() (nullable, ok bool) {
	return false, false
}

func (c *mockColumn) DatabaseTypeName() string {
	return c.goType
}
