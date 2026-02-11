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
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Stdout is the standard output stream used by appenders.
var Stdout io.Writer = os.Stdout

func init() {
	RegisterPlugin[DiscardAppender]("DiscardAppender")
	RegisterPlugin[ConsoleAppender]("ConsoleAppender")
	RegisterPlugin[FileAppender]("FileAppender")
	RegisterPlugin[RollingFileAppender]("RollingFileAppender")
}

// Appender defines components that handle log output.
// All implementations of Appender must be safe for concurrent use.
type Appender interface {
	Lifecycle             // Start/Stop methods for resource management
	GetName() string      // Returns the appender's name
	Append(e *Event)      // Handles writing a log event
	Write(b []byte)       // Directly writes a byte slice
	ConcurrentSafe() bool // Returns true if the appender is concurrent-safe
}

// AppenderBase provides common configuration fields for all appenders.
type AppenderBase struct {
	Name string `PluginAttribute:"name"`
}

// GetName returns the appender's name.
func (c *AppenderBase) GetName() string { return c.Name }

// ReportError reports an error to the appender.
func (c *AppenderBase) ReportError(err error) {
	// TODO: report the error
}

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

func (c *DiscardAppender) Start() error         { return nil }
func (c *DiscardAppender) Stop()                {}
func (c *DiscardAppender) Append(e *Event)      {}
func (c *DiscardAppender) Write(b []byte)       {}
func (c *DiscardAppender) ConcurrentSafe() bool { return true }

// ConsoleAppender writes formatted log events to standard output.
type ConsoleAppender struct {
	AppenderBase
	Layout Layout `PluginElement:"layout,default=TextLayout"`
}

func (c *ConsoleAppender) Start() error { return nil }
func (c *ConsoleAppender) Stop()        {}

// Append formats the event and writes it to standard output.
func (c *ConsoleAppender) Append(e *Event) {
	c.Write(c.Layout.ToBytes(e))
}

// Write writes a byte slice directly to standard output.
func (c *ConsoleAppender) Write(b []byte) {
	if _, err := Stdout.Write(b); err != nil {
		c.ReportError(err)
	}
}

func (c *ConsoleAppender) ConcurrentSafe() bool { return true }

// FileAppender writes formatted log events to a specified file.
type FileAppender struct {
	AppenderBase
	Layout   Layout `PluginElement:"layout,default=TextLayout"`
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

// Stop flushes and closes the file.
func (c *FileAppender) Stop() {
	if c.file != nil {
		_ = c.file.Sync()
		_ = c.file.Close()
	}
}

// Append formats the log event and writes it to the file.
func (c *FileAppender) Append(e *Event) {
	c.Write(c.Layout.ToBytes(e))
}

// Write writes a byte slice directly to the file.
func (c *FileAppender) Write(b []byte) {
	if _, err := c.file.Write(b); err != nil {
		c.ReportError(err)
	}
}

func (c *FileAppender) ConcurrentSafe() bool { return true }

// RollingFileAppender writes logs to files that rotate based on time.
// It is safe for multiple goroutines to call Write() concurrently only if Lock=true.
// If Lock=false, concurrent writes must be serialized by the caller (e.g., async logger).
type RollingFileAppender struct {
	AppenderBase
	Layout   Layout        `PluginElement:"layout,default=TextLayout"`
	FileDir  string        `PluginAttribute:"fileDir,default=./logs"`
	FileName string        `PluginAttribute:"fileName"`
	Interval time.Duration `PluginAttribute:"interval,default=1h"`
	MaxAge   time.Duration `PluginAttribute:"maxAge,default=168h"`
	Lock     bool          `PluginAttribute:"lock,default=false"`

	w *RollingFileWriter
	l sync.Mutex
}

// Start opens the initial log file and prepares for rotation.
func (c *RollingFileAppender) Start() error {
	c.w = &RollingFileWriter{
		fileDir:  c.FileDir,
		fileName: c.FileName,
		interval: c.Interval,
		maxAge:   c.MaxAge,
	}
	if _, err := c.w.Rotate(); err != nil {
		return err
	}
	return nil
}

// Stop flushes and closes the current file.
func (c *RollingFileAppender) Stop() {
	c.w.Close()
}

// Append formats the log event and writes it to the current file.
func (c *RollingFileAppender) Append(e *Event) {
	c.Write(c.Layout.ToBytes(e))
}

// Write writes bytes to the current log file.
// Lock=true ensures thread safety internally;
// Lock=false assumes caller serializes writes.
func (c *RollingFileAppender) Write(b []byte) {
	var (
		file *os.File
		err  error
	)
	if c.Lock { // for sync logger or multi-threaded usage
		c.l.Lock()
		file, err = c.w.Rotate()
		c.l.Unlock()
	} else { // for async logger that ensures serialization
		file, err = c.w.Rotate()
	}
	if err != nil {
		c.ReportError(err)
	}
	_, err = file.Write(b)
	if err != nil {
		c.ReportError(err)
	}
}

func (c *RollingFileAppender) ConcurrentSafe() bool { return c.Lock }

// RollingFileWriter is the low-level sequential writer.
// It is NOT safe for concurrent use;
// synchronization is the responsibility of the caller/appender.
type RollingFileWriter struct {
	fileDir  string
	fileName string
	interval time.Duration
	currFile *os.File
	currTime int64
	maxAge   time.Duration
}

// Close closes the current file.
func (w *RollingFileWriter) Close() {
	if w.currFile != nil {
		_ = w.currFile.Sync()
		_ = w.currFile.Close()
	}
}

// Rotate checks if a new file needs to be created and returns the current file.
// Old files are closed in a delayed goroutine. Concurrency must be handled externally.
func (w *RollingFileWriter) Rotate() (*os.File, error) {
	now := time.Now()
	newTime := now.Truncate(w.interval).Unix()
	if newTime <= w.currTime {
		return w.currFile, nil
	}

	formatTime := now.Format("20060102150405")
	fileName := w.fileName + "." + formatTime
	filePath := filepath.Join(w.fileDir, fileName)
	const fileFlag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	file, err := os.OpenFile(filePath, fileFlag, 0644)
	if err != nil {
		return w.currFile, err
	}

	if w.currFile != nil {
		oldFile := w.currFile
		go func() {
			// Delay closing old file. Some logs may be lost.
			time.Sleep(5 * time.Minute)
			_ = oldFile.Sync()
			_ = oldFile.Close()
			// Optionally clean expired files here if needed
			w.clearExpiredFiles()
		}()
	}

	w.currFile = file
	w.currTime = newTime
	return w.currFile, nil
}

// clearExpiredFiles removes log files older than MaxAge.
func (w *RollingFileWriter) clearExpiredFiles() {
	expiration := time.Now().Add(-w.maxAge)
	entries, _ := os.ReadDir(w.fileDir)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), w.fileName+".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(expiration) {
			_ = os.Remove(filepath.Join(w.fileDir, entry.Name()))
		}
	}
}
