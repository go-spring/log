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
	"sync/atomic"
	"time"

	"github.com/lvan100/golib/errutil"
)

// Stdout is the standard output stream used by appenders.
var Stdout io.Writer = os.Stdout

func init() {
	RegisterConverter(ParseTimeRotation)

	RegisterTimeRotation("h", TimeRotation{Interval: time.Hour})
	RegisterTimeRotation("30m", TimeRotation{Interval: time.Minute * 30})
	RegisterTimeRotation("10m", TimeRotation{Interval: time.Minute * 10})

	RegisterPlugin[DiscardAppender]("Discard", PluginTypeAppender)
	RegisterPlugin[ConsoleAppender]("Console", PluginTypeAppender)
	RegisterPlugin[FileAppender]("File", PluginTypeAppender)
	RegisterPlugin[RollingFileAppender]("RollingFile", PluginTypeAppender)
}

// Appender defines components that handle log output.
// All implementations of Appender must be safe for concurrent use.
type Appender interface {
	Lifecycle        // Start/Stop methods for resource management
	GetName() string // Returns the appender's name
	Append(e *Event) // Handles writing a log event
	Write(b []byte)  // Directly writes a byte slice
}

// AppenderBase provides common configuration fields for all appenders.
type AppenderBase struct {
	Name string `PluginAttribute:"name"`
}

// GetName returns the appender's name.
func (c *AppenderBase) GetName() string { return c.Name }

var (
	_ Appender = (*DiscardAppender)(nil)
	_ Appender = (*ConsoleAppender)(nil)
	_ Appender = (*FileAppender)(nil)
	_ Appender = (*RollingFileAppender)(nil)
)

// DiscardAppender ignores all log events (no-op).
type DiscardAppender struct {
	AppenderBase
}

func (c *DiscardAppender) Start() error    { return nil }
func (c *DiscardAppender) Stop()           {}
func (c *DiscardAppender) Append(e *Event) {}
func (c *DiscardAppender) Write(b []byte)  {}

// ConsoleAppender writes formatted log events to standard output.
type ConsoleAppender struct {
	AppenderBase
	Layout Layout `PluginElement:"Layout,default=TextLayout"`
}

func (c *ConsoleAppender) Start() error { return nil }
func (c *ConsoleAppender) Stop()        {}

// Append formats the event and writes it to standard output.
func (c *ConsoleAppender) Append(e *Event) {
	c.Write(c.Layout.ToBytes(e))
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
	c.Write(c.Layout.ToBytes(e))
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

// timeRotationRegistration stores predefined time rotation policies.
var timeRotationRegistration = map[string]TimeRotation{}

// RegisterTimeRotation registers a time rotation policy with a given name.
func RegisterTimeRotation(name string, policy TimeRotation) {
	timeRotationRegistration[name] = policy
}

// ParseTimeRotation looks up a registered time rotation policy by name.
func ParseTimeRotation(s string) (TimeRotation, error) {
	if policy, ok := timeRotationRegistration[s]; ok {
		return policy, nil
	}
	return TimeRotation{}, fmt.Errorf("unknown rolling policy: %s", s)
}

// TimeRotation defines a log file rotation interval.
type TimeRotation struct {
	Interval time.Duration
}

// Time returns the truncated Unix timestamp for the given rotation interval.
func (r TimeRotation) Time(t time.Time) int64 {
	return t.Truncate(r.Interval).Unix()
}

// Format formats the time into a string with the pattern "yyyyMMddHHmmss".
func (r TimeRotation) Format(t time.Time) string {
	return t.Format("20060102150405")
}

// RollingFileAppender writes logs to files that rotate based on time.
// It is safe for multiple goroutines to call Write() concurrently.
type RollingFileAppender struct {
	AppenderBase
	Layout   Layout       `PluginElement:"Layout,default=TextLayout"`
	FileDir  string       `PluginAttribute:"fileDir,default=./logs"`
	FileName string       `PluginAttribute:"fileName"`
	Rotation TimeRotation `PluginAttribute:"rotation"`
	MaxAge   int32        `PluginAttribute:"maxAge"`

	file     atomic.Pointer[os.File]
	oldFile  atomic.Pointer[os.File]
	currTime atomic.Int64
}

// Start opens the initial log file.
func (c *RollingFileAppender) Start() error {
	now := time.Now()
	nowTime := c.Rotation.Time(now)
	filePath, file, err := c.createFile(c.Rotation.Format(now))
	if err != nil {
		return errutil.Stack(err, "Failed to create log file %s", filePath)
	}
	c.file.Store(file)
	c.currTime.Store(nowTime)
	return nil
}

// Append formats the log event and writes it to the current file.
func (c *RollingFileAppender) Append(e *Event) {
	c.Write(c.Layout.ToBytes(e))
}

// Write writes bytes to the current log file.
func (c *RollingFileAppender) Write(b []byte) {
	c.rotate()
	if file := c.file.Load(); file != nil {
		_, _ = file.Write(b)
	}
}

// Stop flushes and closes both current and previous files.
func (c *RollingFileAppender) Stop() {
	if file := c.oldFile.Swap(nil); file != nil {
		_ = file.Sync()
		_ = file.Close()
	}
	if file := c.file.Swap(nil); file != nil {
		_ = file.Sync()
		_ = file.Close()
	}
}

// rotate checks if a new rotation interval has started.
// If so, it closes the old file, opens a new one, and triggers cleanup.
func (c *RollingFileAppender) rotate() {
	now := time.Now()
	nowTime := c.Rotation.Time(now)
	oldTime := c.currTime.Load()
	if nowTime <= oldTime {
		return
	}
	if !c.currTime.CompareAndSwap(oldTime, nowTime) {
		return
	}

	// Close the previous rotation file
	if file := c.oldFile.Swap(nil); file != nil {
		_ = file.Sync()
		_ = file.Close()
	}

	filePath, file, err := c.createFile(c.Rotation.Format(now))
	if err != nil {
		err = errutil.Stack(err, "Failed to create log file %s", filePath)
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}

	oldFile := c.file.Load()
	c.oldFile.Store(oldFile)

	c.file.Store(file)
	c.currTime.Store(nowTime)

	// Cleanup expired log files asynchronously
	go c.clearExpiredFiles()
}

// createFile creates or opens the current log file for appending.
func (c *RollingFileAppender) createFile(formatTime string) (string, *os.File, error) {
	fileName := c.FileName + "." + formatTime
	filePath := filepath.Join(c.FileDir, fileName)
	const fileFlag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	file, err := os.OpenFile(filePath, fileFlag, 0644)
	if err != nil {
		return filePath, nil, err
	}
	return filePath, file, nil
}

// clearExpiredFiles removes log files older than MaxAge.
func (c *RollingFileAppender) clearExpiredFiles() {
	expiration := time.Now().Add(-time.Duration(c.MaxAge) * time.Hour)
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
