# Go Data Exporter

**Go Data Exporter** is a lightweight and extensible Go library for exporting tabular data from various sources (such as SQL databases, in-memory slices, etc.) into multiple formats, including:

- **CSV**
- **JSON** (standard or newline-delimited)
- **HTML**

## Features

- Pluggable scanner interface to support different data sources (`database/sql`, Hive, slices, etc.)
- Custom export codecs with configurable behavior
- Row preprocessing support
- NULL value customization
- Support for column metadata
- Easy extension with your own codecs or scanners

## Installation

```bash
go get github.com/go-data-exporter/exporter
```

## Examples

### Simple export in-memory data

```go
package main

import (
    "os"
    "time"

    "github.com/go-data-exporter/exporter"
    "github.com/go-data-exporter/exporter/codec"
    "github.com/go-data-exporter/exporter/scanner"
)

func main() {
    // Define in-memory tabular data as a slice of rows.
    // Each inner slice represents a row with mixed types.
    data := [][]any{
        {1, 2, time.Now(), 5, "text", 3.14},
        {4, 5, time.Now(), 5, "text", 3.14},
        {7, 8, time.Now(), 5, "text", 3.14},
    }

    // Create a scanner from the in-memory data.
    // It provides the data through the generic Rows interface.
    s := scanner.FromData(data)

    // Create a CSV codec for exporting the data.
    // Other formats like JSON and HTML are also available.
    c := codec.CSV()

    // Export the scanned data to standard output.
    if err := exporter.New(s, c).Write(os.Stdout); err != nil {
        log.Fatalln(err)
    }
}
```

### Simple export sql.Rows.

```go
package main

import (
    "database/sql"
    "log"
    "os"
    "time"

    "github.com/go-data-exporter/exporter"
    "github.com/go-data-exporter/exporter/codec"
    jsoncodec "github.com/go-data-exporter/exporter/codec/json"
    "github.com/go-data-exporter/exporter/scanner"
)

func main() {
    // Open a connection to the database using the specified driver and DSN.
    db, err := sql.Open("driver", "dsn")
    if err != nil {
        log.Fatalln(err)
    }
    defer db.Close()

    // Execute a SQL query and obtain the result set.
    rows, err := db.Query("SELECT * FROM table")
    if err != nil {
        log.Fatalln(err)
    }
    defer rows.Close()

    // Wrap the sql.Rows in a scanner that implements the exporter.Rows interface.
    s := scanner.FromSQL(rows, "driver")

    // Create a JSON codec.
    c := codec.JSON()

    // Export the data to standard output using the configured scanner and codec.
    if err := exporter.New(s, c).Write(os.Stdout); err != nil {
        log.Fatalln(err)
    }
}
```

### Customization

```go
package main

import (
    "database/sql"
    "log"
    "os"
    "time"

    "github.com/go-data-exporter/exporter"
    "github.com/go-data-exporter/exporter/codec"
    jsoncodec "github.com/go-data-exporter/exporter/codec/json"
    "github.com/go-data-exporter/exporter/scanner"
)

func main() {
    // Open a connection to the database using the specified driver and DSN.
    db, err := sql.Open("driver", "dsn")
    if err != nil {
        log.Fatalln(err)
    }
    defer db.Close()

    // Execute a SQL query and obtain the result set.
    rows, err := db.Query("SELECT * FROM table")
    if err != nil {
        log.Fatalln(err)
    }
    defer rows.Close()

    // Wrap the sql.Rows in a scanner that implements the exporter.Rows interface.
    s := scanner.FromSQL(rows, "driver")

    // Create a JSON codec with custom export behavior.
    c := codec.JSON(
        // Limit the output to the first 100 rows.
        jsoncodec.WithLimit(100),

        // Use newline-delimited JSON format (one JSON object per line).
        jsoncodec.WithNewlineDelimited(true),

        // Customize how time.Time values are serialized.
        jsoncodec.WithCustomType(func(v time.Time, metadata scanner.Metadata) any {
            // If the timestamp has no time component, output only the date part.
            if v.Equal(v.Truncate(24 * time.Hour)) {
                return v.Format(time.DateOnly)
            }
            return v
        }),

        // Filter out specific rows before exporting.
        // For example, exclude users with the username "admin".
        jsoncodec.WithPreProcessorFunc(func(rowID int, row map[string]any) (map[string]any, bool) {
            if row["username"] == "admin" {
                return nil, false
            }
            return row, true
        }),
    )

    // Export the data to standard output using the configured scanner and codec.
    if err := exporter.New(s, c).Write(os.Stdout); err != nil {
        log.Fatalln(err)
    }
}
```

## Supported Formats

Out of the box, the library provides codecs for exporting data to:

- **CSV** — standard comma-separated values with customizable options.
- **JSON** — standard or newline-delimited (JSON Lines).
- **HTML** — styled HTML tables with optional headers and cell formatting.

> ✅ Currently, only CSV, JSON, and HTML are officially supported.

### Custom Codecs

You can implement your own codec by satisfying the following interface:

```go
type Codec interface {
    Write(rows scanner.Rows, writer io.Writer) error
}
```
This allows you to export data to any other format — such as XML, Excel, Markdown, YAML, or even custom binary formats — by plugging in your own encoder logic.

For example, to add support for a new format:
```go
type MyCodec struct{}

func (c *MyCodec) Write(rows scanner.Rows, w io.Writer) error {
    // implement your export logic here
    return nil
}
```
Then use it with:
```go
exporter.New(scanner, &MyCodec{}).Write(os.Stdout)
```

## License
MIT

## Author
https://github.com/armantarkhanian
