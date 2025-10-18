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
	"sync/atomic"

	"github.com/go-spring/spring-base/util"
)

func init() {
	RegisterConverter[BufferFullPolicy](ParseBufferFullPolicy)
}

func init() {
	RegisterPlugin[FileLogger]("File", PluginTypeLogger)
	RegisterPlugin[SyncLogger]("Logger", PluginTypeLogger)
	RegisterPlugin[AsyncLogger]("AsyncLogger", PluginTypeLogger)
	RegisterPlugin[AppenderRef]("AppenderRef", PluginTypeAppenderRef)
}

// Logger is the interface implemented by all loggers.
type Logger interface {
	Appender
	EnableLevel(level Level) bool
}

// AppenderRef represents a reference to an appender by name.
// The actual Appender is resolved and injected later during configuration.
type AppenderRef struct {
	Appender
	Ref      string `PluginAttribute:"ref"`
	MinLevel Level  `PluginAttribute:"minLevel,default=None"`
	MaxLevel Level  `PluginAttribute:"maxLevel,default=Max"`
}

// EnableLevel checks if the given log level is enabled for this appender.
func (c *AppenderRef) EnableLevel(level Level) bool {
	return level.code >= c.MinLevel.code && level.code <= c.MaxLevel.code
}

// Append forwards the event to each child appender.
func (c *AppenderRef) Append(e *Event) {
	if c.EnableLevel(e.Level) {
		c.Appender.Append(e)
	}
}

// Write forwards raw bytes to each child appender.
func (c *AppenderRef) Write(b []byte) {
	c.Appender.Write(b)
}

// LoggerBase contains fields shared by all logger configurations.
type LoggerBase struct {
	Name         string         `PluginAttribute:"name"`          // Logger name
	Tags         string         `PluginAttribute:"tags,default="` // Optional tags
	MinLevel     Level          `PluginAttribute:"minLevel,default=None"`
	MaxLevel     Level          `PluginAttribute:"maxLevel,default=Max"`
	Layout       Layout         `PluginElement:"Layout?"`
	AppenderRefs []*AppenderRef `PluginElement:"AppenderRef"`
}

// GetName returns the name of the logger.
func (c *LoggerBase) GetName() string {
	return c.Name
}

// EnableLevel checks if the given log level is enabled for this logger.
func (c *LoggerBase) EnableLevel(level Level) bool {
	return level.code >= c.MinLevel.code && level.code <= c.MaxLevel.code
}

// sendToAppenders forwards the event to each child appender.
func (c *LoggerBase) sendToAppenders(e *Event) {
	for _, r := range c.AppenderRefs {
		if c.Layout == nil {
			r.Append(e)
		} else {
			r.Write(c.Layout.ToBytes(e))
		}
	}
}

// writeToAppenders forwards raw bytes to each child appender.
func (c *LoggerBase) writeToAppenders(b []byte) {
	for _, r := range c.AppenderRefs {
		r.Write(b)
	}
}

// ----------------------------------------------------------------------------
// common loggers
// ----------------------------------------------------------------------------

var (
	_ Logger = (*DiscardLogger)(nil)
	_ Logger = (*SyncLogger)(nil)
	_ Logger = (*AsyncLogger)(nil)
	_ Logger = (*FileLogger)(nil)
)

type DiscardLogger struct {
	LoggerBase
	DiscardAppender
}

type ConsoleLogger struct {
	LoggerBase
	ConsoleAppender
}

type FileLogger struct {
	LoggerBase
	FileAppender
}

// SyncLogger is a synchronous logger that immediately forwards events to appenders.
type SyncLogger struct {
	LoggerBase
}

func (c *SyncLogger) Start() error { return nil }
func (c *SyncLogger) Stop()        {}

// Append sends the event directly to appenders (blocking).
func (c *SyncLogger) Append(e *Event) {
	c.sendToAppenders(e)
	PutEvent(e) // Return event to the pool
}

// Write writes raw bytes directly to appenders.
func (c *SyncLogger) Write(b []byte) {
	c.writeToAppenders(b)
}

