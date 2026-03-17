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
	Lifecycle             // Start/Stop methods for resource management
	GetName() string      // Appender's name
	GetTags() []string    // Tags associated with this logger
	GetLevel() LevelRange // Level range handled by this logger
	Append(e *Event)      // Handles writing a log event
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

// AppenderRefs is implemented by loggers that support appender references.
type AppenderRefs interface {
	// GetAppenderRefs returns the logger's synchronization mode
	// and the list of appender references.
	//
	// In sync mode, appenders must ensure thread safety.
	// In async mode, they do not require strict thread safety.
	GetAppenderRefs() (syncMode bool, _ []*AppenderRef)
}

// LoggerBase contains fields shared by all logger configurations.
type LoggerBase struct {
	Name  string     `PluginAttribute:"name"`           // Logger name
	Tags  []string   `PluginAttribute:"tags,default="`  // Optional tags associated with this logger
	Level LevelRange `PluginAttribute:"level,default="` // Level range handled by this logger
}

func (c *LoggerBase) GetName() string      { return c.Name }
func (c *LoggerBase) GetTags() []string    { return c.Tags }
func (c *LoggerBase) GetLevel() LevelRange { return c.Level }

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
	AppenderRefs []*AppenderRef `PluginElement:"appenderRef"`
}

// GetAppenderRefs returns true for sync mode and the appender refs.
func (c *SyncLogger) GetAppenderRefs() (syncMode bool, _ []*AppenderRef) {
	return true, c.AppenderRefs
}

func (c *SyncLogger) Start() error { return nil }
func (c *SyncLogger) Stop()        {}

// Append sends the event directly to appenders.
func (c *SyncLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		for _, r := range c.AppenderRefs {
			r.Append(e)
		}
	}
	e.Reset()
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

// AsyncLogger is an asynchronous logger that buffers events in a channel
// and processes them in a background goroutine.
type AsyncLogger struct {
	LoggerBase
	AppenderRefs     []*AppenderRef   `PluginElement:"appenderRef"`
	BufferSize       int              `PluginAttribute:"bufferSize,default=10000"`
	BufferFullPolicy BufferFullPolicy `PluginAttribute:"bufferFullPolicy,default=Discard"`

	buf  chan *Event   // Channel buffering events
	wait chan struct{} // Waiting for the worker goroutine to finish
	stop *Event        // Sentinel value used to signal shutdown

	discardCounter int64 // Count of discarded events
}

// GetDiscardCounter returns the total number of discarded events.
func (c *AsyncLogger) GetDiscardCounter() int64 {
	return atomic.LoadInt64(&c.discardCounter)
}

// GetAppenderRefs returns false for async mode and the appender references.
func (c *AsyncLogger) GetAppenderRefs() (syncMode bool, _ []*AppenderRef) {
	return false, c.AppenderRefs
}

// Start initializes the buffer and starts the background worker goroutine.
func (c *AsyncLogger) Start() error {
	if c.BufferSize < 100 {
		return errutil.Explain(nil, "bufferSize is too small") // todo 详细错误信息
	}

	c.buf = make(chan *Event, c.BufferSize)
	c.wait = make(chan struct{})
	c.stop = &Event{}

	// Worker goroutine that processes events from the buffer
	// and forwards them to appenders.
	go func() {
		for e := range c.buf {
			// 尽可能保证日志写完
			if e == c.stop {
				break
			}
			for _, r := range c.AppenderRefs {
				r.Append(e)
			}
			e.Reset()
		}
		close(c.wait)
	}()
	return nil
}

// Stop gracefully shuts down the AsyncLogger, ensuring that any remaining buffered events
// are processed before the background worker goroutine exits.
func (c *AsyncLogger) Stop() {
	// To ensure that more log events are written, a blocking approach is used here.
	c.buf <- c.stop
	<-c.wait
	close(c.buf)
}

// Append enqueues a log event into the async buffer.
// Behavior on full buffer depends on BufferFullPolicy.
func (c *AsyncLogger) Append(e *Event) {
	if !c.Level.Enable(e.Level) {
		e.Reset()
		return
	}

	select {
	case c.buf <- e:
		return
	default:
	}

	switch c.BufferFullPolicy {
	case BufferFullPolicyDiscardOldest:
		for {
			select {
			case x := <-c.buf: // Remove one element to make space
				atomic.AddInt64(&c.discardCounter, 1)
				x.Reset()
			default: // for linter
			}
			select {
			case c.buf <- e:
				return
			default: // for linter
			}
		}
	case BufferFullPolicyBlock:
		c.buf <- e // Block until space is available
	case BufferFullPolicyDiscard:
		atomic.AddInt64(&c.discardCounter, 1)
		e.Reset()
	default: // for linter
	}
}

// DiscardLogger ignores all log events (no-op).
type DiscardLogger struct {
	LoggerBase
}

func (d DiscardLogger) Start() error    { return nil }
func (d DiscardLogger) Stop()           {}
func (d DiscardLogger) Append(e *Event) { e.Reset() }

