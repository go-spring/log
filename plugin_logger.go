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
	"time"

	"github.com/go-spring/stdlib/errutil"
)

func init() {
	RegisterConverter[BufferFullPolicy](ParseBufferFullPolicy)

	RegisterPlugin[SyncLogger]("Logger")
	RegisterPlugin[SyncLogger]("SyncLogger")
	RegisterPlugin[AsyncLogger]("AsyncLogger")
	RegisterPlugin[DiscardLogger]("DiscardLogger")
	RegisterPlugin[ConsoleLogger]("ConsoleLogger")
	RegisterPlugin[FileLogger]("FileLogger")
	RegisterPlugin[RollingFileLogger]("RollingFileLogger")
}

// Logger is the interface implemented by all logger implementations.
// A Logger receives log events and forwards them to one or more appenders.
type Logger interface {
	Lifecycle                         // Start/Stop methods for resource management
	GetName() string                  // Appender's name
	GetTags() []string                // Tags associated with this logger
	GetLevel() LevelRange             // Level range handled by this logger
	Append(e *Event)                  // Handles writing a log event
	WriteLevel(level Level, b []byte) // Directly writes a byte slice
}

// AppenderRef represents a reference to an Appender by name.
// During configuration loading, the Ref field is resolved and the
// corresponding Appender instance is injected into the Appender field.
//
// Level optionally restricts the level range forwarded to this appender.
type AppenderRef struct {
	Appender
	Ref   string     `PluginAttribute:"ref"`
	Level LevelRange `PluginAttribute:"level,default="`
}

// Append forwards the event to the referenced appender if the level matches.
func (c *AppenderRef) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		c.Appender.Append(e)
	}
}

// WriteLevel forwards raw bytes to the referenced appender.
func (c *AppenderRef) WriteLevel(level Level, b []byte) {
	if c.Level.Enable(level) {
		c.Appender.Write(b)
	}
}

// LoggerBase contains fields shared by all logger configurations.
type LoggerBase struct {
	Name  string     `PluginAttribute:"name"`           // Logger name
	Tags  []string   `PluginAttribute:"tags,default="`  // Optional tags associated with this logger
	Level LevelRange `PluginAttribute:"level,default="` // Level range handled by this logger
}

// GetName returns the name of the logger.
func (c *LoggerBase) GetName() string {
	return c.Name
}

// GetTags returns the tags associated with the logger.
func (c *LoggerBase) GetTags() []string {
	return c.Tags
}

// GetLevel returns the level range supported by the logger.
func (c *LoggerBase) GetLevel() LevelRange {
	return c.Level
}

// AppenderRefs represents a collection of AppenderRef references.
type AppenderRefs struct {
	AppenderRefs []*AppenderRef `PluginElement:"appenderRef"`
}

// sendToAppenders forwards the event to each referenced appender
// whose level range allows the provided level.
func (c *AppenderRefs) sendToAppenders(e *Event) {
	for _, r := range c.AppenderRefs {
		if r.Level.Enable(e.Level) {
			r.Append(e)
		}
	}
}

// writeToAppenders forwards raw bytes to each referenced appender
// whose level range allows the provided level.
func (c *AppenderRefs) writeToAppenders(l Level, b []byte) {
	for _, r := range c.AppenderRefs {
		if r.Level.Enable(l) {
			r.Write(b)
		}
	}
}

var (
	_ Logger = (*DiscardLogger)(nil)
	_ Logger = (*ConsoleLogger)(nil)
	_ Logger = (*SyncLogger)(nil)
	_ Logger = (*AsyncLogger)(nil)
	_ Logger = (*FileLogger)(nil)
	_ Logger = (*RollingFileLogger)(nil)
)

// SyncLogger is a synchronous logger that forwards events to appenders
// immediately in the caller goroutine.
type SyncLogger struct {
	LoggerBase
	AppenderRefs
	Layout Layout `PluginElement:"layout?"` // Optional layout used to format log events
}

func (c *SyncLogger) Start() error { return nil }
func (c *SyncLogger) Stop()        {}

// Append sends the event directly to appenders. If a layout is configured,
// the event is formatted into bytes before forwarding.
func (c *SyncLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		if c.Layout == nil {
			c.sendToAppenders(e)
		} else {
			c.writeToAppenders(e.Level, c.Layout.ToBytes(e))
		}
	}
	PutEvent(e) // Return event to the pool
}

// WriteLevel forwards raw bytes to appenders.
func (c *SyncLogger) WriteLevel(level Level, b []byte) {
	c.writeToAppenders(level, b)
}

// BufferFullPolicy specifies how AsyncLogger behaves when its buffer is full.
type BufferFullPolicy int

const (
	BufferFullPolicyBlock         = BufferFullPolicy(0) // Block until space is available
	BufferFullPolicyDiscard       = BufferFullPolicy(1) // Drop the new event or data
	BufferFullPolicyDiscardOldest = BufferFullPolicy(2) // Drop the oldest buffered item
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
		return -1, errutil.Explain(nil, "invalid BufferFullPolicy %s", s)
	}
}

