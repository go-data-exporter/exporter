// Package jsoncodec provides a JSON implementation of the Codec interface,
// allowing tabular data to be exported in either standard JSON array format
// or newline-delimited JSON (JSON Lines). It supports per-type value mapping,
// row preprocessing, and row limits.
package jsoncodec

import (
	"io"
	"reflect"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-data-exporter/exporter/scanner"
)

// json is a high-performance JSON encoder/decoder compatible with the standard library.
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Option defines a functional configuration option for jsonCodec.
type Option func(*jsonCodec)

// jsonCodec implements the Codec interface for outputting data in JSON format.
type jsonCodec struct {
	customMapper     map[reflect.Type]func(any, scanner.Metadata) any
	preProcessorFunc func(rowID int, row map[string]any) (map[string]any, bool)
	newlineDelimited bool
	limit            int
}

// New creates a new JSON codec with the provided configuration options.
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

// WithPreProcessorFunc sets a function to transform or filter each row before writing.
// The function can modify the row contents or skip the row entirely.
func WithPreProcessorFunc(fn func(rowID int, row map[string]any) (map[string]any, bool)) Option {
	return func(c *jsonCodec) {
		c.preProcessorFunc = fn
	}
}

// WithNewlineDelimited enables newline-delimited JSON (JSON Lines) format.
func WithNewlineDelimited(isNewlineDelimited bool) Option {
	return func(c *jsonCodec) {
		c.newlineDelimited = isNewlineDelimited
	}
}

// WithCustomType registers a custom mapping function to convert a specific Go type
// to its JSON representation, using optional metadata.
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

// WithLimit sets a limit on the number of rows to export.
// A negative value disables the limit.
func WithLimit(limit int) Option {
	return func(c *jsonCodec) {
		c.limit = limit
	}
}

// Write exports the given rows to the writer in JSON format.
// The output can be either a JSON array or newline-delimited JSON.
// Supports per-row preprocessing, type conversion, and row limits.
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
