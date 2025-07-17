package htmlcodec

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/go-data-exporter/exporter/scanner"
	"github.com/go-data-exporter/exporter/tostring"
)

type htmlCodec struct {
	customMapper      map[reflect.Type]func(any, scanner.Metadata) tostring.String
	preProcessorFunc  func(rowID int, row []string) ([]string, bool)
	writeHeader       bool
	writeHeaderNoData bool
	nullValue         string
	limit             int
}

type Option func(*htmlCodec)

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

func WithPreProcessorFunc(fn func(rowID int, row []string) ([]string, bool)) Option {
	return func(c *htmlCodec) {
		c.preProcessorFunc = fn
	}
}

func WithHeader(writeHeader bool) Option {
	return func(c *htmlCodec) {
		c.writeHeader = writeHeader
	}
}

func WithCustomNULL(nullValue string) Option {
	return func(c *htmlCodec) {
		c.nullValue = nullValue
	}
}

func WithWriteHeaderWhenNoData(writeHeaderNoData bool) Option {
	return func(c *htmlCodec) {
		c.writeHeaderNoData = writeHeaderNoData
	}
}

func WithLimit(limit int) Option {
	return func(c *htmlCodec) {
		c.limit = limit
	}
}

func (c *htmlCodec) Write(rows scanner.Rows, writer io.Writer) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	if c.writeHeader && c.writeHeaderNoData && len(cols) != 0 {
		writer.Write([]byte(htmlPrefix))
		writer.Write([]byte(`<thead style="position:sticky;top:0;z-index:99;background:#f9f9f9;">`))
		for _, col := range cols {
			writer.Write(fmt.Appendf(nil, "<th><p>%s</p><p class=typ>%s</p></th>", col.Name(),
				strings.ToLower(col.DatabaseTypeName())))
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
					writer.Write(fmt.Appendf(nil, "<th><p>%s</p><p class=typ>%s</p></th>", col.Name(),
						strings.ToLower(col.DatabaseTypeName())))
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
	  scrollbar-width: none; /* Firefox */
	  -ms-overflow-style: none; /* IE и Edge */
	}
	.td::-webkit-scrollbar {
	  display: none; /* Chrome, Safari, Opera */
	}
	p.typ {
	  margin-top: 5px;
	  color: #333;
	}
	</style> </head><body><table style="width:100%;border-spacing:0px;">`), " ")
