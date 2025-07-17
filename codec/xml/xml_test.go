package xmlcodec

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-data-exporter/exporter/scanner"
	"github.com/go-data-exporter/exporter/tostring"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Error("New() returned nil")
	}
	if c.customMapper == nil {
		t.Error("customMapper not initialized")
	}
	if c.limit != -1 {
		t.Error("default limit should be -1")
	}
}

func TestWithCustomType(t *testing.T) {
	customFn := func(v int, _ scanner.Metadata) tostring.String {
		return tostring.String{String: "custom:" + tostring.ToString(v).String}
	}

	c := New(WithCustomType(customFn))

	var testInt int = 42
	typ := reflect.TypeOf(testInt)
	if _, ok := c.customMapper[typ]; !ok {
		t.Error("custom type not registered")
	}

	// Test with actual data
	data := [][]any{{42}}
	s := scanner.FromData(data)
	var buf bytes.Buffer

	err := c.Write(s, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "custom:42") {
		t.Errorf("custom function not applied, got: %s", output)
	}
}

func TestWithPreProcessorFunc(t *testing.T) {
	preProcess := func(rowID int, row []string) ([]string, bool) {
		if row[1] == "second" {
			return nil, false
		}
		return row, true
	}

	c := New(WithPreProcessorFunc(preProcess))
	if c.preProcessorFunc == nil {
		t.Error("preProcessorFunc not set")
	}

	data := [][]any{
		{1, "first"},
		{2, "second"},
		{3, "third"},
	}
	s := scanner.FromData(data)
	var buf bytes.Buffer

	err := c.Write(s, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "second") {
		t.Error("preProcessorFunc did not filter row 2")
	}
	if !strings.Contains(output, "first") || !strings.Contains(output, "third") {
		t.Error("preProcessorFunc filtered wrong rows")
	}
}

func TestWithLimit(t *testing.T) {
	c := New(WithLimit(2))
	if c.limit != 2 {
		t.Error("limit not set correctly")
	}

	data := [][]any{
		{1, "first"},
		{2, "second"},
		{3, "third"},
	}
	s := scanner.FromData(data)
	var buf bytes.Buffer

	err := c.Write(s, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if strings.Count(output, "<row>") != 2 {
		t.Errorf("expected 2 rows, got %d", strings.Count(output, "<row>"))
	}
	if strings.Contains(output, "<third>") {
		t.Error("limit not respected")
	}
}

func TestWrite(t *testing.T) {
	now := time.Now()
	data := [][]any{
		{1, 2, now, 5, "text", 3.14},
		{4, 5, now, nil, "<text>", 3.14},
		{7, 8, now, 5, "text", 3.14},
	}
	s := scanner.FromData(data)
	c := New()
	var buf bytes.Buffer

	err := c.Write(s, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Check XML declaration and root element
	if !strings.HasPrefix(output, `<?xml version="1.0" encoding="UTF-8"?>`+"\n<data>") {
		t.Error("missing XML declaration or root element")
	}

	// Check row count
	rowCount := strings.Count(output, "<row>")
	if rowCount != 3 {
		t.Errorf("expected 3 rows, got %d", rowCount)
	}

	// Check NULL handling
	if strings.Contains(output, "nil") {
		t.Error("NULL values should be omitted")
	}

	// Check XML escaping
	if !strings.Contains(output, "&lt;text&gt;") {
		t.Error("XML special characters not escaped")
	}

	// Check time formatting
	if !strings.Contains(output, now.Format(time.RFC3339Nano)) {
		t.Error("time not formatted correctly")
	}
}

func TestWriteWithError(t *testing.T) {
	// Test with empty data (should not error)
	c := New()
	var buf bytes.Buffer

	s := scanner.FromData([][]any{})
	err := c.Write(s, &buf)
	if err != nil {
		t.Errorf("Write with empty data should not error, got: %v", err)
	}

	// Test with invalid data (nil slice)
	s = scanner.FromData(nil)
	err = c.Write(s, &buf)
	if err != nil {
		t.Error("unexpected error with nil data")
	}
	if output := buf.String(); output != "" {
		t.Error("unexpected data")
	}
}

func TestToString(t *testing.T) {
	c := New()

	// Test NULL
	result := c.toString(nil, scanner.Metadata{})
	if !result.IsNULL {
		t.Error("nil value should be marked as NULL")
	}

	// Test custom type with actual data
	customFn := func(v string, _ scanner.Metadata) tostring.String {
		return tostring.String{String: "CUSTOM:" + v}
	}
	c = New(WithCustomType(customFn))

	data := [][]any{{"test"}}
	s := scanner.FromData(data)
	var buf bytes.Buffer

	err := c.Write(s, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "CUSTOM:test") {
		t.Errorf("custom function not applied, got: %s", output)
	}

	// Test default conversion
	c = New()
	data = [][]any{{42}}
	s = scanner.FromData(data)
	buf.Reset()

	err = c.Write(s, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "42") {
		t.Errorf("default conversion failed, got: %s", output)
	}
}

func TestWriteEmpty(t *testing.T) {
	c := New()
	var buf bytes.Buffer

	// Empty data
	s := scanner.FromData([][]any{})
	err := c.Write(s, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Error("empty data should produce no output")
	}

	// With limit 0
	c = New(WithLimit(0))
	s = scanner.FromData([][]any{{1, "test"}})
	buf.Reset()
	err = c.Write(s, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("limit 0 should produce no output")
	}
}
