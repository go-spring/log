/*
 * Copyright 2025 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package log

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/go-spring/stdlib/errutil"
)

var (
	bufferPool sync.Pool // Pool to reuse byte buffers for layouts
	bufferCap  int       // Maximum buffer capacity allowed for reuse
)

func init() {
	bufferCap = 10 * int(bytesSizeTable["KB"])
	if s, ok := os.LookupEnv("GS_LOGGER_BUFFER_CAP"); ok {
		n, err := ParseHumanizeBytes(s)
		if err != nil {
			panic(errutil.Explain(err, "invalid value for GS_LOGGER_BUFFER_CAP: %q", s))
		}
		bufferCap = int(n)
	}
}

// Supported size units for human-readable byte values
var bytesSizeTable = map[string]int64{
	"KB": 1024,
	"MB": 1024 * 1024,
}

// HumanizeBytes represents a human-readable byte size
type HumanizeBytes int

// ParseHumanizeBytes converts a human-readable byte string (e.g., "10KB")
// to an integer value in bytes. Returns an error if the format is invalid
// or the unit is unrecognized.
func ParseHumanizeBytes(s string) (HumanizeBytes, error) {
	lastDigit := 0
	for _, r := range s {
		if !unicode.IsDigit(r) {
			break
		}
		lastDigit++
	}
	num := s[:lastDigit]
	f, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return 0, err
	}
	extra := strings.ToUpper(strings.TrimSpace(s[lastDigit:]))
	if m, ok := bytesSizeTable[extra]; ok {
		f *= m
		return HumanizeBytes(f), nil
	}
	return 0, errutil.Explain(nil, "unhandled size name: %q", extra)
}

// getBuffer retrieves a *bytes.Buffer from the pool.
// If the pool is empty, it allocates a new buffer.
func getBuffer() *bytes.Buffer {
	if v := bufferPool.Get(); v != nil {
		return v.(*bytes.Buffer)
	}
	return bytes.NewBuffer(nil)
}

// putBuffer resets a buffer and returns it to the pool for reuse.
// Buffers exceeding bufferCap are discarded to avoid holding large memory.
func putBuffer(buf *bytes.Buffer) {
	if buf.Cap() <= bufferCap {
		buf.Reset()
		bufferPool.Put(buf)
	}
}

// WriteEvent writes a log event to the provided io.Writer using the specified Layout.
// If e.RawBytes is not nil, it writes the raw bytes directly.
// Otherwise, it encodes the event into a buffer and writes it.
func WriteEvent(w io.Writer, e *Event, layout Layout) {
	if e.RawBytes != nil {
		if _, err := w.Write(e.RawBytes); err != nil {
			ReportError(err)
		}
		return
	}

	buf := getBuffer()
	defer putBuffer(buf)
	layout.EncodeTo(e, buf)
	if _, err := w.Write(buf.Bytes()); err != nil {
		ReportError(err)
	}
}