// AsyncLogger is an asynchronous logger that buffers events and raw data
// in a channel and processes them in a dedicated background goroutine.
type AsyncLogger struct {
	LoggerBase
	AppenderRefs

	Layout           Layout           `PluginElement:"layout?"` // Optional layout used to format log events
	BufferSize       int              `PluginAttribute:"bufferSize,default=10000"`
	BufferFullPolicy BufferFullPolicy `PluginAttribute:"bufferFullPolicy,default=Discard"`

	buf  chan *Event   // Channel buffering events or raw byte data
	wait chan struct{} // Closed when the worker goroutine exits
	stop *Event        // Sentinel value used to signal shutdown

	discardCounter int64 // Number of discarded events or data items
}

// GetDiscardCounter returns the total number of discarded items.
func (c *AsyncLogger) GetDiscardCounter() int64 {
	return atomic.LoadInt64(&c.discardCounter)
}

// Start initializes the buffer and launches the background worker goroutine.
func (c *AsyncLogger) Start() error {
	if c.BufferSize < 100 {
		return errutil.Explain(nil, "bufferSize is too small")
	}

	c.buf = make(chan *Event, c.BufferSize)
	c.wait = make(chan struct{})
	c.stop = &Event{}

	// Worker goroutine that drains the buffer and forwards data to appenders.
	go func() {
		for v := range c.buf {
			// 尽可能保证日志写完
			if v == c.stop {
				break
			}
			if v.RawBytes != nil {
				c.writeToAppenders(MaxLevel, v.RawBytes)
			} else {
				if c.Layout == nil {
					c.sendToAppenders(v)
				} else {
					c.writeToAppenders(v.Level, c.Layout.ToBytes(v))
				}
			}
			PutEvent(v)
		}
		close(c.wait)
	}()
	return nil
}

// Append enqueues a log event into the async buffer.
// Behavior on full buffer depends on BufferFullPolicy.
func (c *AsyncLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		select {
		case c.buf <- e:
		default:
			c.onBufferFull(e)
		}
	} else {
		PutEvent(e)
	}
}

// WriteLevel enqueues raw bytes into the async buffer.
// Behavior on full buffer depends on BufferFullPolicy.
func (c *AsyncLogger) WriteLevel(level Level, b []byte) {
	e := GetEvent()
	e.Level = level
	e.RawBytes = b
	select {
	case c.buf <- e:
	default:
		c.onBufferFull(e)
	}
}

// onBufferFull handles buffer overflow according to BufferFullPolicy.
func (c *AsyncLogger) onBufferFull(v *Event) {
	switch c.BufferFullPolicy {
	case BufferFullPolicyDiscardOldest:
		for {
			select {
			case e := <-c.buf: // Remove one element to make space
				atomic.AddInt64(&c.discardCounter, 1)
				PutEvent(e)
			default: // for linter
			}
			select {
			case c.buf <- v:
				return
			default: // for linter
			}
		}
	case BufferFullPolicyBlock:
		c.buf <- v // Block until space is available
	case BufferFullPolicyDiscard:
		atomic.AddInt64(&c.discardCounter, 1)
		PutEvent(v)
	default: // for linter
	}
}

// Stop gracefully shuts down the async logger and waits for the worker to exit.
func (c *AsyncLogger) Stop() {
	c.buf <- c.stop
	<-c.wait
	close(c.buf)
}

// DiscardLogger ignores all log events (no-op).
type DiscardLogger struct {
	LoggerBase
}

func (d DiscardLogger) Start() error                     { return nil }
func (d DiscardLogger) Stop()                            {}
func (d DiscardLogger) Append(e *Event)                  {}
func (d DiscardLogger) WriteLevel(level Level, b []byte) {}

// ConsoleLogger writes log events to standard output.
type ConsoleLogger struct {
	LoggerBase
	appender *ConsoleAppender
	Layout   Layout `PluginElement:"layout,default=TextLayout"`
}

func (c *ConsoleLogger) Start() error {
	c.appender = &ConsoleAppender{
		AppenderBase: AppenderBase{
			Layout: c.Layout,
		},
	}
	return c.appender.Start()
}

func (c *ConsoleLogger) Stop() {
	c.appender.Stop()
}

// Append sends the event to the console if the level is enabled.
func (c *ConsoleLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		c.appender.Append(e)
	}
}

func (c *ConsoleLogger) WriteLevel(level Level, b []byte) {
	if c.Level.Enable(level) {
		c.appender.Write(b)
	}
}

// FileLogger writes log events to a file.
type FileLogger struct {
	LoggerBase
	FileDir  string `PluginAttribute:"fileDir,default=./logs"`
	FileName string `PluginAttribute:"fileName"`
	Layout   Layout `PluginElement:"layout,default=TextLayout"`

	appender *FileAppender
}

func (c *FileLogger) Start() error {
	c.appender = &FileAppender{
		AppenderBase: AppenderBase{
			Layout: c.Layout,
		},
		FileDir:  c.FileDir,
		FileName: c.FileName,
	}
	return c.appender.Start()
}

