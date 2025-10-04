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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-spring/spring-base/util"
)

func init() {
	// register built-in converters
	RegisterConverter(ParseRotateStrategy)

	// register built-in rotate strategies
	RegisterRotateStrategy("1h", FixedRotateStrategy{Interval: time.Hour})
	RegisterRotateStrategy("30m", FixedRotateStrategy{Interval: 30 * time.Minute})
	RegisterRotateStrategy("10m", FixedRotateStrategy{Interval: 10 * time.Minute})
}

var rotateStrategyRegistry = map[string]RotateStrategy{}

// RotateStrategy defines the interface for log rotation strategies.
type RotateStrategy interface {
	Time(t time.Time) int64
	Format(t time.Time) string
}

// FixedRotateStrategy represents a fixed-interval rotation strategy.
type FixedRotateStrategy struct {
	Interval time.Duration // The rotation interval duration.
}

// Time returns the timestamp aligned to the nearest previous rotation point.
func (r FixedRotateStrategy) Time(t time.Time) int64 {
	seconds := int64(r.Interval.Seconds())
	return (t.Unix() / seconds) * seconds
}

// Format formats the time into a string with the pattern "yyyyMMddHHmmss".
func (r FixedRotateStrategy) Format(t time.Time) string {
	return t.Format("20060102150405")
}

// RegisterRotateStrategy registers a rotation strategy with a given name.
func RegisterRotateStrategy(name string, strategy RotateStrategy) {
	rotateStrategyRegistry[name] = strategy
}

// ParseRotateStrategy retrieves a registered rotation strategy by name.
func ParseRotateStrategy(name string) (RotateStrategy, error) {
	s, ok := rotateStrategyRegistry[name]
	if !ok {
		return nil, fmt.Errorf("invalid rotate strategy: %q", name)
	}
	return s, nil
}

// FileWriter defines the interface for log file writers.
type FileWriter interface {
	Start() error
	Write(b []byte)
	Stop()
}

// FileWriterAsAppender is an adapter that wraps a FileWriter
// to make it compatible with components expecting an Appender.
// However, calling Append on this type is explicitly forbidden.
type FileWriterAsAppender struct {
	FileWriter
}

func (c FileWriterAsAppender) Append(e *Event) {
	panic(util.ErrForbiddenMethod)
}

// RotateFileWriterBase provides common fields and methods
// for writing logs into rotated files.
type RotateFileWriterBase struct {
	FileDir        string
	FileName       string
	ClearHours     int32
	RotateStrategy RotateStrategy
}

// createFile creates or opens the current log file for appending.
// The application is responsible for ensuring the directory exists.
func (c *RotateFileWriterBase) createFile(t time.Time) (string, *os.File, error) {
	fileName := c.FileName + "." + c.RotateStrategy.Format(t)
	filePath := filepath.Join(c.FileDir, fileName)
	const fileFlag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	file, err := os.OpenFile(filePath, fileFlag, os.ModePerm)
	if err != nil {
		return filePath, nil, err
	}
	return filePath, file, nil
}

// clearExpiredFiles removes expired log files.
func (c *RotateFileWriterBase) clearExpiredFiles() {
	expiration := time.Now().Add(-time.Duration(c.ClearHours) * time.Hour)
	entries, _ := os.ReadDir(c.FileDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), c.FileName+".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(expiration) {
			filePath := fmt.Sprintf("%s/%s", c.FileDir, entry.Name())
			_ = os.Remove(filePath)
		}
	}
}

// AsyncRotateFileWriter is intended for **single-writer usage**,
// typically in asynchronous logging pipelines where one goroutine
// is responsible for writing logs to disk.
//
// Usage constraints:
// - Only one goroutine should call Write concurrently.
// - Stop() must not be called concurrently with Write().
// - Best used with an async queue where a single worker flushes logs to disk.
type AsyncRotateFileWriter struct {
	RotateFileWriterBase
	file     *os.File
	currTime int64
}

// NewAsyncRotateFileWriter creates a new AsyncRotateFileWriter instance.
func NewAsyncRotateFileWriter(base RotateFileWriterBase) *AsyncRotateFileWriter {
	return &AsyncRotateFileWriter{RotateFileWriterBase: base}
}

// Start opens the initial log file.
func (c *AsyncRotateFileWriter) Start() error {
	now := time.Now()
	filePath, file, err := c.createFile(now)
	if err != nil {
		return util.WrapError(err, "Failed to create log file %s", filePath)
	}
	c.file = file
	c.currTime = c.RotateStrategy.Time(now)
	return nil
}

