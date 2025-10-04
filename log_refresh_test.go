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
)

func TestRefreshFile(t *testing.T) {
	t.Cleanup(func() {
		for _, tag := range tagRegistry {
			tag.setLogger(defaultLogger)
		}
	})

	t.Run("file not exist", func(t *testing.T) {
		defer func() { Destroy() }()
		err := RefreshFile("testdata/file-not-exist.yaml")
		assert.Error(t, err).Matches("open testdata/file-not-exist.yaml")
	})

	t.Run("already refresh", func(t *testing.T) {
		defer func() { Destroy() }()
		err := RefreshFile("testdata/log.Yaml")
		assert.Error(t, err).Nil()
		err = RefreshFile("testdata/log.Yaml")
		assert.Error(t, err).Matches("log refresh already done")
	})
}

func TestRefreshConfig(t *testing.T) {
	t.Cleanup(func() {
		for _, tag := range tagRegistry {
			tag.setLogger(defaultLogger)
		}
	})

	t.Run("unsupported file", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		err := RefreshReader(nil, ".toml")
		assert.Error(t, err).Matches("RefreshReader error: unsupported file type .toml")
	})

	t.Run("appenders section not found", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("appenders section not found")
	})

	t.Run("read appenders error", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			appender=ERROR_PROPERTY
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).String("read appenders section error << property conflict at path appender")
	})

	t.Run("read loggers error", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			appender.console.type=Console
			logger=ERROR_PROPERTY
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).String("read loggers section error << property conflict at path logger")
	})

	t.Run("plugin not found - appender", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			appender.console.type=NonExistentAppender
			logger.test.type=AsyncLogger
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("plugin NonExistentAppender not found")
	})

	t.Run("plugin not found - rootLogger", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=NonExistentRoot
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.test.type=AsyncLogger
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("plugin NonExistentRoot not found")
	})

	t.Run("rootLogger no type", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.level=debug
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.test.type=AsyncLogger
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("attribute 'type' not found")
	})

	t.Run("init AppenderRefs error - rootLogger", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			rootLogger.level=debug
			rootLogger.appenderRef.ref=file
			appender.console.type=Console
			appender.console.layout.type=TextLayout
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("appender file not found")
	})

	t.Run("plugin not found - loggers", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			rootLogger.level=debug
			rootLogger.appenderRef.ref=console
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.myLogger.type=NonExistentLogger
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("plugin NonExistentLogger not found")
	})

	t.Run("loggers no type", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			rootLogger.level=debug
			rootLogger.appenderRef.ref=console
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.myLogger.level=info
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("attribute 'type' not found")
	})

	t.Run("plugin not found - loggers", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			rootLogger.level=debug
			rootLogger.appenderRef.ref=console
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.myLogger.type=Logger
			logger.myLogger.level=info
			logger.myLogger.appenderRef.ref=file
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("appender file not found")
	})

	t.Run("loggers no tags", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			rootLogger.level=debug
			rootLogger.appenderRef.ref=console
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.myLogger.type=Logger
			logger.myLogger.level=info
			logger.myLogger.appenderRef.ref=console
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("logger must have attribute 'tags'")
	})

	t.Run("loggers tags error", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			rootLogger.level=debug
			rootLogger.appenderRef.ref=console
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.myLogger.type=Logger
			logger.myLogger.level=info
			logger.myLogger.tags=**
			logger.myLogger.appenderRef.ref=console
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).Matches("`\\*\\*` regexp compile error")
	})

	t.Run("logger start error", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			rootLogger.type=Root
			rootLogger.level=debug
			rootLogger.appenderRef.ref=console
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.myLogger.type=AsyncLogger
			logger.myLogger.level=info
			logger.myLogger.tags=.*
			logger.myLogger.bufferSize=10
			logger.myLogger.appenderRef.ref=console
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).String("logger myLogger start error << bufferSize is too small")
	})

	t.Run("logger start error", func(t *testing.T) {
		defer func() { global.init.Store(false) }()
		content := `
			bufferCap=1GB
			rootLogger.type=Root
			rootLogger.level=debug
			rootLogger.appenderRef.ref=console
			appender.console.type=Console
			appender.console.layout.type=TextLayout
			logger.myLogger.type=AsyncLogger
			logger.myLogger.level=info
			logger.myLogger.tags=.*
			logger.myLogger.appenderRef.ref=console
		`
		err := RefreshReader(strings.NewReader(content), ".properties")
		assert.Error(t, err).String(`inject property bufferCap error << invalid bufferCap: "1GB" << unhandled size name: "GB"`)
	})

}
