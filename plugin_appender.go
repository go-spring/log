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
	_ Appender = (*DiscardAppender)(nil)
	_ Appender = (*ConsoleAppender)(nil)
	_ Appender = (*FileAppender)(nil)
	_ Appender = (*GroupAppender)(nil)
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
	FileDir  string `PluginAttribute:"dir,default=./log"`
	FileName string `PluginAttribute:"fileName"`

	file *os.File
}

func (c *FileAppender) ConcurrentSafe() bool { return true }

// Start opens the log file for appending.
func (c *FileAppender) Start() error {
	const fileFlag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	fileName := filepath.Join(c.FileDir, c.FileName)
	f, err := os.OpenFile(fileName, fileFlag, os.ModePerm)
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
