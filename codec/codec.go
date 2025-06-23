package codec

import (
	"io"

	csvcodec "github.com/go-data-exporter/exporter/codec/csv"
	htmlcodec "github.com/go-data-exporter/exporter/codec/html"
	jsoncodec "github.com/go-data-exporter/exporter/codec/json"
	"github.com/go-data-exporter/exporter/scanner"
)

type Codec interface {
	Write(rows scanner.Rows, writer io.Writer) error
}

func JSON(opts ...jsoncodec.Option) Codec {
	return jsoncodec.New(opts...)
}

func CSV(opts ...csvcodec.Option) Codec {
	return csvcodec.New(opts...)
}

func HTML(opts ...htmlcodec.Option) Codec {
	return htmlcodec.New(opts...)
}
