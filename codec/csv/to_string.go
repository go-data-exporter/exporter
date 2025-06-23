package csvcodec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-data-exporter/exporter/scanner"
)

// toString converts any value to a string representation suitable for CSV output.
//
// It handles nil values by returning the configured nullValue. Custom type conversions
// can be registered via customMapper. Built-in support includes:
//   - primitive types (bool, int, float, etc.)
//   - string and []byte (converted directly)
//   - time.Time (formatted as RFC3339Nano, nullValue for zero time)
//   - types implementing json.Marshaler or fmt.Stringer
//   - fallback to JSON marshaling or fmt.Sprintf formatting
//
// Parameters:
//   - v:      The value to convert (any type)
//   - driver: Optional driver/context identifier (used in custom mappers)
//   - column: Column metadata (name, type info, etc.)
//
// Returns:
//
// The string representation of v, or nullValue for nil/empty values.
func (cs *csvCodec) toString(v any, driver string, column scanner.Column) string {
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
