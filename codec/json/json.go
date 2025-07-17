package jsoncodec

import (
	"io"
	"reflect"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-data-exporter/exporter/scanner"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Option func(*jsonCodec)

type jsonCodec struct {
	customMapper     map[reflect.Type]func(any, scanner.Metadata) any
	preProcessorFunc func(rowID int, row map[string]any) (map[string]any, bool)
	newlineDelimited bool
	limit            int
}

func New(opts ...Option) *jsonCodec {
	c := &jsonCodec{
		customMapper: make(map[reflect.Type]func(any, scanner.Metadata) any),
		limit:        -1,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithPreProcessorFunc(fn func(rowID int, row map[string]any) (map[string]any, bool)) Option {
	return func(c *jsonCodec) {
		c.preProcessorFunc = fn
	}
}

func WithNewlineDelimited(isNewlineDelimited bool) Option {
	return func(c *jsonCodec) {
		c.newlineDelimited = isNewlineDelimited
	}
}

func WithCustomType[T any](fn func(v T, metadata scanner.Metadata) any) Option {
	return func(c *jsonCodec) {
		var zero T
		typ := reflect.TypeOf(zero)
		if c.customMapper == nil {
			c.customMapper = make(map[reflect.Type]func(any, scanner.Metadata) any)
		}
		c.customMapper[typ] = func(v any, metadata scanner.Metadata) any {
			return fn(v.(T), metadata)
		}
	}
}

func WithLimit(limit int) Option {
	return func(c *jsonCodec) {
		c.limit = limit
	}
}

func (c *jsonCodec) Write(rows scanner.Rows, writer io.Writer) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	columnNames := []string{}
	for _, col := range cols {
		columnNames = append(columnNames, col.Name())
	}
	rowID := 1
	defer func() {
		if !c.newlineDelimited && rowID != 1 {
			writer.Write([]byte("\n]\n"))
		}
	}()
	if c.limit == 0 {
		return nil
	}
	for rows.Next() {
		values, err := rows.ScanRow()
		if err != nil {
			return err
		}
		row := make(map[string]any, len(values))
		for i, col := range columnNames {
			row[col] = values[i]
			fn, ok := c.customMapper[reflect.TypeOf(values[i])]
			if ok {
				meta := scanner.Metadata{
					RowID:  rowID,
					Driver: rows.Driver(),
					Column: cols[i],
				}
				row[col] = fn(row[col], meta)
			}
		}

		writeRow := true
		if c.preProcessorFunc != nil {
			row, writeRow = c.preProcessorFunc(rowID, row)
		}
		if !writeRow {
			continue
		}

		data, err := json.Marshal(row)
		if err != nil {
			return err
		}
		if writeRow && !c.newlineDelimited && rowID == 1 {
			writer.Write([]byte("["))
		}
		if !c.newlineDelimited {
			if rowID != 1 {
				writer.Write([]byte(","))
			}
			writer.Write([]byte("\n"))
			writer.Write(data)
		} else {
			writer.Write(data)
			writer.Write([]byte("\n"))
		}
		if c.limit >= 0 && rowID >= c.limit {
			return nil
		}
		rowID++
	}
	return nil
}