// BufferFullPolicy specifies what to do when an async buffer is full.
type BufferFullPolicy int

const (
	BufferFullPolicyBlock         = BufferFullPolicy(0) // Block until space is available
	BufferFullPolicyDiscard       = BufferFullPolicy(1) // Drop the new event or data
	BufferFullPolicyDiscardOldest = BufferFullPolicy(2) // Drop the oldest event or data
)

// ParseBufferFullPolicy converts a string to a BufferFullPolicy.
func ParseBufferFullPolicy(s string) (BufferFullPolicy, error) {
	switch s {
	case "Block":
		return BufferFullPolicyBlock, nil
	case "Discard":
		return BufferFullPolicyDiscard, nil
	case "DiscardOldest":
		return BufferFullPolicyDiscardOldest, nil
	default:
		return -1, util.FormatError(nil, "invalid BufferFullPolicy %s", s)
	}
}

// AsyncLogger is an asynchronous logger that buffers events
// and processes them in a dedicated background goroutine.
type AsyncLogger struct {
	LoggerBase

	BufferSize       int              `PluginAttribute:"bufferSize,default=10000"`
	BufferFullPolicy BufferFullPolicy `PluginAttribute:"bufferFullPolicy,default=Discard"`

	buf  chan any      // Buffered channel for events and raw data
	wait chan struct{} // Channel closed when worker goroutine exits
	stop *Event        // Special marker event to signal shutdown

	discardCounter int64 // Counter for discarded events
}

// GetDiscardCounter returns the total number of discarded events and data.
func (c *AsyncLogger) GetDiscardCounter() int64 {
	return atomic.LoadInt64(&c.discardCounter)
}

// Start initializes channels and launches the worker goroutine.
func (c *AsyncLogger) Start() error {
	if c.BufferSize < 100 {
		return util.FormatError(nil, "bufferSize is too small")
	}
	c.buf = make(chan any, c.BufferSize)
	c.wait = make(chan struct{})
	c.stop = &Event{}

	// Worker goroutine to process buffered items
	go func() {
		for v := range c.buf {
			if v == c.stop {
				break
			}
			switch x := v.(type) {
			case *Event:
				c.sendToAppenders(x)
				PutEvent(x)
			case []byte:
				c.writeToAppenders(x)
			default: // for linter
			}
		}
		close(c.wait)
	}()
	return nil
}

// Append enqueues a log event into the buffer.
// Behavior on full buffer depends on BufferFullPolicy.
func (c *AsyncLogger) Append(e *Event) {
	select {
	case c.buf <- e:
	default:
		c.onBufferFull(e)
	}
}

// Write enqueues raw bytes into the buffer.
// Behavior on full buffer depends on BufferFullPolicy.
func (c *AsyncLogger) Write(b []byte) {
	select {
	case c.buf <- b:
	default:
		c.onBufferFull(b)
	}
}

// onBufferFull handles the case when the async buffer is full.
func (c *AsyncLogger) onBufferFull(v any) {
	switch c.BufferFullPolicy {
	case BufferFullPolicyDiscardOldest:
		var exit bool
		for {
			select {
			case c.buf <- v:
				exit = true
			default:
				// Remove one element to make space
				select {
				case x := <-c.buf:
					atomic.AddInt64(&c.discardCounter, 1)
					if e, ok := x.(*Event); ok {
						PutEvent(e)
					}
				default: // for linter
				}
			}
			if exit {
				break
			}
		}
	case BufferFullPolicyBlock:
		// Block until space is available
		c.buf <- v
	case BufferFullPolicyDiscard:
		// Discard new item
		atomic.AddInt64(&c.discardCounter, 1)
		if e, ok := v.(*Event); ok {
			PutEvent(e)
		}
		return
	default: // for linter
	}
}

// Stop gracefully shuts down the async logger.
// It signals the worker goroutine to exit and waits for it.
// NOTE: Stop must be called only once, otherwise panic may occur.
func (c *AsyncLogger) Stop() {
	c.buf <- c.stop
	<-c.wait
	close(c.buf)
}

