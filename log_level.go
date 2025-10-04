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
	"strings"

	"github.com/go-spring/spring-base/util"
)

func init() {
	RegisterConverter(ParseLevel)
}

var (
	NoneLevel  = RegisterLevel(0, "NONE")    // No logging
	TraceLevel = RegisterLevel(100, "TRACE") // Very detailed logging, typically used for debugging at a granular level
	DebugLevel = RegisterLevel(200, "DEBUG") // Debugging information useful during development
	InfoLevel  = RegisterLevel(300, "INFO")  // General informational messages about application progress
	WarnLevel  = RegisterLevel(400, "WARN")  // Warnings about potential issues or unusual situations
	ErrorLevel = RegisterLevel(500, "ERROR") // Errors that allow the application to continue running
	PanicLevel = RegisterLevel(600, "PANIC") // Severe issues that may cause a panic in the application
	FatalLevel = RegisterLevel(700, "FATAL") // Critical issues that will terminate the application
	MaxLevel   = RegisterLevel(999, "MAX")   // Maximum level (used as the upper bound for comparisons)
)

var levelRegistry = map[string]Level{}

// Level represents a logging severity level. Each level
// has a numeric code (for comparison) and a string name (for display).
type Level struct {
	code int32
	name string
}

// Code returns the numeric code of the Level.
// Levels with higher codes represent higher severity.
func (l Level) Code() int32 {
	return l.code
}

// String returns the string representation of the Level (e.g., "INFO").
func (l Level) String() string {
	return l.name
}

// RegisterLevel defines a new logging Level with the given code and name.
// The Level is also stored in the global levels map for string lookups.
// Name is normalized to uppercase for consistency.
func RegisterLevel(code int32, name string) Level {
	l := Level{code: code, name: strings.ToUpper(name)}
	levelRegistry[l.name] = l
	return l
}

// ParseLevel converts a string into a Level (case-insensitive).
// Returns an error if the string does not match any registered Level.
func ParseLevel(s string) (Level, error) {
	l, ok := levelRegistry[strings.ToUpper(s)]
	if !ok {
		return NoneLevel, util.FormatError(nil, "invalid log level: %q", s)
	}
	return l, nil
}
