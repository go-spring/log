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

The `log.RefreshFile` function allows loading the logger's configuration from an external file (e.g., yaml or JSON).

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
	"runtime"
	"strconv"
	"time"

	"github.com/lvan100/errutil"
)

var (
	// enableCaller controls whether to enable caller lookup (file/line).
	enableCaller = true

	// fastCaller controls whether to use a faster but less accurate
	// implementation of caller lookup (file/line).
	fastCaller = false
)

func init() {
	// Property: enableCaller
	RegisterProperty("enableCaller", func(s string) error {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return errutil.Stack(err, "invalid enableCaller: %q", s)
		}
		enableCaller = b
		return nil
	})

	// Property: fastCaller
	RegisterProperty("fastCaller", func(s string) error {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return errutil.Stack(err, "invalid fastCaller: %q", s)
		}
		fastCaller = b
		return nil
	})
}

// defaultLogger serves as the fallback logger used when no custom logger
// has been configured for a specific tag.
var defaultLogger Logger = &ConsoleLogger{
	LoggerBase: LoggerBase{
		Level: LevelRange{
			MinLevel: NoneLevel,
			MaxLevel: MaxLevel,
		},
		Layout: &TextLayout{
			BaseLayout: BaseLayout{
				FileLineLength: 48,
			},
		},
	},
	ConsoleAppender: ConsoleAppender{},
}

var (
	// TagAppDef is the default tag for application-related logs.
	TagAppDef = RegisterAppTag("def", "")

	// TagBizDef is the default tag for business-related logs.
	TagBizDef = RegisterBizTag("def", "")
)

// RegisterAppTag registers or retrieves a Tag intended for application-layer logs.
//   - subType: component or module name
//   - action: lifecycle phase or behavior (optional)
func RegisterAppTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("app", subType, action))
}

// RegisterBizTag registers or retrieves a Tag intended for business-logic logs.
//   - subType: business domain or feature name
//   - action: operation being logged (optional)
func RegisterBizTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("biz", subType, action))
}

// RegisterRPCTag registers or retrieves a Tag intended for RPC logs,
// covering external/internal dependency interactions.
//   - subType: protocol or target system (e.g., http, grpc, redis)
//   - action: RPC phase (e.g., send, retry, fail)
func RegisterRPCTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("rpc", subType, action))
}

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

// getLogger returns the logger associated with the given tag.
// If no logger is bound, the default logger is returned.
func getLogger(tag *Tag) Logger {
	if tag.logger != nil {
		return tag.logger
	}
	return defaultLogger
}

// Trace logs a message at TraceLevel using a lazy field generator.
// The generator function is only invoked if the level is enabled.
func Trace(ctx context.Context, tag *Tag, fn func() []Field) {
	if l := getLogger(tag); l.GetLevel().Enable(TraceLevel) {
		record(ctx, TraceLevel, tag.tag, l, 2, fn()...)
	}
}

// Tracef logs a formatted message at TraceLevel.
func Tracef(ctx context.Context, tag *Tag, format string, args ...any) {
	if l := getLogger(tag); l.GetLevel().Enable(TraceLevel) {
		record(ctx, TraceLevel, tag.tag, l, 2, Msgf(format, args...))
	}
}

// Debug logs a message at DebugLevel using a lazy field generator.
// The generator function is only invoked if the level is enabled.
func Debug(ctx context.Context, tag *Tag, fn func() []Field) {
	if l := getLogger(tag); l.GetLevel().Enable(DebugLevel) {
		record(ctx, DebugLevel, tag.tag, l, 2, fn()...)
	}
}

// Debugf logs a formatted message at DebugLevel.
func Debugf(ctx context.Context, tag *Tag, format string, args ...any) {
	if l := getLogger(tag); l.GetLevel().Enable(DebugLevel) {
		record(ctx, DebugLevel, tag.tag, l, 2, Msgf(format, args...))
	}
}

