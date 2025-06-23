package csvcodec

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/go-data-exporter/exporter/scanner"
)

type csvCodec struct {
	customMapper     map[reflect.Type]func(any, string, scanner.Column) string
	preProcessorFunc func(row []string) ([]string, bool)
	delimiter        rune
	useCRLF          bool
	writeHeader      bool
	customHeader     []string
	nullValue        string
}

type Option func(*csvCodec)

func New(opts ...Option) *csvCodec {
	cw := &csvCodec{
		customMapper: make(map[reflect.Type]func(any, string, scanner.Column) string),
		delimiter:    ',',
		useCRLF:      false,
		writeHeader:  true,
	}
	for _, opt := range opts {
		opt(cw)
	}
	return cw
}

func WithCustomType[T any](fn func(v T, driver string, column scanner.Column) string) Option {
	return func(cw *csvCodec) {
		var zero T
		typ := reflect.TypeOf(zero)
		if cw.customMapper == nil {
			cw.customMapper = make(map[reflect.Type]func(any, string, scanner.Column) string)
		}
		cw.customMapper[typ] = func(v any, driver string, column scanner.Column) string {
			return fn(v.(T), driver, column)
		}
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

	if cs.writeHeader {
		if err = csvCodec.Write(header); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}
	for rows.Next() {
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
			csvCodec.Write(row)
		}
	}
	return rows.Err()
}

func WithPreProcessorFunc(fn func(row []string) ([]string, bool)) Option {
	return func(cw *csvCodec) {
		cw.preProcessorFunc = fn
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
