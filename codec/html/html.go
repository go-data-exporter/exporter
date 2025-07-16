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
	customMapper      map[reflect.Type]func(any, string, scanner.Column) tostring.String
	preProcessorFunc  func(row []string) ([]string, bool)
	toStringFunc      func(v any) tostring.String
	writeHeader       bool
	writeHeaderNoData bool
	nullValue         string
}

type Option func(*htmlCodec)

func New(opts ...Option) *htmlCodec {
	cw := &htmlCodec{
		customMapper:      make(map[reflect.Type]func(any, string, scanner.Column) tostring.String),
		writeHeader:       true,
		writeHeaderNoData: true,
		toStringFunc:      tostring.ToString,
		nullValue:         `<span style="color:#aaaaaa;">[NULL]</span>`,
	}
	for _, opt := range opts {
		opt(cw)
	}
	return cw
}

func WithCustomType[T any](fn func(v T, driver string, column scanner.Column) tostring.String) Option {
	return func(cw *htmlCodec) {
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
	return func(cw *htmlCodec) {
		cw.preProcessorFunc = fn
	}
}

func WithCustomToStringFunc(fn func(v any) tostring.String) Option {
	return func(cw *htmlCodec) {
		cw.toStringFunc = fn
	}
}

func WithHeader(writeHeader bool) Option {
	return func(cw *htmlCodec) {
		cw.writeHeader = writeHeader
	}
}

func WithCustomNULL(nullValue string) Option {
	return func(cw *htmlCodec) {
		cw.nullValue = nullValue
	}
}

func WithWriteHeaderWhenNoData(writeHeaderNoData bool) Option {
	return func(cw *htmlCodec) {
		cw.writeHeaderNoData = writeHeaderNoData
	}
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
	i := 0
	defer func() {
		if i != 0 {
			writer.Write([]byte(`</tbody>`))
			writer.Write([]byte(`</table></body></html>`))
		} else if c.writeHeader && c.writeHeaderNoData && len(cols) != 0 {
			writer.Write([]byte(`</table></body></html>`))
		}
	}()
	for ; rows.Next(); i++ {
		values, err := rows.ScanRow()
		if err != nil {
			return err
		}
		row := make([]string, len(values))
		for i := range values {
			row[i] = c.toString(values[i], rows.Driver(), cols[i])
		}
		writeRow := true
		if c.preProcessorFunc != nil {
			row, writeRow = c.preProcessorFunc(row)
		}
		if writeRow {
			if c.writeHeader && i == 0 && !c.writeHeaderNoData {
				writer.Write([]byte(htmlPrefix))
				writer.Write([]byte(`<thead style="position:sticky;top:0;z-index:99;background:#f9f9f9;">`))
				for _, col := range cols {
					writer.Write(fmt.Appendf(nil, "<th><p>%s</p><p class=typ>%s</p></th>", col.Name(),
						strings.ToLower(col.DatabaseTypeName())))
				}
				writer.Write([]byte(`</thead>`))
			}
			if i == 0 {
				writer.Write([]byte(`<tbody>`))
			}
			writer.Write([]byte(`<tr>`))
			for i := range row {
				writer.Write(fmt.Appendf(nil, "<td>%s</td>", row[i]))
			}
			writer.Write([]byte(`</tr>`))
		}
	}
	return rows.Err()
}

func (cs *htmlCodec) toString(v any, driver string, column scanner.Column) string {
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