// Write writes bytes directly to the current log file.
func (c *AsyncRotateFileWriter) Write(b []byte) {
	c.rotate()
	if c.file != nil {
		_, _ = c.file.Write(b)
	}
}

// Stop flushes and closes the current log file.
// Must not be called concurrently with Write().
func (c *AsyncRotateFileWriter) Stop() {
	if c.file != nil {
		_ = c.file.Sync()
		_ = c.file.Close()
	}
}

// rotate checks if the current time has passed into a new rotation slot.
// If so, it closes the old file, opens a new one, and triggers cleanup.
// Risk: If file creation fails during rotation, new logs will be lost
// until the issue is resolved.
func (c *AsyncRotateFileWriter) rotate() {
	now := time.Now()
	nowTime := c.RotateStrategy.Time(now)
	if nowTime <= c.currTime {
		return // still in the current slot
	}

	// close the old file
	if c.file != nil {
		_ = c.file.Sync()
		_ = c.file.Close()
		c.file = nil
	}

	filePath, file, err := c.createFile(now)
	if err != nil {
		err = util.WrapError(err, "Failed to create log file %s", filePath)
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
	c.file = file
	c.currTime = nowTime

	// trigger cleanup after each rotation for timely housekeeping
	go c.clearExpiredFiles()
}

// SyncRotateFileWriter allows **multiple goroutines** to call Write()
// safely, at the cost of slightly higher overhead and potential
// (acceptable) log loss during rotation.
//
// Usage scenarios:
//   - High-concurrency applications where logs may be produced
//     from many goroutines.
//
// Risks:
//   - During rotation, a small number of writes may fail if they
//     occur after the old file is closed but before the new file is ready.
//   - During Stop(), concurrent writes may also be lost.
//   - If zero log loss is required, use AsyncRotateFileWriter
//     with a dedicated logging goroutine instead.
type SyncRotateFileWriter struct {
	RotateFileWriterBase
	file     atomic.Pointer[os.File]
	mutex    sync.Mutex
	currTime atomic.Int64
}

// NewSyncRotateFileWriter creates a new instance of SyncRotateFileWriter.
func NewSyncRotateFileWriter(base RotateFileWriterBase) *SyncRotateFileWriter {
	return &SyncRotateFileWriter{RotateFileWriterBase: base}
}

// Start opens the initial log file.
func (c *SyncRotateFileWriter) Start() error {
	now := time.Now()
	filePath, file, err := c.createFile(now)
	if err != nil {
		return util.WrapError(err, "Failed to create log file %s", filePath)
	}
	c.file.Store(file)
	c.currTime.Store(c.RotateStrategy.Time(now))
	return nil
}

// Write writes bytes to the current log file.
// May lose a few writes during rotation or Stop().
func (c *SyncRotateFileWriter) Write(b []byte) {
	c.rotate()
	if file := c.file.Load(); file != nil {
		_, _ = file.Write(b)
	}
}

// Stop flushes and closes the current file.
func (c *SyncRotateFileWriter) Stop() {
	c.rotate()
	if file := c.file.Swap(nil); file != nil {
		_ = file.Sync()
		_ = file.Close()
	}
}

// rotate checks if the current time has passed into a new rotation slot.
// If so, it closes the old file, opens a new one, and triggers cleanup.
// Risk: If file creation fails during rotation, new logs will be lost
// until the issue is resolved.
func (c *SyncRotateFileWriter) rotate() {
	now := time.Now()
	nowTime := c.RotateStrategy.Time(now)
	if nowTime <= c.currTime.Load() {
		return // still in the current slot
	}

	// serialize rotation
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// double-check after acquiring the lock
	if nowTime <= c.currTime.Load() {
		return
	}

	// close the old file
	if file := c.file.Load(); file != nil {
		_ = file.Sync()
		_ = file.Close()
	}

	filePath, file, err := c.createFile(now)
	if err != nil {
		err = util.WrapError(err, "Failed to create log file %s", filePath)
		_, _ = fmt.Fprintln(os.Stderr, err)
		c.file.Store(nil)
		return
	}
	c.file.Store(file)
	c.currTime.Store(nowTime)

	// trigger cleanup after each rotation for timely housekeeping
	go c.clearExpiredFiles()
}
