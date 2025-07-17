// Package htmlcodec provides an HTML implementation of the Codec interface,
// generating well-formatted HTML tables with optional headers, NULL styling,
// row preprocessing, and type-specific string conversion.
package htmlcodec

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/go-data-exporter/exporter/scanner"
	"github.com/go-data-exporter/exporter/tostring"
)

// htmlCodec implements the Codec interface to export tabular data as HTML.
type htmlCodec struct {
	customMapper      map[reflect.Type]func(any, scanner.Metadata) tostring.String
	preProcessorFunc  func(rowID int, row []string) ([]string, bool)
	writeHeader       bool
	writeHeaderNoData bool

	nullValue string
	limit     int
}

// Option defines a functional configuration option for htmlCodec.
type Option func(*htmlCodec)

// New creates a new HTML codec with the provided configuration options.
func New(opts ...Option) *htmlCodec {
	c := &htmlCodec{
		customMapper:      make(map[reflect.Type]func(any, scanner.Metadata) tostring.String),
		writeHeader:       true,
		writeHeaderNoData: true,
		nullValue:         `<span style="color:#aaaaaa;">[NULL]</span>`,
		limit:             -1,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithCustomType registers a custom string conversion function for a specific Go type.
func WithCustomType[T any](fn func(v T, metadata scanner.Metadata) tostring.String) Option {
	return func(c *htmlCodec) {
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
	return func(c *htmlCodec) {
		c.preProcessorFunc = fn
	}
}

// WithHeader controls whether the HTML output should include a header row.
func WithHeader(writeHeader bool) Option {
	return func(c *htmlCodec) {
		c.writeHeader = writeHeader
	}
}

// WithCustomNULL sets the HTML string to be used for NULL values.
func WithCustomNULL(nullValue string) Option {
	return func(c *htmlCodec) {
		c.nullValue = nullValue
	}
}

// WithWriteHeaderWhenNoData controls whether a header should be written even when no data rows exist.
func WithWriteHeaderWhenNoData(writeHeaderNoData bool) Option {
	return func(c *htmlCodec) {
		c.writeHeaderNoData = writeHeaderNoData
	}
}

// WithLimit sets a limit on the number of rows to write. Negative means unlimited.
func WithLimit(limit int) Option {
	return func(c *htmlCodec) {
		c.limit = limit
	}
}

// Write writes the scanned rows as an HTML table to the provided writer.
// It supports headers, NULL styling, row limits, and optional preprocessing.
func (c *htmlCodec) Write(rows scanner.Rows, writer io.Writer) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	if c.writeHeader && c.writeHeaderNoData && len(cols) != 0 {
		writer.Write([]byte(htmlPrefix))
		writer.Write([]byte(`<thead style="position:sticky;top:0;z-index:99;background:#f9f9f9;">`))
		for _, col := range cols {
			writer.Write(fmt.Appendf(nil, "<th><p>%s</p><p class=typ>%s</p></th>",
				col.Name(), strings.ToLower(col.DatabaseTypeName())))
		}
		writer.Write([]byte(`</thead>`))
	}

	rowID := 1
	defer func() {
		if rowID != 1 {
			writer.Write([]byte(`</tbody>`))
			writer.Write([]byte(`</table></body></html>`))
		} else if c.writeHeader && c.writeHeaderNoData && len(cols) != 0 {
			writer.Write([]byte(`</table></body></html>`))
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
		row := make([]string, len(values))
		for i := range values {
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
				writer.Write([]byte(htmlPrefix))
				writer.Write([]byte(`<thead style="position:sticky;top:0;z-index:99;background:#f9f9f9;">`))
				for _, col := range cols {
					writer.Write(fmt.Appendf(nil, "<th><p>%s</p><p class=typ>%s</p></th>",
						col.Name(), strings.ToLower(col.DatabaseTypeName())))
				}
				writer.Write([]byte(`</thead>`))
			}
			if rowID == 1 {
				writer.Write([]byte(`<tbody>`))
			}
			writer.Write([]byte(`<tr>`))
			for i := range row {
				writer.Write(fmt.Appendf(nil, "<td>%s</td>", row[i]))
			}
			writer.Write([]byte(`</tr>`))
			if c.limit >= 0 && rowID >= c.limit {
				return nil
			}
			rowID++
		}
	}

	return rows.Err()
}

// toString converts a value to a string using a custom mapper if available,
// or falls back to default conversion logic. Returns nullValue if the value is considered NULL.
func (c *htmlCodec) toString(v any, metadata scanner.Metadata) string {
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

// htmlPrefix defines the beginning of the HTML document including styles and table structure.
var htmlPrefix = strings.Join(strings.Fields(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>Go Export</title><style>
	body, html {
	  margin: 0;
	  padding: 0;
	}
	* {
	  margin: 0;
	  padding: 0;
	}
	th {
	  border:1px solid #dedede;
	  padding: 15px;
	  border-top: 0px solid red;
	  border-left: 0px solid red;
	}
	td {
	  border: 1px solid #dedede;
	  border-top: 0px solid red;
	  border-left: 0px solid red;
	  padding: 10px 10px 10px 10px;
	  max-width:700px;
	  overflow-x: auto;
	  white-space: nowrap;
	  scrollbar-width: none;
	  -ms-overflow-style: none;
	}
	.td::-webkit-scrollbar {
	  display: none;
	}
	p.typ {
	  margin-top: 5px;
	  color: #333;
	}
	</style> </head><body><table style="width:100%;border-spacing:0px;">`), " ")
