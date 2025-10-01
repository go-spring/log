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
	RegisterPlugin[AppenderRef]("AppenderRef", PluginTypeAppenderRef)
	RegisterPlugin[SyncLogger]("Root", PluginTypeRoot)
	RegisterPlugin[AsyncLogger]("AsyncRoot", PluginTypeAsyncRoot)
	RegisterPlugin[SyncLogger]("Logger", PluginTypeLogger)
	RegisterPlugin[AsyncLogger]("AsyncLogger", PluginTypeAsyncLogger)
}

// Logger is the interface implemented by all loggers.
type Logger interface {
	Lifecycle                          // Start/Stop methods
	GetName() string                   // Get the name of the logger
	Publish(e *Event)                  // Send events to appenders
	EnableLevel(level Level) bool      // Whether a log level is enabled
	Write(b []byte) (n int, err error) // Write raw bytes to appenders
}

// AppenderRef represents a reference to an appender by name.
// The actual appender is resolved and injected later during configuration.
type AppenderRef struct {
	Ref      string   `PluginAttribute:"ref"`
	appender Appender // Resolved appender instance
}

// LoggerBase contains fields shared by all logger configurations.
type LoggerBase struct {
	Name         string         `PluginAttribute:"name"`          // Logger name
	Level        Level          `PluginAttribute:"level"`         // Log level
	Tags         string         `PluginAttribute:"tags,default="` // Optional tags
	AppenderRefs []*AppenderRef `PluginElement:"AppenderRef"`     // Attached appenders
}

// GetName returns the name of the logger.
func (c *LoggerBase) GetName() string {
	return c.Name
}

// callAppenders sends a log event to all configured appenders.
func (c *LoggerBase) callAppenders(e *Event) {
	for _, r := range c.AppenderRefs {
		r.appender.Append(e)
	}
}

// writeAppenders writes raw bytes directly to all appenders.
func (c *LoggerBase) writeAppenders(b []byte) {
	for _, r := range c.AppenderRefs {
		r.appender.Write(b)
	}
}

// EnableLevel checks if the given log level is enabled for this logger.
func (c *LoggerBase) EnableLevel(level Level) bool {
	return level.code >= c.Level.code
}

// SyncLogger is a synchronous logger that immediately forwards events to appenders.
type SyncLogger struct {
	LoggerBase
}

func (c *SyncLogger) Start() error { return nil }
func (c *SyncLogger) Stop()        {}

// Publish sends the event directly to appenders (blocking).
func (c *SyncLogger) Publish(e *Event) {
	c.callAppenders(e)
	PutEvent(e) // Return event to the pool
}

// Write writes raw bytes directly to appenders.
func (c *SyncLogger) Write(b []byte) (n int, err error) {
	c.writeAppenders(b)
	return len(b), nil
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
				c.callAppenders(x)
				PutEvent(x)
			case []byte:
				c.writeAppenders(x)
			default: // for linter
			}
		}
		close(c.wait)
	}()
	return nil
}

// Publish enqueues a log event into the buffer.
// Behavior on full buffer depends on BufferFullPolicy.
func (c *AsyncLogger) Publish(e *Event) {
	select {
	case c.buf <- e:
	default:
		c.onBufferFull(e)
	}
}

// Write enqueues raw bytes into the buffer.
// Behavior on full buffer depends on BufferFullPolicy.
func (c *AsyncLogger) Write(b []byte) (n int, err error) {
	select {
	case c.buf <- b:
	default:
		c.onBufferFull(b)
	}
	return len(b), nil
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
