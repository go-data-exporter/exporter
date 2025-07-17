// Package codec defines the Codec interface and provides factory functions
// to create different output format encoders such as CSV, JSON, and HTML.
// A Codec is responsible for writing tabular data to an output stream.
package codec

import (
	"io"

	csvcodec "github.com/go-data-exporter/exporter/codec/csv"
	htmlcodec "github.com/go-data-exporter/exporter/codec/html"
	jsoncodec "github.com/go-data-exporter/exporter/codec/json"
	"github.com/go-data-exporter/exporter/scanner"
)

// Codec defines the interface for encoding and writing tabular data
// from a scanner.Rows source to an io.Writer.
type Codec interface {
	Write(rows scanner.Rows, writer io.Writer) error
}

// JSON returns a Codec that writes data in JSON format.
// Optional configuration can be provided via functional options.
func JSON(opts ...jsoncodec.Option) Codec {
	return jsoncodec.New(opts...)
}

// CSV returns a Codec that writes data in CSV format.
// Optional configuration can be provided via functional options.
func CSV(opts ...csvcodec.Option) Codec {
	return csvcodec.New(opts...)
}

// HTML returns a Codec that writes data as an HTML table.
// Optional configuration can be provided via functional options.
func HTML(opts ...htmlcodec.Option) Codec {
	return htmlcodec.New(opts...)
}
