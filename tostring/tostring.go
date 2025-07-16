package tostring

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// String ...
type String struct {
	String string
	IsNULL bool
}

// ToString ...
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
	if jsonData, err := json.Marshal(v); err == nil {
		s := strings.Trim(string(jsonData), `"`)
		// TODO (research): does [], {} mean NULL?
		if s == "[]" || s == "{}" || s == "null" {
			return String{"", true}
		}
		return String{s, false}
	}
	return String{fmt.Sprintf("%v", v), false}
}
