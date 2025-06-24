package htmlcodec

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-data-exporter/exporter/scanner"
)

type htmlCodec struct {
	customMapper      map[reflect.Type]func(any, string, scanner.Column) string
	preProcessorFunc  func(row []string) ([]string, bool)
	writeHeader       bool
	writeHeaderNoData bool
	nullValue         string
}

type Option func(*htmlCodec)

func New(opts ...Option) *htmlCodec {
	cw := &htmlCodec{
		customMapper:      make(map[reflect.Type]func(any, string, scanner.Column) string),
		writeHeader:       true,
		writeHeaderNoData: true,
	}
	for _, opt := range opts {
		opt(cw)
	}
	return cw
}

func WithCustomType[T any](fn func(v T, driver string, column scanner.Column) string) Option {
	return func(cw *htmlCodec) {
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
	if i != 0 {
		writer.Write([]byte(`</tbody>`))
		writer.Write([]byte(`</table></body></html>`))
	} else if c.writeHeader && c.writeHeaderNoData && len(cols) != 0 {
		writer.Write([]byte(`</table></body></html>`))
	}
	return rows.Err()
}

func (cs *htmlCodec) toString(v any, driver string, column scanner.Column) string {
	if v == nil {
		return cs.nullValue
	}
	if fn, ok := cs.customMapper[reflect.TypeOf(v)]; ok {
		return fn(v, driver, column)
	}
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case time.Time:
		if v.IsZero() {
			return cs.nullValue
		}
		return v.Format(time.RFC3339Nano)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	if jsonMarshaler, ok := v.(json.Marshaler); ok {
		if jsonData, err := jsonMarshaler.MarshalJSON(); err == nil {
			s := strings.Trim(string(jsonData), `"`)
			if s == "[]" || s == "{}" || s == "null" {
				return cs.nullValue
			}
			return s
		}
	}
	if fmtStringer, ok := v.(fmt.Stringer); ok {
		return fmtStringer.String()
	}
	if jsonData, err := json.Marshal(v); err == nil {
		s := strings.Trim(string(jsonData), `"`)
		if s == "[]" || s == "{}" || s == "null" {
			return cs.nullValue
		}
		return s
	}
	return fmt.Sprintf("%v", v)
}

func WithPreProcessorFunc(fn func(row []string) ([]string, bool)) Option {
	return func(cw *htmlCodec) {
		cw.preProcessorFunc = fn
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
