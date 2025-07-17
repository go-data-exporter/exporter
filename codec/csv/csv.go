// Package csvcodec provides an implementation of the Codec interface
// for writing data in CSV (Comma-Separated Values) format. It supports
// custom delimiters, NULL handling, optional headers, row preprocessing,
// and type-specific string conversion.
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

// csvCodec implements the Codec interface for exporting tabular data in CSV format.
type csvCodec struct {
	customMapper     map[reflect.Type]func(any, scanner.Metadata) tostring.String
	preProcessorFunc func(rowID int, row []string) ([]string, bool)

	delimiter         rune
	useCRLF           bool
	writeHeader       bool
	writeHeaderNoData bool
	customHeader      []string

	nullValue string
	limit     int
}

// Option defines a functional option for configuring the CSV codec.
type Option func(*csvCodec)

// New creates a new CSV codec with the provided options.
func New(opts ...Option) *csvCodec {
	c := &csvCodec{
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

// WithCustomType registers a custom string conversion function for a specific Go type.
func WithCustomType[T any](fn func(v T, metadata scanner.Metadata) tostring.String) Option {
	return func(c *csvCodec) {
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

// WithPreProcessorFunc sets a function to preprocess or filter each row before writing.
// The function receives the row ID and the row values, and can return modified values or skip the row.
func WithPreProcessorFunc(fn func(rowID int, row []string) ([]string, bool)) Option {
	return func(c *csvCodec) {
		c.preProcessorFunc = fn
	}
}

// WithCustomDelimiter sets a custom delimiter for the CSV file (default is comma).
func WithCustomDelimiter(delimiter rune) Option {
	return func(c *csvCodec) {
		c.delimiter = delimiter
	}
}

// WithCRLF enables or disables CRLF line endings in the CSV output.
func WithCRLF(useCRLF bool) Option {
	return func(c *csvCodec) {
		c.useCRLF = useCRLF
	}
}

// WithHeader controls whether the CSV output should include a header row.
func WithHeader(writeHeader bool) Option {
	return func(c *csvCodec) {
		c.writeHeader = writeHeader
	}
}

// WithWriteHeaderWhenNoData controls whether a header should be written even when no data rows exist.
func WithWriteHeaderWhenNoData(writeHeaderNoData bool) Option {
	return func(c *csvCodec) {
		c.writeHeaderNoData = writeHeaderNoData
	}
}

// WithCustomHeader sets a custom header to be used instead of automatically derived column names.
func WithCustomHeader(customHeader []string) Option {
	return func(c *csvCodec) {
		c.customHeader = customHeader
	}
}

// WithCustomNULL sets the string to be used when representing NULL values in the output.
func WithCustomNULL(nullValue string) Option {
	return func(c *csvCodec) {
		c.nullValue = nullValue
	}
}

// WithLimit sets a limit on the number of rows to write. A negative value means no limit.
func WithLimit(limit int) Option {
	return func(c *csvCodec) {
		c.limit = limit
	}
}

// Write writes the scanned rows to the given writer in CSV format.
// It supports optional headers, row preprocessing, NULL conversion, and row limits.
func (c *csvCodec) Write(rows scanner.Rows, writer io.Writer) error {
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
	csvWriter := csv.NewWriter(writer)
	if c.delimiter != 0 {
		csvWriter.Comma = c.delimiter
	}
	csvWriter.UseCRLF = c.useCRLF
	defer csvWriter.Flush()

	if c.writeHeader && c.writeHeaderNoData && len(header) != 0 {
		if err = csvWriter.Write(header); err != nil {
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
				if err = csvWriter.Write(header); err != nil {
					return fmt.Errorf("failed to write headers: %w", err)
				}
			}
			if err = csvWriter.Write(row); err != nil {
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

// toString converts a single value to its string representation,
// using a custom type mapper if available, or falling back to the default converter.
// If the value is NULL, the configured nullValue is returned.
func (c *csvCodec) toString(v any, metadata scanner.Metadata) string {
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
