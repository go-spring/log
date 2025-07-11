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

	"github.com/go-spring/log"
	"github.com/lvan100/go-assert"
)

var (
	keyTraceID int
	keySpanID  int
)

var TagDefault = log.GetTag("_def")
var TagRequestIn = log.GetTag("_com_request_in")
var TagRequestOut = log.GetTag("_com_request_out")

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
	assert.Nil(t, err)

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
	log.Error(ctx, TagDefault, log.Fields(map[string]any{
		"key1": "value1",
		"key2": "value2",
	})...)

	expectLog := `
[INFO][2025-06-01T00:00:00.000][log_test.go:82] _def||trace_id=||span_id=||msg=hello world
[INFO][2025-06-01T00:00:00.000][log_test.go:83] _com_request_in||trace_id=||span_id=||msg=hello world
[WARN][2025-06-01T00:00:00.000][log_test.go:127] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[ERROR][2025-06-01T00:00:00.000][log_test.go:128] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[PANIC][2025-06-01T00:00:00.000][log_test.go:129] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[ERROR][2025-06-01T00:00:00.000][log_test.go:132] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||key1=value1||key2=value2
`

	assert.ThatString(t, logBuf.String()).Equal(strings.TrimLeft(expectLog, "\n"))

	expectLog = `
{"level":"trace","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:92","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"debug","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:99","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"trace","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:106","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"debug","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:107","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"info","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:110","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"warn","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:111","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"error","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:112","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"panic","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:113","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"fatal","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:114","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"info","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:117","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"warn","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:118","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"error","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:119","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"panic","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:120","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"fatal","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:121","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
`

	b, err := os.ReadFile("log.txt")
	assert.Nil(t, err)
	assert.ThatString(t, string(b)).Equal(strings.TrimLeft(expectLog, "\n"))
}
