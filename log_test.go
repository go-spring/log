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
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-spring/log"
	"github.com/lvan100/golib/flatten"
	"github.com/lvan100/golib/testing/assert"
)

var (
	keyTraceID int
	keySpanID  int
)

var TagDefault = log.RegisterTag("_def")
var TagRequestIn = log.RegisterTag("_com_request_in")
var TagRequestOut = log.RegisterTag(log.BuildTag("com", "request", "out"))

var rootLogger = log.GetLogger(log.RootLoggerName)

// Loggers with same name 'myLogger'
var myLogger = log.GetLogger("myLogger")
var myLoggerV2 = log.GetLogger("myLogger")

///////////////////////////////////////////////////////////////////////////////

var _ log.Appender = (*SampleAppender)(nil)

type SampleAppender struct {
	log.AppenderBase
	Layout log.Layout `PluginElement:"Layout,default=TextLayout"`
}

func (a *SampleAppender) Start() error   { return nil }
func (a *SampleAppender) Stop()          {}
func (a *SampleAppender) Write(b []byte) {}
func (a *SampleAppender) Append(e *log.Event) {
	data := a.Layout.ToBytes(e)
	_, _ = os.Stdout.Write(data)
	_, _ = os.Stderr.Write(data)
}

func init() {
	log.RegisterPlugin[SampleAppender]("Sample", log.PluginTypeAppender)
}

///////////////////////////////////////////////////////////////////////////////

func readConfig() map[string]string {
	s := `
	{
	  "bufferCap": "1KB",
	  "bufferSize": 1000,
	  "appender": {
	    "file": {
	      "type": "File",
	      "fileName": "log.txt",
	      "layout!": "JSONLayout{}"
	    },
	    "console!": "Console{layout=TextLayout{}}",
	    "sample!": "Sample{layout.type=TextLayout}"
	  },
	  "logger": {
	    "root": {
	      "type": "Logger",
	      "level": "warn",
	      "appenderRef": {
	        "ref": "console"
	      }
	    },
	    "myLogger": {
	      "type": "AsyncLogger",
	      "level": "trace",
	      "tags": "_com_request_in,_com_request_*",
	      "bufferSize": "${bufferSize}",
	      "appenderRef": [
	        {
	          "ref": "file"
	        },
	        {
	          "ref": "sample"
	        }
	      ]
	    }
	  }
	}`

	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		panic(err)
	}
	return flatten.Flatten(m)
}

func TestLog(t *testing.T) {
	ctx := t.Context()
	_ = os.Remove("log.txt")

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

	err := log.Refresh(readConfig())
	assert.Error(t, err).Nil()

	// should panic after init.
	assert.Panic(t, func() {
		log.GetLogger("myLogger")
	}, "log refresh already done")

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

	_, _ = rootLogger.Write([]byte("this message is written directly\n"))
	_, _ = rootLogger.Write([]byte("this message is written directly\n"))

	expectLog := `
[INFO][2025-06-01T00:00:00.000][log_test.go:110] _def||trace_id=||span_id=||msg=hello world
[INFO][2025-06-01T00:00:00.000][log_test.go:111] _com_request_in||trace_id=||span_id=||msg=hello world
[WARN][2025-06-01T00:00:00.000][log_test.go:160] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[ERROR][2025-06-01T00:00:00.000][log_test.go:161] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[PANIC][2025-06-01T00:00:00.000][log_test.go:162] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[ERROR][2025-06-01T00:00:00.000][log_test.go:165] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||key1=value1||key2=value2
this message is written directly
this message is written directly
`

	assert.String(t, logBuf.String()).Equal(strings.TrimLeft(expectLog, "\n"))

	assert.Panic(t, func() {
		log.RegisterRPCTag("def", "")
	}, "log refresh already done")

	_, _ = myLogger.Write([]byte("this message is written directly\n"))
	_, _ = myLoggerV2.Write([]byte("this message is written directly\n"))

	expectLog = `
{"level":"trace","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:125","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"debug","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:132","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"trace","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:139","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"debug","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:140","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"info","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:143","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"warn","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:144","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"error","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:145","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"panic","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:146","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"fatal","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:147","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"info","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:150","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"warn","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:151","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"error","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:152","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"panic","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:153","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"fatal","time":"2025-06-01T00:00:00.000","fileLine":"log_test.go:154","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
this message is written directly
this message is written directly
`

	// Since an asynchronous logger is used, it is necessary to call stop first
	// before performing the assertion to ensure all log entries are flushed and
	// the logger is properly stopped.
	log.Destroy()

	b, err := os.ReadFile("log.txt")
	assert.Error(t, err).Nil()
	assert.String(t, string(b)).Equal(strings.TrimLeft(expectLog, "\n"))
}
