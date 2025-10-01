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
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-spring/spring-base/util"
)

// fastCaller controls whether to use a faster but less accurate
// implementation of caller lookup (file/line).
var fastCaller atomic.Bool

func init() {
	fastCaller.Store(true)
	RegisterProperty("fastCaller", func(s string) error {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return util.WrapError(err, "invalid fastCaller: %q", s)
		}
		fastCaller.Store(b)
		return nil
	})
}

// defaultLogger is the default logger associated with tags created
// before custom loggers are fully configured.
var defaultLogger Logger = &SyncLogger{
	LoggerBase: LoggerBase{
		Level: InfoLevel,
		AppenderRefs: []*AppenderRef{
			{
				appender: &ConsoleAppender{
					AppenderBase: AppenderBase{
						Layout: &TextLayout{
							BaseLayout: BaseLayout{
								FileLineLength: 48,
							},
						},
					},
				},
			},
		},
	},
}

var (
	// TagAppDef is the default tag for application-related logs.
	TagAppDef = RegisterAppTag("def", "")

	// TagBizDef is the default tag for business-related logs.
	TagBizDef = RegisterBizTag("def", "")
)

var (
	// TimeNow is an overrideable function to provide custom timestamps.
	// For example, this can be replaced during testing to return a fixed time.
	TimeNow func(ctx context.Context) time.Time

	// StringFromContext is an optional hook to extract a string (e.g., trace ID)
	// from the context. This string will be attached to the log event.
	StringFromContext func(ctx context.Context) string

	// FieldsFromContext is an optional hook to extract structured fields
	// (e.g., trace ID, span ID, or request metadata) from the context.
	FieldsFromContext func(ctx context.Context) []Field
)

// Trace logs at TraceLevel using a tag and a lazy field generator function.
// The function fn() is only executed if TraceLevel logging is enabled.
func Trace(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.getLogger().EnableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, 2, fn()...)
	}
}

// Tracef logs at TraceLevel using a tag and a formatted message.
// Message formatting is only performed if TraceLevel logging is enabled.
func Tracef(ctx context.Context, tag *Tag, format string, args ...any) {
	if tag.getLogger().EnableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, 2, Msgf(format, args...))
	}
}

// Debug logs at DebugLevel using a tag and a lazy field generator function.
// The function fn() is only executed if DebugLevel logging is enabled.
func Debug(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.getLogger().EnableLevel(DebugLevel) {
		Record(ctx, DebugLevel, tag, 2, fn()...)
	}
}

// Debugf logs at DebugLevel using a tag and a formatted message.
// Message formatting is only performed if DebugLevel logging is enabled.
func Debugf(ctx context.Context, tag *Tag, format string, args ...any) {
	if tag.getLogger().EnableLevel(DebugLevel) {
		Record(ctx, DebugLevel, tag, 2, Msgf(format, args...))
	}
}

// Info logs at InfoLevel with structured fields.
func Info(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, InfoLevel, tag, 2, fields...)
}

// Infof logs at InfoLevel using a formatted message.
func Infof(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, InfoLevel, tag, 2, Msgf(format, args...))
}

// Warn logs at WarnLevel with structured fields.
func Warn(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, WarnLevel, tag, 2, fields...)
}

// Warnf logs at WarnLevel using a formatted message.
func Warnf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, WarnLevel, tag, 2, Msgf(format, args...))
}

// Error logs at ErrorLevel with structured fields.
func Error(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, ErrorLevel, tag, 2, fields...)
}

// Errorf logs at ErrorLevel using a formatted message.
func Errorf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, ErrorLevel, tag, 2, Msgf(format, args...))
}

// Panic logs at PanicLevel with structured fields.
func Panic(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, PanicLevel, tag, 2, fields...)
}

// Panicf logs at PanicLevel using a formatted message.
func Panicf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, PanicLevel, tag, 2, Msgf(format, args...))
}

// Fatal logs at FatalLevel with structured fields.
func Fatal(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, FatalLevel, tag, 2, fields...)
}

// Fatalf logs at FatalLevel using a formatted message.
func Fatalf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, FatalLevel, tag, 2, Msgf(format, args...))
}

// Record is the core logging function.
//
// Responsibilities:
//  1. Check whether the logger is enabled for the given level.
//  2. Capture caller information (file, line). When fastCaller is enabled,
//     a faster but less precise lookup is used.
//  3. Determine the log timestamp, either via TimeNow (if set) or time.Now().
//  4. Extract context-based metadata via StringFromContext and FieldsFromContext.
//  5. Populate a pooled Event object with all gathered data.
//  6. Publish the Event to the logger.
func Record(ctx context.Context, level Level, tag *Tag, skip int, fields ...Field) {
	l := tag.getLogger()

	// Step 1: check if logging is enabled for this level.
	if !l.EnableLevel(level) {
		return
	}

	// Step 2: resolve caller information.
	file, line := Caller(skip, fastCaller.Load())

	// Step 3: determine log timestamp.
	now := time.Now()
	if TimeNow != nil {
		now = TimeNow(ctx)
	}

	// Step 4: extract metadata from context.
	var ctxString string
	if StringFromContext != nil {
		ctxString = StringFromContext(ctx)
	}

	var ctxFields []Field
	if FieldsFromContext != nil {
		ctxFields = FieldsFromContext(ctx)
	}

	// Step 5: populate event.
	e := GetEvent()
	e.Level = level
	e.Time = now
	e.File = file
	e.Line = line
	e.Tag = tag.name
	e.Fields = fields
	e.CtxString = ctxString
	e.CtxFields = ctxFields

	// Step 6: publish the event.
	l.Publish(e)
}
