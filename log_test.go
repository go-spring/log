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

package log_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/log"
)

var (
	keyTraceID int
	keySpanID  int
)

var TagDefault = log.RegisterTag("_def")
var TagRequestIn = log.RegisterTag("_com_request_in")
var TagRequestOut = log.RegisterTag("_com_request_out")

///////////////////////////////////////////////////////////////////////////////

type MultiAppender struct {
	log.BaseAppender
}

func (a *MultiAppender) Append(e *log.Event) {
	data := a.Layout.ToBytes(e)
	_, _ = os.Stdout.Write(data) // To share a layout, you need a multi-appender
	_, _ = os.Stderr.Write(data) // You can integrate with other logging systems
}

func init() {
	log.RegisterPlugin[MultiAppender]("MultiAppender", log.PluginTypeAppender)
}

///////////////////////////////////////////////////////////////////////////////

func TestLog(t *testing.T) {
	ctx := t.Context()
	_ = os.Remove("log.txt")

	oldCaller := log.Caller
	defer func() { log.Caller = oldCaller }()

	log.Caller = func(skip int, fast bool) (file string, line int) {
		file, line = oldCaller(skip+1, fast)
		file = filepath.Base(file)
		return
	}

	logBuf := bytes.NewBuffer(nil)
	log.Stdout = logBuf
	defer func() {
		log.Stdout = os.Stdout
	}()

	log.TimeNow = func(ctx context.Context) time.Time {
		return time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	}

	log.StringFromContext = func(ctx context.Context) string {
		return ""
	}

	log.FieldsFromContext = func(ctx context.Context) []log.Field {
		traceID, _ := ctx.Value(&keyTraceID).(string)
		spanID, _ := ctx.Value(&keySpanID).(string)
		return []log.Field{
			log.String("trace_id", traceID),
			log.String("span_id", spanID),
		}
	}

	// not print
	log.Tracef(ctx, TagRequestOut, "hello %s", "world")
	log.Debugf(ctx, TagRequestOut, "hello %s", "world")

	// print
	log.Info(ctx, TagDefault, log.Msgf("hello %s", "world"))
	log.Info(ctx, TagRequestIn, log.Msgf("hello %s", "world"))

	err := log.RefreshFile("testdata/log.xml")
	assert.ThatError(t, err).Nil()

	ctx = context.WithValue(ctx, &keyTraceID, "0a882193682db71edd48044db54cae88")
	ctx = context.WithValue(ctx, &keySpanID, "50ef0724418c0a66")

	// print
	log.Trace(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	// print
	log.Debug(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	// print
	log.Tracef(ctx, TagRequestOut, "hello %s", "world")
	log.Debugf(ctx, TagRequestOut, "hello %s", "world")

	// print
	log.Info(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Warn(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Error(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Panic(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Fatal(ctx, TagRequestIn, log.Msgf("hello %s", "world"))

	// print
	log.Infof(ctx, TagRequestIn, "hello %s", "world")
	log.Warnf(ctx, TagRequestIn, "hello %s", "world")
	log.Errorf(ctx, TagRequestIn, "hello %s", "world")
	log.Panicf(ctx, TagRequestIn, "hello %s", "world")
	log.Fatalf(ctx, TagRequestIn, "hello %s", "world")

	// not print
	log.Info(ctx, TagDefault, log.Msgf("hello %s", "world"))

	// print
	log.Warn(ctx, TagDefault, log.Msgf("hello %s", "world"))
	log.Error(ctx, TagDefault, log.Msgf("hello %s", "world"))
	log.Panic(ctx, TagDefault, log.Msgf("hello %s", "world"))

	// print
	log.Error(ctx, TagDefault, log.FieldsFromMap(map[string]any{
		"key1": "value1",
		"key2": "value2",
	}))

	expectLog := `
[INFO][2025-06-01T00:00:00.000][log_test.go:100] _def||trace_id=||span_id=||msg=hello world
[INFO][2025-06-01T00:00:00.000][log_test.go:101] _com_request_in||trace_id=||span_id=||msg=hello world
[WARN][2025-06-01T00:00:00.000][log_test.go:145] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[ERROR][2025-06-01T00:00:00.000][log_test.go:146] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[PANIC][2025-06-01T00:00:00.000][log_test.go:147] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[ERROR][2025-06-01T00:00:00.000][log_test.go:150] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||key1=value1||key2=value2
`

	assert.ThatString(t, logBuf.String()).Equal(strings.TrimLeft(expectLog, "\n"))

	expectLog = `
{"level":"trace","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:110","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"debug","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:117","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"trace","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:124","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"debug","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:125","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"info","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:128","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"warn","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:129","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"error","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:130","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"panic","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:131","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"fatal","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:132","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"info","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:135","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"warn","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:136","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"error","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:137","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"panic","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:138","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"fatal","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:139","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
`

	b, err := os.ReadFile("log.txt")
	assert.ThatError(t, err).Nil()
	assert.ThatString(t, string(b)).Equal(strings.TrimLeft(expectLog, "\n"))
}
