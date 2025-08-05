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
)

// OnDropEvent is a callback function that is called when an event is dropped.
var OnDropEvent func(logger string, v interface{})

func init() {
	RegisterPlugin[AppenderRef]("AppenderRef", PluginTypeAppenderRef)
	RegisterPlugin[SyncLogger]("Root", PluginTypeRoot)
	RegisterPlugin[AsyncLogger]("AsyncRoot", PluginTypeAsyncRoot)
	RegisterPlugin[SyncLogger]("Logger", PluginTypeLogger)
	RegisterPlugin[AsyncLogger]("AsyncLogger", PluginTypeAsyncLogger)
}

// Logger is the interface implemented by all logger configs.
type Logger interface {
	Lifecycle                     // Start/Stop methods
	GetName() string              // Get the name of the logger
	Publish(e *Event)             // Logic for sending events to appenders
	Write(b []byte)               // Direct write of raw bytes to appenders
	EnableLevel(level Level) bool // Whether a log level is enabled
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
	return level >= c.Level
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
func (c *SyncLogger) Write(b []byte) {
	c.writeAppenders(b)
}

// AsyncLogger is an asynchronous logger configuration.
// It buffers log events and processes them in a separate goroutine.
type AsyncLogger struct {
	BaseLogger
	BufferSize int `PluginAttribute:"bufferSize,default=10000"`

	buf  chan interface{} // Channel buffer for log events
	wait chan struct{}
}

// Start initializes the asynchronous logger and starts its worker goroutine.
func (c *AsyncLogger) Start() error {
	if c.BufferSize < 100 {
		return errors.New("bufferSize is too small")
	}
	c.buf = make(chan interface{}, c.BufferSize)
	c.wait = make(chan struct{})

	// Launch a background goroutine to process events
	go func() {
		for v := range c.buf {
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
		// Drop the event if the buffer is full
		if OnDropEvent != nil {
			OnDropEvent(c.Name, e)
		}
		PutEvent(e)
	}
}

// Write writes the raw bytes directly to the appenders.
func (c *AsyncLogger) Write(b []byte) {
	select {
	case c.buf <- b:
	default:
		// Drop the event if the buffer is full
		if OnDropEvent != nil {
			OnDropEvent(c.Name, b)
		}
	}
}

// Stop shuts down the asynchronous logger and waits for the worker goroutine to finish.
func (c *AsyncLogger) Stop() {
	close(c.buf)
	<-c.wait
}
