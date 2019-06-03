// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package csvwriter

import (
	"bytes"
	"errors"
	"testing"
)

var writeTests = []struct {
	Input  [][]string
	Output string
	Error  error
}{
	{Input: [][]string{{"abc"}}, Output: "abc\n"},
	{Input: [][]string{{`"abc"`}}, Output: `"""abc"""` + "\n"},
	{Input: [][]string{{`a"b`}}, Output: `"a""b"` + "\n"},
	{Input: [][]string{{`"a"b"`}}, Output: `"""a""b"""` + "\n"},
	{Input: [][]string{{" abc"}}, Output: `" abc"` + "\n"},
	{Input: [][]string{{"abc,def"}}, Output: `"abc,def"` + "\n"},
	{Input: [][]string{{"abc", "def"}}, Output: "abc,def\n"},
	{Input: [][]string{{"abc"}, {"def"}}, Output: "abc\ndef\n"},
	{Input: [][]string{{"abc\ndef"}}, Output: "\"abc\ndef\"\n"},
	{Input: [][]string{{"abc\rdef"}}, Output: "\"abc\rdef\"\n"},
	{Input: [][]string{{""}}, Output: "\n"},
	{Input: [][]string{{"", ""}}, Output: ",\n"},
	{Input: [][]string{{"", "", ""}}, Output: ",,\n"},
	{Input: [][]string{{"", "", "a"}}, Output: ",,a\n"},
	{Input: [][]string{{"", "a", ""}}, Output: ",a,\n"},
	{Input: [][]string{{"", "a", "a"}}, Output: ",a,a\n"},
	{Input: [][]string{{"a", "", ""}}, Output: "a,,\n"},
	{Input: [][]string{{"a", "", "a"}}, Output: "a,,a\n"},
	{Input: [][]string{{"a", "a", ""}}, Output: "a,a,\n"},
	{Input: [][]string{{"a", "a", "a"}}, Output: "a,a,a\n"},
	{Input: [][]string{{`\.`}}, Output: "\"\\.\"\n"},
	{Input: [][]string{{"x09\x41\xb4\x1c", "aktau"}}, Output: "x09\x41\xb4\x1c,aktau\n"},
	{Input: [][]string{{",x09\x41\xb4\x1c", "aktau"}}, Output: "\",x09\x41\xb4\x1c\",aktau\n"},
}

func TestWrite(t *testing.T) {
	for n, tt := range writeTests {
		b := &bytes.Buffer{}
		f := NewWriter(b, 1024)
		err := writeAll(f, tt.Input)
		if err != tt.Error {
			t.Errorf("Unexpected error:\ngot  %v\nwant %v", err, tt.Error)
		}
		out := b.String()
		if out != tt.Output {
			t.Errorf("#%d: out=%q want %q", n, out, tt.Output)
		}
	}
}

func writeAll(w *Writer, input [][]string) error {
	for _, in := range input {
		inBytes := make([][]byte, len(in))
		for i, str := range in {
			inBytes[i] = []byte(str)
		}
		if err := w.Write(inBytes); err != nil {
			return err
		}
	}
	return w.Flush()
}

type errorWriter struct{}

func (e errorWriter) Write(b []byte) (int, error) {
	return 0, errors.New("Test")
}

func TestError(t *testing.T) {
	b := &bytes.Buffer{}
	f := NewWriter(b, 1024)
	f.Write([][]byte{[]byte("abc")})
	err := f.Flush()

	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
	}

	f = NewWriter(errorWriter{}, 1024)
	f.Write([][]byte{[]byte("abc")})
	err = f.Flush()

	if err == nil {
		t.Error("Error should not be nil")
	}
}
