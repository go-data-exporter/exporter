package cvcodec

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/go-data-exporter/exporter/scanner"
	"github.com/go-data-exporter/exporter/tostring"
)

type cvCodec struct {
	customMapper      map[reflect.Type]func(any, scanner.Metadata) tostring.String
	preProcessorFunc  func(rowID int, row []string) ([]string, bool)
	delimiter         rune
	useCRLF           bool
	writeHeader       bool
	writeHeaderNoData bool
	customHeader      []string
	nullValue         string
	limit             int
}

type Option func(*cvCodec)

func New(opts ...Option) *cvCodec {
	c := &cvCodec{
		customMapper:      make(map[reflect.Type]func(any, scanner.Metadata) tostring.String),
		delimiter:         ',',
		writeHeader:       true,
		writeHeaderNoData: true,
		limit:             -1,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithCustomType[T any](fn func(v T, metadata scanner.Metadata) tostring.String) Option {
	return func(c *cvCodec) {
		var zero T
		typ := reflect.TypeOf(zero)
		if c.customMapper == nil {
			c.customMapper = make(map[reflect.Type]func(any, scanner.Metadata) tostring.String)
		}
		c.customMapper[typ] = func(v any, metadata scanner.Metadata) tostring.String {
			return fn(v.(T), metadata)
		}
	}
}

func WithPreProcessorFunc(fn func(rowID int, row []string) ([]string, bool)) Option {
	return func(c *cvCodec) {
		c.preProcessorFunc = fn
	}
}

func WithCustomDelimiter(delimiter rune) Option {
	return func(c *cvCodec) {
		c.delimiter = delimiter
	}
}

func WithCRLF(useCRLF bool) Option {
	return func(c *cvCodec) {
		c.useCRLF = useCRLF
	}
}

func WithHeader(writeHeader bool) Option {
	return func(c *cvCodec) {
		c.writeHeader = writeHeader
	}
}

func WithWriteHeaderWhenNoData(writeHeaderNoData bool) Option {
	return func(c *cvCodec) {
		c.writeHeaderNoData = writeHeaderNoData
	}
}

func WithCustomHeader(customHeader []string) Option {
	return func(c *cvCodec) {
		c.customHeader = customHeader
	}
}

func WithCustomNULL(nullValue string) Option {
	return func(c *cvCodec) {
		c.nullValue = nullValue
	}
}

func WithLimit(limit int) Option {
	return func(c *cvCodec) {
		c.limit = limit
	}
}

func (c *cvCodec) Write(rows scanner.Rows, writer io.Writer) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	columnNames := []string{}
	for _, col := range cols {
		columnNames = append(columnNames, col.Name())
	}
	header := columnNames
	if c.customHeader != nil {
		if len(c.customHeader) != len(columnNames) {
			return errors.New("invalid header length")
		}
		header = c.customHeader
	}
	cvCodec := csv.NewWriter(writer)
	if c.delimiter != 0 {
		cvCodec.Comma = c.delimiter
	}
	cvCodec.UseCRLF = c.useCRLF
	defer cvCodec.Flush()

	if c.writeHeader && c.writeHeaderNoData && len(header) != 0 {
		if err = cvCodec.Write(header); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}
	if c.limit == 0 {
		return nil
	}
	rowID := 1
	for rows.Next() {
		values, err := rows.ScanRow()
		if err != nil {
			return err
		}
		row := make([]string, len(values))
		for i := range columnNames {
			meta := scanner.Metadata{
				RowID:  rowID,
				Driver: rows.Driver(),
				Column: cols[i],
			}
			row[i] = c.toString(values[i], meta)
		}
		writeRow := true
		if c.preProcessorFunc != nil {
			row, writeRow = c.preProcessorFunc(rowID, row)
		}
		if writeRow {
			if c.writeHeader && rowID == 1 && !c.writeHeaderNoData {
				if err = cvCodec.Write(header); err != nil {
					return fmt.Errorf("failed to write headers: %w", err)
				}
			}
			if err = cvCodec.Write(row); err != nil {
				return fmt.Errorf("could not write %d row: %s", rowID, err.Error())
			}
			if c.limit >= 0 && rowID >= c.limit {
				return nil
			}
			rowID++
		}
	}
	return rows.Err()
}

func (c *cvCodec) toString(v any, metadata scanner.Metadata) string {
	if v == nil {
		return c.nullValue
	}
	if fn, ok := c.customMapper[reflect.TypeOf(v)]; ok {
		s := fn(v, metadata)
		if s.IsNULL {
			return c.nullValue
		}
		return s.String
	}
	s := tostring.ToString(v)
	if s.IsNULL {
		return c.nullValue
	}
	return s.String
}
