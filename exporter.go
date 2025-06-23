package exporter

import (
	"io"
	"os"

	"github.com/go-data-exporter/exporter/codec"
	"github.com/go-data-exporter/exporter/scanner"
)

type Exporter struct {
	rows  scanner.Rows
	codec codec.Codec
}

func New(rows scanner.Rows, codec codec.Codec) *Exporter {
	return &Exporter{
		rows:  rows,
		codec: codec,
	}
}

func (cs *Exporter) Write(writer io.Writer) error {
	return cs.codec.Write(cs.rows, writer)
}

func (cs *Exporter) WriteFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := cs.Write(f); err != nil {
		return err
	}
	return f.Close()
}
