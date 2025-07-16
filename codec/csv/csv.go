package csvcodec

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/go-data-exporter/exporter/scanner"
	"github.com/go-data-exporter/exporter/tostring"
)

type csvCodec struct {
	customMapper      map[reflect.Type]func(any, string, scanner.Column) tostring.String
	preProcessorFunc  func(row []string) ([]string, bool)
	toStringFunc      func(v any) tostring.String
	delimiter         rune
	useCRLF           bool
	writeHeader       bool
	writeHeaderNoData bool
	customHeader      []string
	nullValue         string
}

type Option func(*csvCodec)

func New(opts ...Option) *csvCodec {
	cw := &csvCodec{
		customMapper:      make(map[reflect.Type]func(any, string, scanner.Column) tostring.String),
		delimiter:         ',',
		toStringFunc:      tostring.ToString,
		writeHeader:       true,
		writeHeaderNoData: true,
	}
	for _, opt := range opts {
		opt(cw)
	}
	return cw
}

func WithCustomType[T any](fn func(v T, driver string, column scanner.Column) tostring.String) Option {
	return func(cw *csvCodec) {
		var zero T
		typ := reflect.TypeOf(zero)
		if cw.customMapper == nil {
			cw.customMapper = make(map[reflect.Type]func(any, string, scanner.Column) tostring.String)
		}
		cw.customMapper[typ] = func(v any, driver string, column scanner.Column) tostring.String {
			return fn(v.(T), driver, column)
		}
	}
}

func WithPreProcessorFunc(fn func(row []string) ([]string, bool)) Option {
	return func(cw *csvCodec) {
		cw.preProcessorFunc = fn
	}
}

func WithCustomToStringFunc(fn func(v any) tostring.String) Option {
	return func(cw *csvCodec) {
		cw.toStringFunc = fn
	}
}

func WithCustomDelimiter(delimiter rune) Option {
	return func(cw *csvCodec) {
		cw.delimiter = delimiter
	}
}

func WithCRLF(useCRLF bool) Option {
	return func(cw *csvCodec) {
		cw.useCRLF = useCRLF
	}
}

func WithHeader(writeHeader bool) Option {
	return func(cw *csvCodec) {
		cw.writeHeader = writeHeader
	}
}

func WithWriteHeaderWhenNoData(writeHeaderNoData bool) Option {
	return func(cw *csvCodec) {
		cw.writeHeaderNoData = writeHeaderNoData
	}
}

func WithCustomHeader(customHeader []string) Option {
	return func(cw *csvCodec) {
		cw.customHeader = customHeader
	}
}

func WithCustomNULL(nullValue string) Option {
	return func(cw *csvCodec) {
		cw.nullValue = nullValue
	}
}

func (cs *csvCodec) Write(rows scanner.Rows, writer io.Writer) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	columnNames := []string{}
	for _, col := range cols {
		columnNames = append(columnNames, col.Name())
	}
	header := columnNames
	if cs.customHeader != nil {
		if len(cs.customHeader) != len(columnNames) {
			return errors.New("invalid header length")
		}
		header = cs.customHeader
	}
	csvCodec := csv.NewWriter(writer)
	if cs.delimiter != 0 {
		csvCodec.Comma = cs.delimiter
	}
	csvCodec.UseCRLF = cs.useCRLF
	defer csvCodec.Flush()

	if cs.writeHeader && cs.writeHeaderNoData && len(header) != 0 {
		if err = csvCodec.Write(header); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	for i := 0; rows.Next(); i++ {
		values, err := rows.ScanRow()
		if err != nil {
			return err
		}
		row := make([]string, len(values))
		for i := range columnNames {
			row[i] = cs.toString(values[i], rows.Driver(), cols[i])
		}
		writeRow := true
		if cs.preProcessorFunc != nil {
			row, writeRow = cs.preProcessorFunc(row)
		}
		if writeRow {
			if cs.writeHeader && i == 0 && !cs.writeHeaderNoData {
				if err = csvCodec.Write(header); err != nil {
					return fmt.Errorf("failed to write headers: %w", err)
				}
			}
			csvCodec.Write(row)
		}
	}
	return rows.Err()
}

func (cs *csvCodec) toString(v any, driver string, column scanner.Column) string {
	if v == nil {
		return cs.nullValue
	}
	if fn, ok := cs.customMapper[reflect.TypeOf(v)]; ok {
		s := fn(v, driver, column)
		if s.IsNULL {
			return cs.nullValue
		}
		return s.String
	}
	s := cs.toStringFunc(v)
	if s.IsNULL {
		return cs.nullValue
	}
	return s.String
}
