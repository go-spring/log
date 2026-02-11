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
	RegisterPlugin[AsyncLogger]("AsyncLogger")
	RegisterPlugin[DiscardLogger]("DiscardLogger")
	RegisterPlugin[ConsoleLogger]("ConsoleLogger")
	RegisterPlugin[FileLogger]("FileLogger")
	RegisterPlugin[RollingFileLogger]("RollingFileLogger")
}

// Logger is the interface implemented by all loggers.
type Logger interface {
	Appender
	GetTags() string
	GetLevel() LevelRange
}

// AppenderRef represents a reference to an Appender by name.
// The actual Appender is resolved and injected during configuration.
type AppenderRef struct {
	Appender
	Ref   string     `PluginAttribute:"ref"`
	Level LevelRange `PluginAttribute:"level,default="`
}

// Append forwards the event to each child appender.
func (c *AppenderRef) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		c.Appender.Append(e)
	}
}

// Write forwards raw bytes to each child appender.
func (c *AppenderRef) Write(b []byte) {
	c.Appender.Write(b)
}

// LoggerBase contains fields shared by all logger configurations.
type LoggerBase struct {
	Name   string     `PluginAttribute:"name"`           // Logger name
	Tags   string     `PluginAttribute:"tags,default="`  // Optional tags
	Level  LevelRange `PluginAttribute:"level,default="` // Logger level range
	Layout Layout     `PluginElement:"layout?"`          // Layout for formatting logs
}

// GetName returns the name of the logger.
func (c *LoggerBase) GetName() string {
	return c.Name
}

// GetTags returns the tags of the logger.
func (c *LoggerBase) GetTags() string {
	return c.Tags
}

// GetLevel returns the level range of the logger.
func (c *LoggerBase) GetLevel() LevelRange {
	return c.Level
}

var (
	_ Logger = (*DiscardLogger)(nil)
	_ Logger = (*ConsoleLogger)(nil)
	_ Logger = (*SyncLogger)(nil)
	_ Logger = (*AsyncLogger)(nil)
	_ Logger = (*FileLogger)(nil)
	_ Logger = (*RollingFileLogger)(nil)
)

// DiscardLogger ignores all log events (no-op).
type DiscardLogger struct {
	LoggerBase
	DiscardAppender
}

// ConsoleLogger writes log events to standard output.
type ConsoleLogger struct {
	LoggerBase
	ConsoleAppender
}

// Append sends the event to the console if the level is enabled.
func (c *ConsoleLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		c.ConsoleAppender.Append(e)
	}
}

// FileLogger writes log events to a file.
type FileLogger struct {
	LoggerBase
	FileAppender
}

// Append writes the event to the file if the level is enabled.
func (c *FileLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		c.FileAppender.Append(e)
	}
}

// AppenderRefs represents a collection of AppenderRef objects.
type AppenderRefs struct {
	AppenderRefs []*AppenderRef `PluginElement:"appenderRef"`
}

// Append forwards events to each child appender.
func (c *AppenderRefs) sendToAppenders(e *Event) {
	for _, r := range c.AppenderRefs {
		if r.Level.Enable(e.Level) {
			r.Append(e)
		}
	}
}

// Write forwards raw bytes to each child appender.
func (c *AppenderRefs) writeToAppenders(l Level, b []byte) {
	for _, r := range c.AppenderRefs {
		if r.Level.Enable(l) {
			r.Write(b)
		}
	}
}

// SyncLogger is a synchronous logger that immediately forwards events to appenders.
type SyncLogger struct {
	LoggerBase
	AppenderRefs
}

func (c *SyncLogger) Start() error { return nil }
func (c *SyncLogger) Stop()        {}

// Append sends the event directly to appenders (blocking).
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

// Write writes raw bytes directly to appenders.
func (c *SyncLogger) Write(b []byte) {
	c.writeToAppenders(MaxLevel, b)
}

func (c *SyncLogger) ConcurrentSafe() bool { return true }

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
		return -1, errutil.Explain(nil, "invalid BufferFullPolicy %s", s)
	}
}

// AsyncLogger is an asynchronous logger that buffers events
// and processes them in a dedicated background goroutine.
type AsyncLogger struct {
	LoggerBase
	AppenderRefs

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

// Start initializes the AsyncLogger and launches the worker goroutine.
func (c *AsyncLogger) Start() error {
	if c.BufferSize < 100 {
		return errutil.Explain(nil, "bufferSize is too small")
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
				if c.Layout == nil {
					c.sendToAppenders(x)
				} else {
					c.writeToAppenders(x.Level, c.Layout.ToBytes(x))
				}
				PutEvent(x)
			case []byte:
				c.writeToAppenders(MaxLevel, x)
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
		c.buf <- v // Block until space is available
	case BufferFullPolicyDiscard:
		atomic.AddInt64(&c.discardCounter, 1)
		if e, ok := v.(*Event); ok {
			PutEvent(e)
		}
		return
	default: // for linter
	}
}

// Stop gracefully shuts down the async logger.
func (c *AsyncLogger) Stop() {
	c.buf <- c.stop
	<-c.wait
	close(c.buf)
}

func (c *AsyncLogger) ConcurrentSafe() bool { return true }

// RollingFileLogger writes log events to files with optional rotation, separation, and async behavior.
type RollingFileLogger struct {
	LoggerBase
	logger    Logger
	appenders []*AppenderRef

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

// Start initializes the RollingFileLogger, either synchronous or asynchronous.
func (f *RollingFileLogger) Start() error {
	if f.AsyncWrite {
		return initRollingFileLogger(f, func(f *RollingFileLogger) Logger {
			return &AsyncLogger{
				LoggerBase:       f.LoggerBase,
				BufferSize:       f.BufferSize,
				BufferFullPolicy: f.BufferFullPolicy,
			}
		})
	}
	return initRollingFileLogger(f, func(f *RollingFileLogger) Logger {
		return &SyncLogger{
			LoggerBase: f.LoggerBase,
		}
	})
}

// initRollingFileLogger is a helper to configure appenders for RollingFileLogger.
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

	// Create appenders for warning and error logs if Separate is enabled
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

	// Attach the final appender to the logger
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
	return nil
}

// Append forwards the event to the underlying logger.
func (f *RollingFileLogger) Append(e *Event) {
	f.logger.Append(e)
}

// Write forwards raw bytes to the underlying logger.
func (f *RollingFileLogger) Write(b []byte) {
	f.logger.Write(b)
}

// Stop stops all appenders.
func (f *RollingFileLogger) Stop() {
	for _, a := range f.appenders {
		a.Stop()
	}
}

func (f *RollingFileLogger) ConcurrentSafe() bool { return true }
