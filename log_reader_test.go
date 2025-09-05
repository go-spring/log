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

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/gs-assert/require"
)

func TestReaders(t *testing.T) {
	expected := map[string]string{
		"appender.console.layout.type":        "TextLayout",
		"appender.console.type":               "Console",
		"appender.file.fileName":              "log.txt",
		"appender.file.layout.type":           "JSONLayout",
		"appender.file.type":                  "File",
		"appender.multi.layout.type":          "TextLayout",
		"appender.multi.type":                 "MultiAppender",
		"bufferCap":                           "1KB",
		"bufferSize":                          "1000",
		"logger.myLogger.appenderRef[0].ref":  "file",
		"logger.myLogger.appenderRef[0].type": "AppenderRef",
		"logger.myLogger.appenderRef[1].ref":  "multi",
		"logger.myLogger.appenderRef[1].type": "AppenderRef",
		"logger.myLogger.bufferSize":          "${bufferSize}",
		"logger.myLogger.level":               "trace",
		"logger.myLogger.tags":                "_com_request_in,_com_request_*",
		"logger.myLogger.type":                "AsyncLogger",
		"rootLogger.appenderRef.ref":          "console",
		"rootLogger.appenderRef.type":         "AppenderRef",
		"rootLogger.level":                    "warn",
		"rootLogger.type":                     "Root",
	}
	testFiles := []string{
		"testdata/log.properties",
		"testdata/log.json",
		"testdata/log.xml",
		"testdata/log.yaml",
	}
	for _, fileName := range testFiles {
		s, err := readConfigFromFile(fileName)
		require.ThatError(t, err).Nil()
		s.Set("rootLogger.appenderRef.type", "AppenderRef", 0)
		s.Set("logger.myLogger.appenderRef[0].type", "AppenderRef", 0)
		s.Set("logger.myLogger.appenderRef[1].type", "AppenderRef", 0)
		assert.ThatMap(t, s.Data()).Equal(expected, fileName)
	}
}

func TestReadJSON(t *testing.T) {
	_, err := readConfigFromReader(strings.NewReader(`
		=1
	`), ".json")
	require.ThatError(t, err).Matches(`ReadJSON error: invalid character '=' looking for beginning of value`)
}

func TestReadYAML(t *testing.T) {
	_, err := readConfigFromReader(strings.NewReader(`
		=1
	`), ".yaml")
	require.ThatError(t, err).Matches(`ReadYAML error: yaml: line 2: found character that cannot start any token`)
}

func TestReadProperties(t *testing.T) {

	t.Run("syntax error", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			=1
		`), ".properties")
		require.ThatError(t, err).Matches(`ReadProperties error: properties: Line 2: \"1\"`)
	})

	t.Run("key error", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			a=b
			a.b=c
		`), ".properties")
		require.ThatError(t, err).Matches(`property conflict at path a`)
	})
}

func TestReadXML(t *testing.T) {

	t.Run("syntax error - 1", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
			</Configuration
		`), ".xml")
		require.ThatError(t, err).Matches(`ReadXML error: XML syntax error on line 4: .*`)
	})

	t.Run("syntax error - 2", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Properties>
				</Properties
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Matches(`ReadXML error: XML syntax error on line 5: .*`)
	})

	t.Run("syntax error - 3", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Properties>
					<Property></Property
				</Properties>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Matches(`ReadXML error: XML syntax error on line 5: .*`)
	})

	t.Run("missing Appenders", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Properties>
					<Property name="bufferCap">1KB</Property>
				</Properties>
				<Loggers>
					<Root level="warn">
						<AppenderRef ref="console"/>
					</Root>
				</Loggers>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Matches(`missing Appenders`)
	})

	t.Run("missing Loggers", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Properties>
					<Property name="bufferCap">1KB</Property>
				</Properties>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
				</Appenders>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Matches(`missing Loggers`)
	})

	t.Run("unsupported xml tag - 1", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<UnsupportedTag>
				</UnsupportedTag>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Matches(`ReadXML error: unsupported xml tag UnsupportedTag`)
	})

	t.Run("unsupported xml tag - 2", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Properties>
					<UnsupportedProperty name="test">value</UnsupportedProperty>
				</Properties>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
				</Appenders>
				<Loggers>
					<Root level="warn">
						<AppenderRef ref="console"/>
					</Root>
				</Loggers>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Matches(`unsupported xml tag UnsupportedProperty`)
	})

	t.Run("multiple root", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
				</Appenders>
				<Loggers>
					<Root level="warn">
						<AppenderRef ref="console"/>
					</Root>
					<AsyncRoot level="info">
						<AppenderRef ref="console"/>
					</AsyncRoot>
				</Loggers>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Matches(`found multiple root loggers`)
	})

	t.Run("missing root", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
				</Appenders>
				<Loggers>
					<Logger name="test">
						<AppenderRef ref="console"/>
					</Logger>
				</Loggers>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Matches(`missing Root or AsyncRoot`)
	})

	t.Run("success - Root", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
				</Appenders>
				<Loggers>
					<Root>
						<AppenderRef ref="console"/>
					</Root>
					<Logger name="test">
						<AppenderRef ref="console"/>
					</Logger>
				</Loggers>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Nil()
	})

	t.Run("success - AsyncRoot", func(t *testing.T) {
		_, err := readConfigFromReader(strings.NewReader(`
			<Configuration>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
				</Appenders>
				<Loggers>
					<AsyncRoot>
						<AppenderRef ref="console"/>
					</AsyncRoot>
					<Logger name="test">
						<AppenderRef ref="console"/>
						<AppenderRef ref="console"/>
						<AppenderRef ref="console"/>
					</Logger>
				</Loggers>
			</Configuration>
		`), ".xml")
		require.ThatError(t, err).Nil()
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
		{"rootLogger.level", "rootLogger.level"},
		{"rootLogger.appenderRef.ref", "rootLogger.appenderRef.ref"},
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
		assert.ThatString(t, result).Equal(test.expected)
	}
}
