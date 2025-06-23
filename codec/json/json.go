package jsoncodec

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/go-data-exporter/exporter/scanner"
)

type Option func(*jsonCodec)

type jsonCodec struct {
	customMapper     map[reflect.Type]func(any, string, scanner.Column) any
	callback         func(row []string) ([]string, bool)
	preProcessorFunc func(row map[string]any) (map[string]any, bool)
	newlineDelimited bool
}

func New(opts ...Option) *jsonCodec {
	cw := &jsonCodec{
		customMapper: make(map[reflect.Type]func(any, string, scanner.Column) any),
	}
	for _, opt := range opts {
		opt(cw)
	}
	return cw
}

func WithPreProcessorFunc(fn func(row map[string]any) (map[string]any, bool)) Option {
	return func(cw *jsonCodec) {
		cw.preProcessorFunc = fn
	}
}

func WithNewlineDelimited(isNewlineDelimited bool) Option {
	return func(cw *jsonCodec) {
		cw.newlineDelimited = isNewlineDelimited
	}
}

func WithCustomType[T any](fn func(v T, driver string, column scanner.Column) any) Option {
	return func(cw *jsonCodec) {
		var zero T
		typ := reflect.TypeOf(zero)
		if cw.customMapper == nil {
			cw.customMapper = make(map[reflect.Type]func(any, string, scanner.Column) any)
		}
		cw.customMapper[typ] = func(v any, driver string, column scanner.Column) any {
			return fn(v.(T), driver, column)
		}
	}
}

func (cs *jsonCodec) Write(rows scanner.Rows, writer io.Writer) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	columnNames := []string{}
	for _, col := range cols {
		columnNames = append(columnNames, col.Name())
	}

	if !cs.newlineDelimited {
		writer.Write([]byte("["))
	}
	i := 0

	for rows.Next() {
		values, err := rows.ScanRow()
		if err != nil {
			return err
		}
		row := make(map[string]any, len(values))
		for i, col := range columnNames {
			row[col] = values[i]
			fn, ok := cs.customMapper[reflect.TypeOf(values[i])]
			if ok {
				row[col] = fn(row[col], rows.Driver(), cols[i])
			}
		}

		writeRow := true
		if cs.preProcessorFunc != nil {
			row, writeRow = cs.preProcessorFunc(row)
		}
		if !writeRow {
			continue
		}

		data, err := json.Marshal(row)
		if err != nil {
			return err
		}
		if !cs.newlineDelimited {
			if i != 0 {
				writer.Write([]byte(","))
			}
			writer.Write([]byte("\n"))
			writer.Write(data)
		} else {
			writer.Write(data)
			writer.Write([]byte("\n"))
		}
		i++
	}
	if !cs.newlineDelimited {
		writer.Write([]byte("\n]\n"))
	}
	return nil
}
