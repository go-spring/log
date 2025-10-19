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
	"testing"

	"github.com/go-spring/spring-base/testing/assert"
	"github.com/go-spring/spring-base/testing/require"
)

func TestReaders(t *testing.T) {
	expected := map[string]string{
		"appender.console.layout.type":        "TextLayout",
		"appender.console.type":               "Console",
		"appender.file.fileName":              "log.txt",
		"appender.file.layout.type":           "JSONLayout",
		"appender.file.type":                  "File",
		"appender.sample.layout.type":         "TextLayout",
		"appender.sample.type":                "Sample",
		"bufferCap":                           "1KB",
		"bufferSize":                          "1000",
		"logger.myLogger.appenderRef[0].ref":  "file",
		"logger.myLogger.appenderRef[0].type": "AppenderRef",
		"logger.myLogger.appenderRef[1].ref":  "sample",
		"logger.myLogger.appenderRef[1].type": "AppenderRef",
		"logger.myLogger.bufferSize":          "${bufferSize}",
		"logger.myLogger.level":               "trace",
		"logger.myLogger.tags":                "_com_request_in,_com_request_*",
		"logger.myLogger.type":                "AsyncLogger",
		"logger.root.appenderRef.ref":         "console",
		"logger.root.appenderRef.type":        "AppenderRef",
		"logger.root.level":                   "warn",
		"logger.root.type":                    "Logger",
	}
	testFiles := []string{
		"testdata/log.properties",
		"testdata/log.json",
		"testdata/log.Yaml",
	}
	for _, fileName := range testFiles {
		s, err := readConfigFromFile(fileName)
		require.Error(t, err).Nil()
		_ = s.Set("logger.root.appenderRef.type", "AppenderRef", 0)
		_ = s.Set("logger.myLogger.appenderRef[0].type", "AppenderRef", 0)
		_ = s.Set("logger.myLogger.appenderRef[1].type", "AppenderRef", 0)
		assert.Map(t, s.Data()).Equal(expected, fileName)
	}
}

func TestReadJSON(t *testing.T) {
	_, err := readConfigFromReader(strings.NewReader(`
		=1
	`), ".json")
	require.Error(t, err).Matches(`ReadJSON error: invalid character '=' looking for beginning of value`)
}

func TestReadYAML(t *testing.T) {
	_, err := readConfigFromReader(strings.NewReader(`=1`), ".yaml")
	require.Error(t, err).NotNil()
	require.Error(t, err).String(`ReadYAML error: yaml: unmarshal errors:
  line 1: cannot unmarshal !!str ` + "`=1`" + ` into map[string]interface {}`)

	s, err := readConfigFromReader(strings.NewReader(`123: abc`), ".yaml")
	require.Error(t, err).Nil()
	require.Map(t, s.Data()).Equal(map[string]string{"123": "abc"})
}

func TestReadProperties(t *testing.T) {

	t.Run("syntax error", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			=1
		`), ".properties")
		require.Error(t, err).Matches(`ReadProperties error: properties: Line 2: \"1\"`)
	})

	t.Run("key error", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			a=b
			a.b=c
		`), ".properties")
		require.Error(t, err).Matches(`property conflict at path a`)
	})
}

func TestToCamelKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"Key", "key"},
		{"KeyTest", "keyTest"},
		{"logger.root.level", "logger.root.level"},
		{"logger.root.appenderRef.ref", "logger.root.appenderRef.ref"},
		{"appender_ref", "appenderRef"},
		{"appender_ref_type", "appenderRefType"},
		{"logger-myLogger", "loggerMyLogger"},
		{"logger-myLogger-level", "loggerMyLoggerLevel"},
		{"appender.file_name", "appender.fileName"},
		{"appender-file_name", "appenderFileName"},
		{"logger_MyLogger.appenderRef.ref", "loggerMyLogger.appenderRef.ref"},
		{"logger.MyLogger.appenderRef[0].ref", "logger.myLogger.appenderRef[0].ref"},
		{"appender.console.layout.type", "appender.console.layout.type"},
		{"bufferCap", "bufferCap"},
		{"buffer_size", "bufferSize"},
	}
	for _, test := range tests {
		result := toCamelKey(test.input)
		assert.String(t, result).Equal(test.expected)
	}
}