// Info logs structured fields at InfoLevel.
func Info(ctx context.Context, tag *Tag, fields ...Field) {
	if l := getLogger(tag); l.GetLevel().Enable(InfoLevel) {
		record(ctx, InfoLevel, tag.tag, l, 2, fields...)
	}
}

// Infof logs a formatted message at InfoLevel.
func Infof(ctx context.Context, tag *Tag, format string, args ...any) {
	if l := getLogger(tag); l.GetLevel().Enable(InfoLevel) {
		record(ctx, InfoLevel, tag.tag, l, 2, Msgf(format, args...))
	}
}

// Warn logs structured fields at WarnLevel.
func Warn(ctx context.Context, tag *Tag, fields ...Field) {
	if l := getLogger(tag); l.GetLevel().Enable(WarnLevel) {
		record(ctx, WarnLevel, tag.tag, l, 2, fields...)
	}
}

// Warnf logs a formatted message at WarnLevel.
func Warnf(ctx context.Context, tag *Tag, format string, args ...any) {
	if l := getLogger(tag); l.GetLevel().Enable(WarnLevel) {
		record(ctx, WarnLevel, tag.tag, l, 2, Msgf(format, args...))
	}
}

// Error logs structured fields at ErrorLevel.
func Error(ctx context.Context, tag *Tag, fields ...Field) {
	if l := getLogger(tag); l.GetLevel().Enable(ErrorLevel) {
		record(ctx, ErrorLevel, tag.tag, l, 2, fields...)
	}
}

// Errorf logs a formatted message at ErrorLevel.
func Errorf(ctx context.Context, tag *Tag, format string, args ...any) {
	if l := getLogger(tag); l.GetLevel().Enable(ErrorLevel) {
		record(ctx, ErrorLevel, tag.tag, l, 2, Msgf(format, args...))
	}
}

// Panic logs structured fields at PanicLevel.
func Panic(ctx context.Context, tag *Tag, fields ...Field) {
	if l := getLogger(tag); l.GetLevel().Enable(PanicLevel) {
		record(ctx, PanicLevel, tag.tag, l, 2, fields...)
	}
}

// Panicf logs a formatted message at PanicLevel.
func Panicf(ctx context.Context, tag *Tag, format string, args ...any) {
	if l := getLogger(tag); l.GetLevel().Enable(PanicLevel) {
		record(ctx, PanicLevel, tag.tag, l, 2, Msgf(format, args...))
	}
}

// Fatal logs structured fields at FatalLevel.
func Fatal(ctx context.Context, tag *Tag, fields ...Field) {
	if l := getLogger(tag); l.GetLevel().Enable(FatalLevel) {
		record(ctx, FatalLevel, tag.tag, l, 2, fields...)
	}
}

// Fatalf logs a formatted message at FatalLevel.
func Fatalf(ctx context.Context, tag *Tag, format string, args ...any) {
	if l := getLogger(tag); l.GetLevel().Enable(FatalLevel) {
		record(ctx, FatalLevel, tag.tag, l, 2, Msgf(format, args...))
	}
}

// Record logs a message at the given level for the given tag.
func Record(ctx context.Context, level Level, tag *Tag, skip int, fields ...Field) {
	if l := getLogger(tag); l.GetLevel().Enable(level) {
		record(ctx, level, tag.tag, l, skip, fields...)
	}
}

// record performs the actual logging logic after level checking.
func record(ctx context.Context, level Level, tag string, logger Logger, skip int, fields ...Field) {

	// Step 1: check if logging is enabled for this level.
	if !logger.GetLevel().Enable(level) {
		return
	}

	// Step 2: capture caller information.
	var (
		file string
		line int
	)
	if enableCaller {
		if fastCaller {
			file, line = FastCaller(skip)
		} else {
			_, file, line, _ = runtime.Caller(skip + 1)
		}
	}

	// Step 3: get log timestamp.
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
	e.Tag = tag
	e.Fields = fields
	e.CtxString = ctxString
	e.CtxFields = ctxFields

	// Step 6: publish the event.
	logger.Append(e)
}
