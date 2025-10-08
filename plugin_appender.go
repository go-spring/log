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
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-spring/spring-base/util"
)

// Stdout is the standard output stream used by appenders.
var Stdout io.Writer = os.Stdout

func init() {
	RegisterPlugin[DiscardAppender]("Discard", PluginTypeAppender)
	RegisterPlugin[ConsoleAppender]("Console", PluginTypeAppender)
	RegisterPlugin[FileAppender]("File", PluginTypeAppender)
	RegisterPlugin[GroupAppender]("Group", PluginTypeAppender)
}

// Appender defines components that handle log output.
type Appender interface {
	Lifecycle        // Start/Stop methods for resource management
	GetName() string // Returns the appender's name
	Append(e *Event) // Handles writing a log event
	Write(b []byte)  // Directly writes a byte slice

	// ConcurrentSafe indicates that all appenders must be safe
	// for concurrent use by multiple goroutines.
	ConcurrentSafe() bool
}

var (
	_ Appender = (*GroupAppender)(nil)
	_ Appender = (*DiscardAppender)(nil)
	_ Appender = (*ConsoleAppender)(nil)
	_ Appender = (*FileAppender)(nil)
)

// AppenderBase provides common configuration fields for all appenders.
type AppenderBase struct {
	Name     string `PluginAttribute:"name"`
	MinLevel Level  `PluginAttribute:"minLevel,default=None"`
	MaxLevel Level  `PluginAttribute:"maxLevel,default=Max"`
}

// GetName returns the appender's name.
func (c *AppenderBase) GetName() string { return c.Name }

// EnableLevel checks if the given log level is enabled for this appender.
func (c *AppenderBase) EnableLevel(level Level) bool {
	return level.code >= c.MinLevel.code && level.code <= c.MaxLevel.code
}

// AppenderRef represents a reference to an appender by name.
// The actual Appender is resolved and injected later during configuration.
type AppenderRef struct {
	Appender
	Ref string `PluginAttribute:"ref"`
}

// GroupAppender forwards log events to a group of other appenders.
type GroupAppender struct {
	AppenderBase
	AppenderRefs []*AppenderRef `PluginElement:"AppenderRef"`
}

func (c *GroupAppender) Start() error         { return nil }
func (c *GroupAppender) Stop()                {}
func (c *GroupAppender) ConcurrentSafe() bool { return true }

// Append forwards the event to each child appender.
func (c *GroupAppender) Append(e *Event) {
	for _, r := range c.AppenderRefs {
		r.Append(e)
	}
}

// Write forwards raw bytes to each child appender.
func (c *GroupAppender) Write(b []byte) {
	for _, r := range c.AppenderRefs {
		r.Write(b)
	}
}

// DiscardAppender ignores all log events (no-op).
type DiscardAppender struct {
	AppenderBase
}

func (c *DiscardAppender) Start() error         { return nil }
func (c *DiscardAppender) Stop()                {}
func (c *DiscardAppender) Append(e *Event)      {}
func (c *DiscardAppender) Write(b []byte)       {}
func (c *DiscardAppender) ConcurrentSafe() bool { return true }

// ConsoleAppender writes formatted log events to standard output.
type ConsoleAppender struct {
	AppenderBase
	Layout Layout `PluginElement:"Layout,default=TextLayout"`
}

func (c *ConsoleAppender) Start() error         { return nil }
func (c *ConsoleAppender) Stop()                {}
func (c *ConsoleAppender) ConcurrentSafe() bool { return true }

// Append formats the event and writes it to standard output.
func (c *ConsoleAppender) Append(e *Event) {
	if c.EnableLevel(e.Level) {
		c.Write(c.Layout.ToBytes(e))
	}
}

// Write writes a byte slice directly to standard output.
func (c *ConsoleAppender) Write(b []byte) {
	_, _ = Stdout.Write(b)
}

// FileAppender writes formatted log events to a specified file.
type FileAppender struct {
	AppenderBase
	Layout   Layout `PluginElement:"Layout,default=TextLayout"`
	FileDir  string `PluginAttribute:"fileDir,default=./logs"`
	FileName string `PluginAttribute:"fileName"`

	file *os.File
}

func (c *FileAppender) ConcurrentSafe() bool { return true }

// Start opens the log file for appending.
func (c *FileAppender) Start() error {
	const fileFlag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	fileName := filepath.Join(c.FileDir, c.FileName)
	f, err := os.OpenFile(fileName, fileFlag, 0644)
	if err != nil {
		return err
	}
	c.file = f
	return nil
}

// Append formats the log event and writes it to the file.
func (c *FileAppender) Append(e *Event) {
	if c.EnableLevel(e.Level) {
		c.Write(c.Layout.ToBytes(e))
	}
}

// Write writes a byte slice directly to the file.
func (c *FileAppender) Write(b []byte) {
	_, _ = c.file.Write(b)
}

// Stop flushes and closes the file.
func (c *FileAppender) Stop() {
	if c.file != nil {
		_ = c.file.Sync()
		_ = c.file.Close()
	}
}

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

// RotateFileAppender allows **multiple goroutines** to call Write()
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
type RotateFileAppender struct {
	FileDir        string
	FileName       string
	ClearHours     int32
	RotateStrategy RotateStrategy
	file           atomic.Pointer[os.File]
	mutex          sync.Mutex
	currTime       atomic.Int64
}

// Start opens the initial log file.
func (c *RotateFileAppender) Start() error {
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
func (c *RotateFileAppender) Write(b []byte) {
	c.rotate()
	if file := c.file.Load(); file != nil {
		_, _ = file.Write(b)
	}
}

// Stop flushes and closes the current file.
func (c *RotateFileAppender) Stop() {
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
func (c *RotateFileAppender) rotate() {
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

func (c *RotateFileAppender) ConcurrentSafe() bool { return true }

func (c *RotateFileAppender) Append(e *Event) {
	panic(util.ErrForbiddenMethod)
}

// createFile creates or opens the current log file for appending.
// The application is responsible for ensuring the directory exists.
func (c *RotateFileAppender) createFile(t time.Time) (string, *os.File, error) {
	fileName := c.FileName + "." + c.RotateStrategy.Format(t)
	filePath := filepath.Join(c.FileDir, fileName)
	const fileFlag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	file, err := os.OpenFile(filePath, fileFlag, 0644)
	if err != nil {
		return filePath, nil, err
	}
	return filePath, file, nil
}

// clearExpiredFiles removes expired log files.
func (c *RotateFileAppender) clearExpiredFiles() {
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
