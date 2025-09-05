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
	"errors"
	"fmt"
	"sync/atomic"
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

// AppenderRef represents a reference to an appender by name,
// which will be resolved and bound later.
type AppenderRef struct {
	Ref      string `PluginAttribute:"ref"`
	appender Appender
}

// BaseLogger contains shared fields for all logger configurations.
type BaseLogger struct {
	Name         string         `PluginAttribute:"name"`
	Level        Level          `PluginAttribute:"level"`
	Tags         string         `PluginAttribute:"tags,default="`
	AppenderRefs []*AppenderRef `PluginElement:"AppenderRef"`
}

// GetName returns the name of the logger.
func (c *BaseLogger) GetName() string {
	return c.Name
}

// callAppenders sends the event to all configured appenders.
func (c *BaseLogger) callAppenders(e *Event) {
	for _, r := range c.AppenderRefs {
		r.appender.Append(e)
	}
}

// writeAppenders writes the raw bytes directly to the appenders.
func (c *BaseLogger) writeAppenders(b []byte) {
	for _, r := range c.AppenderRefs {
		r.appender.Write(b)
	}
}

// EnableLevel returns true if the specified log level is enabled.
func (c *BaseLogger) EnableLevel(level Level) bool {
	return level.code >= c.Level.code
}

// SyncLogger is a synchronous logger configuration.
type SyncLogger struct {
	BaseLogger
}

func (c *SyncLogger) Start() error { return nil }
func (c *SyncLogger) Stop()        {}

// Publish sends the event directly to the appenders.
func (c *SyncLogger) Publish(e *Event) {
	c.callAppenders(e)
	PutEvent(e)
}

// Write writes the raw bytes directly to the appenders.
func (c *SyncLogger) Write(b []byte) (n int, err error) {
	c.writeAppenders(b)
	return len(b), nil
}

// BufferFullPolicy specifies the behavior when the buffer is full.
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
		return -1, fmt.Errorf("invalid BufferFullPolicy %s", s)
	}
}

// AsyncLogger is an asynchronous logger configuration.
// It buffers log events and processes them in a separate goroutine.
type AsyncLogger struct {
	BaseLogger
	BufferSize       int              `PluginAttribute:"bufferSize,default=10000"`
	BufferFullPolicy BufferFullPolicy `PluginAttribute:"bufferFullPolicy,default=Discard"`

	buf  chan any      // Channel buffer for log events
	wait chan struct{} // Channel for waiting for the worker goroutine to finish
	stop *Event        // Event for stopping the worker goroutine

	discardCounter int64 // Counter for discarded events
}

// GetDiscardCounter returns the count of discarded events.
func (c *AsyncLogger) GetDiscardCounter() int64 {
	return atomic.LoadInt64(&c.discardCounter)
}

// Start initializes the asynchronous logger and starts its worker goroutine.
func (c *AsyncLogger) Start() error {
	if c.BufferSize < 100 {
		return errors.New("bufferSize is too small")
	}
	c.buf = make(chan any, c.BufferSize)
	c.wait = make(chan struct{})
	c.stop = &Event{}

	// Launch a background goroutine to process events
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

// Publish places the event in the buffer if there's space; drops it otherwise.
func (c *AsyncLogger) Publish(e *Event) {
	select {
	case c.buf <- e:
	default:
		c.onBufferFull(e)
	}
}

// Write writes the raw bytes directly to the appenders.
func (c *AsyncLogger) Write(b []byte) (n int, err error) {
	select {
	case c.buf <- b:
	default:
		c.onBufferFull(b)
	}
	return len(b), nil
}

// onBufferFull is called when the buffer is full.
func (c *AsyncLogger) onBufferFull(v any) {
	switch c.BufferFullPolicy {
	case BufferFullPolicyDiscardOldest:
		var exit bool
		for {
			select {
			case c.buf <- v:
				exit = true
			default:
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
		c.buf <- v
	case BufferFullPolicyDiscard:
		atomic.AddInt64(&c.discardCounter, 1)
		if e, ok := v.(*Event); ok {
			PutEvent(e)
		}
		return
	default: // for linter
	}
}

// Stop shuts down the asynchronous logger and waits for the worker goroutine to finish.
func (c *AsyncLogger) Stop() {
	c.buf <- c.stop
	<-c.wait
	close(c.buf)
}
