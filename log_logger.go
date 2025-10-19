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

// loggerMap stores LoggerWrapper instances keyed by their names.
// Note: This map is not concurrency-safe. It is expected to be modified
// only during the initialization phase.
var loggerMap = map[string]*LoggerWrapper{}

// LoggerWrapper wraps a Logger instance and allows atomic replacement
// of the underlying Logger at runtime. This ensures that concurrent
// readers always see a consistent Logger reference without needing locks.
type LoggerWrapper struct {
	name   string // Logical name of the logger
	logger Logger // Underlying Logger instance
}

// Write forwards the given byte slice to the currently active Logger.
// It implements the io.Writer interface, allowing LoggerWrapper to be
// used anywhere an io.Writer is expected.
func (m *LoggerWrapper) Write(b []byte) (n int, err error) {
	m.logger.Write(b)
	return len(b), nil
}

// GetLogger retrieves an existing LoggerWrapper by name,
// or creates a new one if it does not exist yet.
//
// This function must be called only during the initialization phase.
// It panics if called after global.init is set, indicating that
// the logging system has already been finalized.
func GetLogger(name string) *LoggerWrapper {
	if global.init {
		panic("log refresh already done")
	}
	m, ok := loggerMap[name]
	if !ok {
		m = &LoggerWrapper{name: name}
		loggerMap[name] = m
	}
	return m
}
