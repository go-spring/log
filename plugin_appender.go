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
)

// Stdout is the standard output stream used by appenders.
var Stdout io.Writer = os.Stdout

func init() {
	RegisterPlugin[DiscardAppender]("Discard", PluginTypeAppender)
	RegisterPlugin[ConsoleAppender]("Console", PluginTypeAppender)
	RegisterPlugin[FileAppender]("File", PluginTypeAppender)
	RegisterPlugin[MultiAppender]("Multi", PluginTypeAppender)
	RegisterPlugin[LayoutAppender]("Layout", PluginTypeAppender)
	RegisterPlugin[LevelFilterAppender]("LevelFilter", PluginTypeAppender)
}

// Appender is an interface that defines components that handle log output.
type Appender interface {
	Lifecycle        // Appenders must be startable and stoppable
	GetName() string // Returns the appender name
	Append(e *Event) // Handles writing a log event
	Write(b []byte)  // Directly writes a byte slice
}

// AppenderRef represents a reference to an appender by name.
// The actual appender is resolved and injected later during configuration.
type AppenderRef struct {
	Appender
	Ref string `PluginAttribute:"ref"`
}

type AppenderRefs struct {
	AppenderRefs []*AppenderRef `PluginElement:"AppenderRef"` // Attached appenders
}

func (c *AppenderRefs) Start() error { return nil }

func (c *AppenderRefs) Stop() {}

func (c *AppenderRefs) Append(e *Event) {
	for _, r := range c.AppenderRefs {
		r.Append(e)
	}
}

func (c *AppenderRefs) Write(b []byte) {
	for _, r := range c.AppenderRefs {
		r.Write(b)
	}
}

var (
	_ Appender = (*DiscardAppender)(nil)
	_ Appender = (*ConsoleAppender)(nil)
	_ Appender = (*FileAppender)(nil)
)

// AppenderBase provides common configuration and default behavior for appenders.
type AppenderBase struct {
	Name string `PluginAttribute:"name"`
}

func (c *AppenderBase) GetName() string { return c.Name }

// DiscardAppender ignores all log events (no output).
type DiscardAppender struct {
	AppenderBase
}

func (c *DiscardAppender) Start() error    { return nil }
func (c *DiscardAppender) Stop()           {}
func (c *DiscardAppender) Append(e *Event) {}
func (c *DiscardAppender) Write(b []byte)  {}

// ConsoleAppender writes formatted log events to stdout.
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

// Write writes a byte slice directly to the stdout.
func (c *ConsoleAppender) Write(b []byte) {
	_, _ = Stdout.Write(b)
}

// FileAppender writes formatted log events to a specified file.
type FileAppender struct {
	AppenderBase
	Layout   Layout `PluginElement:"Layout,default=TextLayout"`
	FileName string `PluginAttribute:"fileName"`

	file *os.File
}

// Start opens the log file for appending.
func (c *FileAppender) Start() error {
	const fileFlag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	f, err := os.OpenFile(c.FileName, fileFlag, os.ModePerm)
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

// -----------------------------------------------------------------------------
// Appender Utilities
// -----------------------------------------------------------------------------

var (
	_ Appender = (*MultiAppender)(nil)
	_ Appender = (*LayoutAppender)(nil)
	_ Appender = (*LevelFilterAppender)(nil)
)

// MultiAppender delegates log events to multiple underlying appenders.
// It is useful when you want to send log events to several outputs.
type MultiAppender struct {
	AppenderBase
	AppenderRefs
}

// LayoutAppender is an Appender wrapper that applies a Layout
// to each log event before delegating the formatted output
// to the underlying Appender.
type LayoutAppender struct {
	AppenderBase
	AppenderRefs
	Layout Layout `PluginElement:"Layout,default=TextLayout"`
}

// Append formats the given log event using the configured Layout
// and then writes the formatted bytes to the underlying Appender.
func (c *LayoutAppender) Append(e *Event) {
	c.AppenderRefs.Write(c.Layout.ToBytes(e))
}

// LevelFilterAppender filters log events based on their severity level.
// Only events with levels between MinLevel and MaxLevel (exclusive)
// will be passed to the underlying Appender.
type LevelFilterAppender struct {
	AppenderBase
	AppenderRefs
	MinLevel Level
	MaxLevel Level
}

// Append filters the incoming log event according to the defined level range.
// If the event's level is outside the range [MinLevel, MaxLevel),
// the event will be ignored. Otherwise, it is forwarded to the underlying Appender.
func (c *LevelFilterAppender) Append(e *Event) {
	if e.Level.code >= c.MinLevel.code && e.Level.code < c.MaxLevel.code {
		c.AppenderRefs.Append(e)
	}
}
