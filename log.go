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
language. It offers flexible and structured logging capabilities, including context field extraction, multi-level
logging configuration, and multiple output options, making it ideal for a wide range of server-side applications.

Context field extraction can be customized:

	log.StringFromContext = func(ctx context.Context) string {
		return ""
	}

	log.FieldsFromContext = func(ctx context.Context) []log.Field {
		return []log.Field{
			log.String("trace_id", "0a882193682db71edd48044db54cae88"),
			log.String("span_id", "50ef0724418c0a66"),
		}
	}

Load configuration from a file:

	err := log.RefreshFile("log.xml")
	if err != nil {
		panic(err)
	}

Log messages with formatted output:

	log.Tracef(ctx, TagRequestOut, "hello %s", "world")
	log.Debugf(ctx, TagRequestOut, "hello %s", "world")
	log.Infof(ctx, TagRequestIn, "hello %s", "world")
	log.Warnf(ctx, TagRequestIn, "hello %s", "world")
	log.Errorf(ctx, TagRequestIn, "hello %s", "world")
	log.Panicf(ctx, TagRequestIn, "hello %s", "world")
	log.Fatalf(ctx, TagRequestIn, "hello %s", "world")

Structured logging using field functions:

	log.Trace(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	log.Debug(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	log.Info(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Warn(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Error(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Panic(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Fatal(ctx, TagRequestIn, log.Msgf("hello %s", "world"))

Log structured fields using a map:

	log.Error(ctx, TagDefault, log.Fields(map[string]any{
		"key1": "value1",
		"key2": "value2",
	})...)
*/
package log

import (
	"context"
	"time"
)

var (
	TagAppDef = GetAppTag("def", "")
	TagBizDef = GetBizTag("def", "")
)

// GetAppTag returns a Tag used for application-layer logs (e.g., framework events, lifecycle).
// subType represents the component or module, action represents the lifecycle phase or behavior.
func GetAppTag(subType, action string) *Tag {
	return GetTag(buildTag("app", subType, action))
}

// GetBizTag returns a Tag used for business-logic logs (e.g., use cases, domain events).
// subType is the business domain or feature name, action is the operation being logged.
func GetBizTag(subType, action string) *Tag {
	return GetTag(buildTag("biz", subType, action))
}

// GetRPCTag returns a Tag used for RPC or external/internal dependency logs.
// subType is the protocol or target system, action is the RPC phase (e.g., sent, retry, fail).
func GetRPCTag(subType, action string) *Tag {
	return GetTag(buildTag("rpc", subType, action))
}

// buildTag constructs a structured tag string based on main type, sub type, and action.
// The format is: _<mainType>_<subType> or _<mainType>_<subType>_<action>.
func buildTag(mainType, subType, action string) string {
	if action == "" {
		return "_" + mainType + "_" + subType
	}
	return "_" + mainType + "_" + subType + "_" + action
}

// TimeNow is a function that can be overridden to provide custom timestamp behavior (e.g., for testing).
var TimeNow func(ctx context.Context) time.Time

// StringFromContext can be set to extract a string from the context.
var StringFromContext func(ctx context.Context) string

// FieldsFromContext can be set to extract structured fields from the context (e.g., trace IDs, user IDs).
var FieldsFromContext func(ctx context.Context) []Field

// Fields converts a map of string keys to any values to a slice of Field.
func Fields(fields map[string]any) []Field {
	var ret []Field
	for _, k := range OrderedMapKeys(fields) {
		ret = append(ret, Any(k, fields[k]))
	}
	return ret
}

// Trace logs a message at TraceLevel using a tag and a lazy field-generating function.
func Trace(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.GetLogger().EnableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, 2, fn()...)
	}
}

// Tracef logs a message at TraceLevel using a tag and a formatted message.
func Tracef(ctx context.Context, tag *Tag, format string, args ...any) {
	if tag.GetLogger().EnableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, 2, Msgf(format, args...))
	}
}

// Debug logs a message at DebugLevel using a tag and a lazy field-generating function.
func Debug(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.GetLogger().EnableLevel(DebugLevel) {
		Record(ctx, DebugLevel, tag, 2, fn()...)
	}
}

// Debugf logs a message at DebugLevel using a tag and a formatted message.
func Debugf(ctx context.Context, tag *Tag, format string, args ...any) {
	if tag.GetLogger().EnableLevel(DebugLevel) {
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
	var l Logger

	// Check if the logger is enabled for the given level
	if l = tag.GetLogger(); !l.EnableLevel(level) {
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
	e.Tag = tag.GetName()
	e.Fields = fields
	e.CtxString = ctxString
	e.CtxFields = ctxFields

	l.Publish(e)
}
