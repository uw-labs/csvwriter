// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package csvwriter

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
	"unicode/utf8"
)

// A Writer writes records using CSV encoding.
//
// As returned by NewWriter, a Writer writes records terminated by a
// newline and uses ',' as the field delimiter. The exported fields can be
// changed to customize the details before the first call to Write or WriteAll.
//
// The writes of individual records are buffered.
// After all data has been written, the client should call the
// Flush method to guarantee all data has been forwarded to
// the underlying io.Writer.  Any errors that occurred should
// be checked by calling the Error method.
type Writer struct {
	w *bufio.Writer
}

// NewWriter returns a new Writer that writes to w.
func NewWriter(w io.Writer, bufferSize int) *Writer {
	return &Writer{
		w: bufio.NewWriterSize(w, bufferSize),
	}
}

// Writer writes a single CSV record to w along with any necessary quoting.
// A record is a slice of strings with each string being one field.
// Writes are buffered, so Flush must eventually be called to ensure
// that the record is written to the underlying io.Writer.
func (w *Writer) Write(record [][]byte) error {

	for n, field := range record {
		if n > 0 {
			if err := w.w.WriteByte(','); err != nil {
				return err
			}
		}

		// If we don't have to have a quoted field then just
		// write out the field and continue to the next field.
		if !w.fieldNeedsQuotes(field) {
			if _, err := w.w.Write(field); err != nil {
				return err
			}
			continue
		}

		if err := w.w.WriteByte('"'); err != nil {
			return err
		}
		for len(field) > 0 {
			// Search for special characters.
			i := bytes.IndexAny(field, "\"\r\n")
			if i < 0 {
				i = len(field)
			}

			// Copy verbatim everything before the special character.
			if _, err := w.w.Write(field[:i]); err != nil {
				return err
			}
			field = field[i:]

			// Encode the special character.
			if len(field) > 0 {
				var err error
				switch field[0] {
				case '"':
					_, err = w.w.WriteString(`""`)
				case '\r':
					err = w.w.WriteByte('\r')
				case '\n':
					err = w.w.WriteByte('\n')
				}
				field = field[1:]
				if err != nil {
					return err
				}
			}
		}
		if err := w.w.WriteByte('"'); err != nil {
			return err
		}
	}
	var err error
	err = w.w.WriteByte('\n')
	return err
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *Writer) Flush() error {
	return w.w.Flush()
}

// fieldNeedsQuotes reports whether our field must be enclosed in quotes.
// Fields with a Comma, fields with a quote or newline, and
// fields which start with a space must be enclosed in quotes.
// We used to quote empty strings, but we do not anymore (as of Go 1.4).
// The two representations should be equivalent, but Postgres distinguishes
// quoted vs non-quoted empty string during database imports, and it has
// an option to force the quoted behavior for non-quoted CSV but it has
// no option to force the non-quoted behavior for quoted CSV, making
// CSV with quoted empty strings strictly less useful.
// Not quoting the empty string also makes this package match the behavior
// of Microsoft Excel and Google Drive.
// For Postgres, quote the data terminating string `\.`.
func (w *Writer) fieldNeedsQuotes(field []byte) bool {
	if len(field) == 0 {
		return false
	}

	for _, c := range field {
		if special[c] == 1 {
			return true
		}
	}

	if len(field) == 2 {
		if field[0] == '\\' && field[1] == '.' {
			return true
		}
	}

	if field[0] < utf8.RuneSelf {
		if asciiSpace[field[0]] == 1 {
			return true
		}
	} else {
		r1, _ := utf8.DecodeRune(field)
		if unicode.IsSpace(r1) {
			return true
		}
	}

	return false
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}
var special = [256]uint8{',': 1, '"': 1, '\r': 1, '\n': 1}
