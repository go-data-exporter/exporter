// Package exporter provides a unified interface for exporting tabular data
// using pluggable codecs (e.g., CSV, JSON, HTML). It wraps a data source
// and a codec implementation to perform the export.
package exporter

import (
	"io"
	"os"

	"github.com/go-data-exporter/exporter/codec"
	"github.com/go-data-exporter/exporter/scanner"
)

// Exporter is the main struct that coordinates exporting data.
// It uses a scanner.Rows as the data source and a codec.Codec
// to determine the output format.
type Exporter struct {
	rows  scanner.Rows
	codec codec.Codec
}

// New creates a new Exporter instance using the given data source and codec.
func New(rows scanner.Rows, codec codec.Codec) *Exporter {
	return &Exporter{
		rows:  rows,
		codec: codec,
	}
}

// Write writes the exported data to the given io.Writer using the codec.
func (cs *Exporter) Write(writer io.Writer) error {
	return cs.codec.Write(cs.rows, writer)
}

// WriteFile writes the exported data directly to a file specified by filename.
func (cs *Exporter) WriteFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = cs.Write(f); err != nil {
		_ = f.Sync()
		return err
	}
	_ = f.Sync()
	return f.Close()
}