// ----------------------------------------------------------------------------
// rolling file logger
// ----------------------------------------------------------------------------

// RollingFileLogger is a logger implementation that writes log events to files.
// It can work in either synchronous or asynchronous mode depending on AsyncWrite.
// It also supports splitting warning/error logs into a separate file.
type RollingFileLogger struct {
	LoggerBase
	logger Logger

	// File output configuration
	FileDir  string `PluginAttribute:"fileDir,default=./logs"`
	FileName string `PluginAttribute:"fileName,default=app.log"`

	// If true, warning/error logs go to a separate .wf file.
	Separate bool `PluginAttribute:"separate,default=false"`

	// rotation and cleanup configuration
	ClearHours    int32         `PluginAttribute:"clearHours,default=168"`
	RollingPolicy RollingPolicy `PluginAttribute:"rollingPolicy,default=1h"`

	// asynchronous logging configuration
	AsyncWrite       bool             `PluginAttribute:"async,default=false"`
	BufferSize       int              `PluginAttribute:"bufferSize,default=10000"`
	BufferFullPolicy BufferFullPolicy `PluginAttribute:"bufferFullPolicy,default=Discard"`
}

// Start initializes the FileLogger according to AsyncWrite flag
// and then starts the underlying logger and its appenders.
func (f *RollingFileLogger) Start() error {
	if f.AsyncWrite {
		// Async mode: use AsyncLogger and AsyncRotateFileWriter
		return initRollingFileLogger(f, func(f *RollingFileLogger) Logger {
			return &AsyncLogger{
				LoggerBase:       f.LoggerBase,
				BufferSize:       f.BufferSize,
				BufferFullPolicy: f.BufferFullPolicy,
			}
		})
	} else {
		// Sync mode: use SyncLogger and SyncRotateFileWriter
		return initRollingFileLogger(f, func(f *RollingFileLogger) Logger {
			return &SyncLogger{
				LoggerBase: f.LoggerBase,
			}
		})
	}
}

// initRollingFileLogger is a generic helper to configure both synchronous and asynchronous FileLogger.
//   - fnAppender creates either SyncRotateFileWriter or AsyncRotateFileWriter.
//   - fnLogger creates either SyncLogger or AsyncLogger.
func initRollingFileLogger(
	f *RollingFileLogger,
	fnLogger func(f *RollingFileLogger) Logger,
) error {

	// Decide the maximum level for the normal log file.
	// If Separate is true, warning and above go to a separate .wf file.
	normalMaxLevel := MaxLevel
	if f.Separate {
		normalMaxLevel = WarnLevel
	}

	// Create appenders for the normal log file
	appenders := []*AppenderRef{
		{
			Appender: &RollingFileAppender{
				FileDir:       f.FileDir,
				FileName:      f.FileName,
				ClearHours:    f.ClearHours,
				RollingPolicy: f.RollingPolicy,
			},
			MinLevel: f.MinLevel,
			MaxLevel: normalMaxLevel,
		},
	}

	// Create appenders for warning and error logs if Separate is enabled
	if f.Separate {
		appenders = append(appenders, &AppenderRef{
			Appender: &RollingFileAppender{
				FileDir:       f.FileDir,
				FileName:      f.FileName + ".wf",
				ClearHours:    f.ClearHours,
				RollingPolicy: f.RollingPolicy,
			},
			MinLevel: normalMaxLevel,
			MaxLevel: f.MaxLevel,
		})
	}

	f.logger = fnLogger(f)

	// Attach the final appender to the logger
	switch x := f.logger.(type) {
	case *SyncLogger:
		x.AppenderRefs = appenders
	case *AsyncLogger:
		x.AppenderRefs = appenders
	default: // for linter
	}

	// Start the underlying logger (and all appenders)
	return f.logger.Start()
}

func (f *RollingFileLogger) Append(e *Event) {
	f.logger.Append(e)
}

func (f *RollingFileLogger) Write(b []byte) {
	f.logger.Write(b)
}

func (f *RollingFileLogger) Stop() {

}
