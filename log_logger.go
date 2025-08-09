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
)

// loggerMap safely holds LoggerWrapper instances keyed by their names.
var loggerMap = map[string]*LoggerWrapper{}

// LoggerWrapper wraps a Logger instance, allowing atomic replacement of the logger.
type LoggerWrapper struct {
	logger atomic.Value
	name   string
}

// Write passes the byte slice to the currently set logger's Write method.
func (m *LoggerWrapper) Write(b []byte) {
	m.getLogger().Write(b)
}

// getLogger retrieves the currently stored Logger instance.
func (m *LoggerWrapper) getLogger() Logger {
	return m.logger.Load().(LoggerHolder)
}

// setLogger updates the Logger instance atomically.
func (m *LoggerWrapper) setLogger(logger Logger) {
	m.logger.Store(LoggerHolder{logger})
}

// GetLogger retrieves an existing LoggerWrapper by name or creates a new one.
// It panics if the global initialization phase has completed.
func GetLogger(name string) *LoggerWrapper {
	if global.init.Load() {
		panic("log refresh already done")
	}
	m, ok := loggerMap[name]
	if !ok {
		m = &LoggerWrapper{name: name}
		loggerMap[name] = m
	}
	return m
}

// GetRootLogger retrieves the root LoggerWrapper.
func GetRootLogger() *LoggerWrapper {
	return GetLogger(rootLoggerName)
}
