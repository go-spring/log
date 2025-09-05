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
	"fmt"
	"strings"
)

func init() {
	RegisterConverter(ParseLevel)
}

var (
	NoneLevel  = RegisterLevel(0, "NONE")    // No logging
	TraceLevel = RegisterLevel(100, "TRACE") // Very detailed logging, typically for debugging at a granular level
	DebugLevel = RegisterLevel(200, "DEBUG") // Debugging information
	InfoLevel  = RegisterLevel(300, "INFO")  // General informational messages
	WarnLevel  = RegisterLevel(400, "WARN")  // Warnings that may indicate a potential problem
	ErrorLevel = RegisterLevel(500, "ERROR") // Errors that allow the application to continue running
	PanicLevel = RegisterLevel(600, "PANIC") // Severe issues that may lead to a panic
	FatalLevel = RegisterLevel(700, "FATAL") // Critical issues that will cause application termination
)

var levels = map[string]Level{}

// Level represents a logging severity level.
// Each level has a numeric code (for ordering/comparison)
// and a string name (for readability).
type Level struct {
	code int32
	name string
}

// Code returns the numeric value of the Level.
func (l Level) Code() int32 {
	return l.code
}

// String returns the string representation of the Level.
func (l Level) String() string {
	return l.name
}

// RegisterLevel creates a new logging Level with the given code and name.
// It also stores the level in the global levels map for later lookup.
func RegisterLevel(code int32, name string) Level {
	l := Level{code: code, name: strings.ToUpper(name)}
	levels[l.name] = l
	return l
}

// ParseLevel converts a string into a Level object (case-insensitive).
// Returns an error if the string does not match any registered Level.
func ParseLevel(s string) (Level, error) {
	l, ok := levels[strings.ToUpper(s)]
	if !ok {
		return NoneLevel, fmt.Errorf("invalid level %s", s)
	}
	return l, nil
}
