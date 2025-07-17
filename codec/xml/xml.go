// Package xmlcodec provides an XML implementation of the Codec interface,
// generating well-formatted XML tables with optional headers, NULL styling,
// row preprocessing, and type-specific string conversion.
package xmlcodec

import (
	"encoding/xml"
	"io"
	"reflect"

	"github.com/go-data-exporter/exporter/scanner"
	"github.com/go-data-exporter/exporter/tostring"
)

// xmlCodec implements the Codec interface to export tabular data as XML.
type xmlCodec struct {
	customMapper     map[reflect.Type]func(any, scanner.Metadata) tostring.String
	preProcessorFunc func(rowID int, row []string) ([]string, bool)
	limit            int
}

// Option defines a functional configuration option for xmlCodec.
type Option func(*xmlCodec)

// New creates a new XML codec with the provided configuration options.
func New(opts ...Option) *xmlCodec {
	c := &xmlCodec{
		customMapper: make(map[reflect.Type]func(any, scanner.Metadata) tostring.String),
		limit:        -1,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithCustomType registers a custom string conversion function for a specific Go type.
func WithCustomType[T any](fn func(v T, metadata scanner.Metadata) tostring.String) Option {
	return func(c *xmlCodec) {
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
func WithPreProcessorFunc(fn func(rowID int, row []string) ([]string, bool)) Option {
	return func(c *xmlCodec) {
		c.preProcessorFunc = fn
	}
}

// WithLimit sets a limit on the number of rows to write. Negative means unlimited.
func WithLimit(limit int) Option {
	return func(c *xmlCodec) {
		c.limit = limit
	}
}

// Write writes the scanned rows as an XML table to the provided writer.
// It supports headers, NULL styling, row limits, and optional preprocessing.
func (c *xmlCodec) Write(rows scanner.Rows, writer io.Writer) error {
	if c.limit == 0 {
		return nil
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	rowID := 0
	defer func() {
		if rowID > 0 {
			writer.Write([]byte("</data>\n"))
		}
	}()
	for rows.Next() {
		values, err := rows.ScanRow()
		if err != nil {
			return err
		}
		row := make([]string, len(values))
		for i := range values {
			meta := scanner.Metadata{
				RowID:  rowID + 1,
				Driver: rows.Driver(),
				Column: cols[i],
			}
			s := c.toString(values[i], meta)
			if s.IsNULL {
				values[i] = nil
			}
			row[i] = s.String
		}

		writeRow := true
		if c.preProcessorFunc != nil {
			row, writeRow = c.preProcessorFunc(rowID+1, row)
		}
		if !writeRow {
			continue
		}
		if rowID == 0 {
			writer.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
			writer.Write([]byte("\n<data>\n"))
		}
		writer.Write([]byte("<row>"))
		for i := range row {
			if values[i] == nil {
				continue
			}
			colName := cols[i].Name()
			writer.Write([]byte("<" + colName + ">"))
			xml.EscapeText(writer, []byte(row[i]))
			writer.Write([]byte("</" + colName + ">"))

		}
		writer.Write([]byte("</row>\n"))
		rowID++
		if c.limit >= 0 && rowID >= c.limit {
			return nil
		}
	}

	return rows.Err()
}

// toString converts a value to a string using a custom mapper if available,
// or falls back to default conversion logic. Returns nullValue if the value is considered NULL.
func (c *xmlCodec) toString(v any, metadata scanner.Metadata) tostring.String {
	if v == nil {
		return tostring.String{IsNULL: true}
	}
	if fn, ok := c.customMapper[reflect.TypeOf(v)]; ok {
		return fn(v, metadata)
	}
	return tostring.ToString(v)
}
