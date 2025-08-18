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
		"logger.myLogger.tags":                "_com_request_*",
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

func TestReadJson(t *testing.T) {
	_, err := readConfigFromReader(strings.NewReader(`
		=1
	`), ".json")
	require.ThatError(t, err).Matches(`ReadJSON error: invalid character '=' looking for beginning of value`)
}

func TestReadYaml(t *testing.T) {
	_, err := readConfigFromReader(strings.NewReader(`
		=1
	`), ".yaml")
	require.ThatError(t, err).Matches(`ReadYaml error: yaml: line 2: found character that cannot start any token`)
}

func TestReadProperties(t *testing.T) {
	_, err := readConfigFromReader(strings.NewReader(`
		=1
	`), ".properties")
	require.ThatError(t, err).Matches(`ReadProperties error: properties: Line 2: \"1\"`)
}
