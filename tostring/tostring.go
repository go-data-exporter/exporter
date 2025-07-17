// Package tostring provides functionality to convert arbitrary Go values
// into their string representation, while also detecting NULL or zero-equivalent values.
// It is primarily used for consistent string serialization in data export scenarios.
package tostring

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// jsonStd is a high-performance JSON encoder/decoder compatible with the standard library.
var jsonStd = jsoniter.ConfigCompatibleWithStandardLibrary

// String represents a string value along with a flag indicating whether it was NULL.
// If IsNULL is true, then the value should be considered as NULL or absent.
type String struct {
	String string
	IsNULL bool
}

// ToString converts an arbitrary value to a String type, which contains
// a string representation of the value and a flag indicating if the value was NULL.
//
// The conversion logic supports common Go primitive types, slices, time.Time,
// and types implementing json.Marshaler or fmt.Stringer interfaces.
//
// If the input is nil or represents an empty/null value (like zero time,
// "null", "[]", or "{}" in JSON), the result will have IsNULL set to true.
func ToString(v any) String {
	if v == nil {
		return String{"", true}
	}
	switch v := v.(type) {
	case string:
		return String{v, false}
	case []byte:
		return String{string(v), false}
	case bool:
		return String{strconv.FormatBool(v), false}
	case int:
		return String{strconv.Itoa(v), false}
	case int8:
		return String{strconv.FormatInt(int64(v), 10), false}
	case int16:
		return String{strconv.FormatInt(int64(v), 10), false}
	case int32:
		return String{strconv.FormatInt(int64(v), 10), false}
	case int64:
		return String{strconv.FormatInt(v, 10), false}
	case uint:
		return String{strconv.FormatUint(uint64(v), 10), false}
	case uint8:
		return String{strconv.FormatUint(uint64(v), 10), false}
	case uint16:
		return String{strconv.FormatUint(uint64(v), 10), false}
	case uint32:
		return String{strconv.FormatUint(uint64(v), 10), false}
	case uint64:
		return String{strconv.FormatUint(v, 10), false}
	case time.Time:
		// TODO (research): does zero time mean NULL?
		if v.IsZero() {
			return String{"", true}
		}
		return String{v.Format(time.RFC3339Nano), false}
	case float32:
		return String{strconv.FormatFloat(float64(v), 'f', -1, 32), false}
	case float64:
		return String{strconv.FormatFloat(v, 'f', -1, 64), false}
	}
	if jsonMarshaler, ok := v.(json.Marshaler); ok {
		if jsonData, err := jsonMarshaler.MarshalJSON(); err == nil {
			s := strings.Trim(string(jsonData), `"`)
			// TODO (research): does [], {} mean NULL?
			if s == "[]" || s == "{}" || s == "null" {
				return String{"", true}
			}
			return String{s, false}
		}
	}
	if fmtStringer, ok := v.(fmt.Stringer); ok {
		return String{fmtStringer.String(), false}
	}
	if jsonData, err := jsonStd.Marshal(v); err == nil {
		s := strings.Trim(string(jsonData), `"`)
		// TODO (research): does [], {} mean NULL?
		if s == "[]" || s == "{}" || s == "null" {
			return String{"", true}
		}
		return String{s, false}
	}
	return String{fmt.Sprintf("%v", v), false}
}