// ConsoleLogger writes log events to standard output.
type ConsoleLogger struct {
	LoggerBase
	appender *ConsoleAppender
	Layout   Layout `PluginElement:"layout,default=TextLayout"`
}

// Start initializes the console appender and starts it.
func (c *ConsoleLogger) Start() error {
	c.appender = &ConsoleAppender{
		AppenderBase: AppenderBase{
			Layout: c.Layout,
		},
	}
	// Append operation is not managed by the framework,
	// so we start the appender manually.
	return c.appender.Start()
}

// Stop stops the console appender manually.
func (c *ConsoleLogger) Stop() {
	// Appenders are not managed by the framework,
	// so they need to be manually stopped.
	c.appender.Stop()
}

// Append writes the event to the console if its level is enabled.
func (c *ConsoleLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		c.appender.Append(e)
	}
	e.Reset()
}

// FileLogger writes log events to a file.
type FileLogger struct {
	LoggerBase
	FileDir  string `PluginAttribute:"fileDir,default=./logs"`
	FileName string `PluginAttribute:"fileName"`
	Layout   Layout `PluginElement:"layout,default=TextLayout"`

	appender *FileAppender
}

// Start initializes the file appender and starts it.
func (c *FileLogger) Start() error {
	c.appender = &FileAppender{
		AppenderBase: AppenderBase{
			Layout: c.Layout,
		},
		FileDir:  c.FileDir,
		FileName: c.FileName,
	}
	// Append operation is not managed by the framework,
	// so we start the appender manually.
	return c.appender.Start()
}

// Stop stops the file appender manually.
func (c *FileLogger) Stop() {
	// Appenders are not managed by the framework,
	// so they need to be stopped manually.
	c.appender.Stop()
}

// Append writes the log event to the file if its level is enabled.
func (c *FileLogger) Append(e *Event) {
	if c.Level.Enable(e.Level) {
		c.appender.Append(e)
	}
	e.Reset()
}

// RollingFileLogger writes log events to log files with automatic file rotation.
// It can separate warning and error logs into a separate ".wf" file if configured.
// It supports both synchronous and asynchronous logging modes base on configuration.
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

	// Asynchronous logging options
	AsyncWrite       bool             `PluginAttribute:"async,default=false"`
	BufferSize       int              `PluginAttribute:"bufferSize,default=10000"`
	BufferFullPolicy BufferFullPolicy `PluginAttribute:"bufferFullPolicy,default=Discard"`
}

// Start initializes the internal logger and configures rolling file appenders.
// Depending on AsyncWrite, either SyncLogger or AsyncLogger will be used.
func (f *RollingFileLogger) Start() error {

	normalMaxLevel := MaxLevel
	if f.Separate {
		normalMaxLevel = WarnLevel
	}

	// Create the appender for the normal log file
	f.appenders = []*AppenderRef{
		{
			Appender: &RollingFileAppender{
				AppenderBase: AppenderBase{
					Layout: f.Layout,
				},
				FileDir:  f.FileDir,
				FileName: f.FileName,
				Interval: f.Interval,
				MaxAge:   f.MaxAge,
				Lock:     !f.AsyncWrite,
			},
			Level: LevelRange{
				MinLevel: f.Level.MinLevel,
				MaxLevel: normalMaxLevel,
			},
		},
	}

	if f.Separate {
		// Create the second appender for warning/error logs.
		f.appenders = append(f.appenders, &AppenderRef{
			Appender: &RollingFileAppender{
				AppenderBase: AppenderBase{
					Layout: f.Layout,
				},
				FileDir:  f.FileDir,
				FileName: f.FileName + ".wf",
				Interval: f.Interval,
				MaxAge:   f.MaxAge,
				Lock:     !f.AsyncWrite,
			},
			Level: LevelRange{
				MinLevel: normalMaxLevel,
				MaxLevel: f.Level.MaxLevel,
			},
		})
	}

	// Initialize the underlay logger
	if f.AsyncWrite {
		f.logger = &AsyncLogger{
			LoggerBase:       f.LoggerBase,
			AppenderRefs:     f.appenders,
			BufferSize:       f.BufferSize,
			BufferFullPolicy: f.BufferFullPolicy,
		}
	} else {
		f.logger = &SyncLogger{
			LoggerBase:   f.LoggerBase,
			AppenderRefs: f.appenders,
		}
	}

	// Start the appenders manually (since they aren't managed by the framework)
	for _, a := range f.appenders {
		if err := a.Start(); err != nil {
			return err
		}
	}
	return f.logger.Start()
}

// Stop stops all appenders managed by this logger
// and gracefully shuts down the logger.
func (f *RollingFileLogger) Stop() {
	// Stop the logger first to ensure buffered events are written
	f.logger.Stop()
	// Appenders are not managed by the framework,
	// so they need to be stopped manually.
	for _, a := range f.appenders {
		a.Stop()
	}
}

// Append forwards the log event to the internal logger.
func (f *RollingFileLogger) Append(e *Event) {
	f.logger.Append(e)
}
