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

/*
Package log is a high-performance and extensible logging library designed specifically for the Go programming
language. It provides flexible and structured logging capabilities, including context field extraction, multi-level
logging configuration, and multiple output options, making it ideal for server-side applications.

## Core Concepts:

Tags:

Tags are a core concept in the log package used to categorize logs. By registering a tag via the `RegisterTag`
function, you can use regular expressions to match the user-defined tags. This approach allows for a unified API
for logging without explicitly creating logger instances. Even third-party libraries can write logs without
setting up a logger object.

Loggers:

A Logger is the object that actually handles the logging process. You can obtain a logger instance using the
`GetLogger` function, which is mainly provided for compatibility with legacy projects. This allows you to
directly retrieve a logger by its name and log pre-formatted messages using the `Write` function.

Context Field Extraction:

Contextual data can be extracted and included in log entries via configurable functions:
- `log.StringFromContext`: Extracts a string value (e.g., a request ID) from the context.
- `log.FieldsFromContext`: Returns a list of structured fields from the context, such as trace IDs or user IDs.

Configuration from File:

The `log.RefreshFile` function allows loading the logger's configuration from an external file (e.g., XML or JSON).

Logger Initialization and Logging:

- Using `GetLogger`, you can fetch a logger instance (often for compatibility with older systems).
- You can also register custom tags using `RegisterTag` to classify logs according to your needs.

Logging Messages:

The package provides various logging functions, such as `Tracef`, `Debugf`, `Infof`, `Warnf`, etc.,
which log messages at different levels (e.g., Trace, Debug, Info, Warn).
These functions can either accept structured fields or formatted messages.

Structured Logging:

The logger also supports structured logging, where fields are captured as key-value pairs and logged with the message.
The fields can be provided directly in the log functions or through a map.

## Examples:

Using a pre-registered tag:

	log.Tracef(ctx, TagRequestOut, "hello %s", "world")
	log.Debugf(ctx, TagRequestOut, "hello %s", "world")
	log.Infof(ctx, TagRequestIn, "hello %s", "world")
	log.Warnf(ctx, TagRequestIn, "hello %s", "world")
	log.Errorf(ctx, TagRequestIn, "hello %s", "world")
	log.Panicf(ctx, TagRequestIn, "hello %s", "world")
	log.Fatalf(ctx, TagRequestIn, "hello %s", "world")

Using structured fields:

	log.Trace(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	log.Error(ctx, TagRequestIn, log.FieldsFromMap(map[string]any{
		"key1": "value1",
		"key2": "value2",
	}))
*/
package log

import (
	"context"
	"time"
)

var (
	// TagAppDef is the default tag used for application logs.
	TagAppDef = RegisterAppTag("def", "")

	// TagBizDef is the default tag used for business-related logs.
	TagBizDef = RegisterBizTag("def", "")
)

var (
	// TimeNow is a function that can be overridden to provide custom
	// timestamp behavior, e.g., for testing or mocking.
	TimeNow func(ctx context.Context) time.Time

	// StringFromContext allows extraction of a string (e.g., trace ID) from the context.
	StringFromContext func(ctx context.Context) string

	// FieldsFromContext allows extraction of structured fields (e.g., trace ID, span ID) from the context.
	FieldsFromContext func(ctx context.Context) []Field
)

// Trace logs a message at TraceLevel using a tag and a lazy field-generating function.
func Trace(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.getLogger().EnableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, 2, fn()...)
	}
}

// Tracef logs a message at TraceLevel using a tag and a formatted message.
func Tracef(ctx context.Context, tag *Tag, format string, args ...any) {
	if tag.getLogger().EnableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, 2, Msgf(format, args...))
	}
}

// Debug logs a message at DebugLevel using a tag and a lazy field-generating function.
func Debug(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.getLogger().EnableLevel(DebugLevel) {
		Record(ctx, DebugLevel, tag, 2, fn()...)
	}
}

// Debugf logs a message at DebugLevel using a tag and a formatted message.
func Debugf(ctx context.Context, tag *Tag, format string, args ...any) {
	if tag.getLogger().EnableLevel(DebugLevel) {
		Record(ctx, DebugLevel, tag, 2, Msgf(format, args...))
	}
}

// Info logs a message at InfoLevel using structured fields.
func Info(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, InfoLevel, tag, 2, fields...)
}

// Infof logs a message at InfoLevel using a formatted message.
func Infof(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, InfoLevel, tag, 2, Msgf(format, args...))
}

// Warn logs a message at WarnLevel using structured fields.
func Warn(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, WarnLevel, tag, 2, fields...)
}

// Warnf logs a message at WarnLevel using a formatted message.
func Warnf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, WarnLevel, tag, 2, Msgf(format, args...))
}

// Error logs a message at ErrorLevel using structured fields.
func Error(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, ErrorLevel, tag, 2, fields...)
}

// Errorf logs a message at ErrorLevel using a formatted message.
func Errorf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, ErrorLevel, tag, 2, Msgf(format, args...))
}

// Panic logs a message at PanicLevel using structured fields.
func Panic(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, PanicLevel, tag, 2, fields...)
}

// Panicf logs a message at PanicLevel using a formatted message.
func Panicf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, PanicLevel, tag, 2, Msgf(format, args...))
}

// Fatal logs a message at FatalLevel using structured fields.
func Fatal(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, FatalLevel, tag, 2, fields...)
}

// Fatalf logs a message at FatalLevel using a formatted message.
func Fatalf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, FatalLevel, tag, 2, Msgf(format, args...))
}

// Record is the core function that handles publishing log events.
// It checks the logger level, captures caller information, gathers context fields,
// and sends the log event to the logger.
func Record(ctx context.Context, level Level, tag *Tag, skip int, fields ...Field) {
	l := tag.getLogger()

	// Check if the logger is enabled for the given level
	if !l.EnableLevel(level) {
		return
	}

	file, line := Caller(skip, true)

	// Determine the log timestamp
	now := time.Now()
	if TimeNow != nil {
		now = TimeNow(ctx)
	}

	// Extract a string from the context
	var ctxString string
	if StringFromContext != nil {
		ctxString = StringFromContext(ctx)
	}

	// Extract contextual fields from the context
	var ctxFields []Field
	if FieldsFromContext != nil {
		ctxFields = FieldsFromContext(ctx)
	}

	e := GetEvent()
	e.Level = level
	e.Time = now
	e.File = file
	e.Line = line
	e.Tag = tag.name
	e.Fields = fields
	e.CtxString = ctxString
	e.CtxFields = ctxFields

	l.Publish(e)
}