func (c *FileLogger) Stop() {
	c.appender.Stop()
}

// Append writes the event to the file if the level is enabled.
func (c *FileLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		c.appender.Append(e)
	}
}

func (c *FileLogger) WriteLevel(level Level, b []byte) {
	if c.Level.Enable(level) {
		c.appender.Write(b)
	}
}

// RollingFileLogger writes log events to rolling files and optionally
// separates warning/error logs into a dedicated ".wf" file. It can operate
// in synchronous or asynchronous mode depending on configuration.
type RollingFileLogger struct {
	LoggerBase
	logger    Logger
	appenders []*AppenderRef

	Layout Layout `PluginElement:"layout,default=TextLayout"`

	// File output configuration
	FileDir  string `PluginAttribute:"fileDir,default=./logs"`
	FileName string `PluginAttribute:"fileName,default=app.log"`

	// If true, warning/error logs go to a separate .wf file.
	Separate bool `PluginAttribute:"separate,default=false"`

	// Rotation and retention
	Interval time.Duration `PluginAttribute:"interval,default=1h"`
	MaxAge   time.Duration `PluginAttribute:"maxAge,default=168h"`

	// Async logging options
	AsyncWrite       bool             `PluginAttribute:"async,default=false"`
	BufferSize       int              `PluginAttribute:"bufferSize,default=10000"`
	BufferFullPolicy BufferFullPolicy `PluginAttribute:"bufferFullPolicy,default=Discard"`
}

// Start initializes the internal logger and configures rolling file appenders.
// Depending on AsyncWrite, either SyncLogger or AsyncLogger will be used.
func (f *RollingFileLogger) Start() error {
	if f.AsyncWrite {
		return initRollingFileLogger(f, func(f *RollingFileLogger) Logger {
			return &AsyncLogger{
				LoggerBase:       f.LoggerBase,
				Layout:           f.Layout,
				BufferSize:       f.BufferSize,
				BufferFullPolicy: f.BufferFullPolicy,
			}
		})
	}
	return initRollingFileLogger(f, func(f *RollingFileLogger) Logger {
		return &SyncLogger{
			LoggerBase: f.LoggerBase,
			Layout:     f.Layout,
		}
	})
}

// initRollingFileLogger configures the appenders and underlying logger
// used by a RollingFileLogger.
//
// It creates one or two RollingFileAppender instances depending on whether
// warning/error logs are separated. If Separate is enabled, logs with level
// Warn and above are written to a dedicated ".wf" file.
//
// The provided factory function fnLogger determines whether a SyncLogger
// or AsyncLogger will be used internally.
func initRollingFileLogger(
	f *RollingFileLogger,
	fnLogger func(f *RollingFileLogger) Logger,
) error {

	// Determine the maximum level written to the normal log file.
	// If Separate is enabled, warning and above are redirected to the ".wf" file.
	normalMaxLevel := MaxLevel
	if f.Separate {
		normalMaxLevel = WarnLevel
	}

	// Create appenders for the normal log file.
	appenders := []*AppenderRef{
		{
			Appender: &RollingFileAppender{
				FileDir:  f.FileDir,
				FileName: f.FileName,
				Interval: f.Interval,
				MaxAge:   f.MaxAge,
				Lock:     false,
			},
			Level: LevelRange{
				MinLevel: f.Level.MinLevel,
				MaxLevel: normalMaxLevel,
			},
		},
	}

	// If Separate is enabled, create a second appender for warning/error logs.
	if f.Separate {
		appenders = append(appenders, &AppenderRef{
			Appender: &RollingFileAppender{
				FileDir:  f.FileDir,
				FileName: f.FileName + ".wf",
				Interval: f.Interval,
				MaxAge:   f.MaxAge,
				Lock:     false,
			},
			Level: LevelRange{
				MinLevel: normalMaxLevel,
				MaxLevel: f.Level.MaxLevel,
			},
		})
	}

	f.logger = fnLogger(f)

	// Attach appenders to the underlying logger.
	switch x := f.logger.(type) {
	case *SyncLogger:
		for _, a := range appenders {
			a.Appender.(*RollingFileAppender).Lock = true
		}
		x.AppenderRefs = AppenderRefs{AppenderRefs: appenders}
	case *AsyncLogger:
		x.AppenderRefs = AppenderRefs{AppenderRefs: appenders}
	default: // for linter
	}

	f.appenders = appenders
	for _, a := range appenders {
		if err := a.Start(); err != nil {
			return err
		}
	}
	return f.logger.Start()
}

// Append forwards the log event to the internal logger implementation.
func (f *RollingFileLogger) Append(e *Event) {
	f.logger.Append(e)
}

// WriteLevel forwards raw bytes directly to the internal logger implementation.
func (f *RollingFileLogger) WriteLevel(level Level, b []byte) {
	f.logger.WriteLevel(level, b)
}

// Stop stops all appenders managed by this logger.
// The underlying logger is expected to have already flushed
// any buffered data before this is called.
func (f *RollingFileLogger) Stop() {
	f.logger.Stop()
	for _, a := range f.appenders {
		a.Stop()
	}
}
