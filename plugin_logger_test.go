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
	"testing"
	"time"

	"github.com/go-spring/gs-assert/assert"
)

type CountAppender struct {
	Appender
	count int
}

func (c *CountAppender) Append(e *Event) {
	c.count++
	c.Appender.Append(e)
}

func TestLoggerConfig(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.ThatError(t, err).Nil()

		l := &SyncLogger{BaseLogger{
			Level: InfoLevel,
			Tags:  "_com_*",
			AppenderRefs: []*AppenderRef{
				{appender: a},
			},
		}}

		err = l.Start()
		assert.ThatError(t, err).Nil()

		assert.That(t, l.EnableLevel(TraceLevel)).False()
		assert.That(t, l.EnableLevel(DebugLevel)).False()
		assert.That(t, l.EnableLevel(InfoLevel)).True()
		assert.That(t, l.EnableLevel(WarnLevel)).True()
		assert.That(t, l.EnableLevel(ErrorLevel)).True()
		assert.That(t, l.EnableLevel(PanicLevel)).True()
		assert.That(t, l.EnableLevel(FatalLevel)).True()

		for range 5 {
			l.Publish(&Event{})
		}

		assert.That(t, a.count).Equal(5)

		l.Stop()
		a.Stop()
	})
}

func TestAsyncLoggerConfig(t *testing.T) {

	t.Run("enable level", func(t *testing.T) {
		l := &AsyncLogger{
			BaseLogger: BaseLogger{
				Level: InfoLevel,
			},
		}

		assert.That(t, l.EnableLevel(TraceLevel)).False()
		assert.That(t, l.EnableLevel(DebugLevel)).False()
		assert.That(t, l.EnableLevel(InfoLevel)).True()
		assert.That(t, l.EnableLevel(WarnLevel)).True()
		assert.That(t, l.EnableLevel(ErrorLevel)).True()
		assert.That(t, l.EnableLevel(PanicLevel)).True()
		assert.That(t, l.EnableLevel(FatalLevel)).True()
	})

	t.Run("error BufferSize", func(t *testing.T) {
		l := &AsyncLogger{
			BaseLogger: BaseLogger{
				Name: "file",
			},
			BufferSize: 10,
		}

		err := l.Start()
		assert.ThatError(t, err).Matches("bufferSize is too small")
	})

	t.Run("drop events", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.ThatError(t, err).Nil()

		dropCount := 0
		OnDropEvent = func(logger string, v interface{}) {
			dropCount++
		}
		defer func() {
			OnDropEvent = nil
		}()

		l := &AsyncLogger{
			BaseLogger: BaseLogger{
				Level: InfoLevel,
				Tags:  "_com_*",
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize: 100,
		}

		err = l.Start()
		assert.ThatError(t, err).Nil()

		for range 5000 {
			l.Publish(GetEvent())
		}

		time.Sleep(200 * time.Millisecond)

		l.Stop()
		a.Stop()

		assert.That(t, dropCount > 0).True()
	})

	t.Run("success", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.ThatError(t, err).Nil()

		l := &AsyncLogger{
			BaseLogger: BaseLogger{
				Level: InfoLevel,
				Tags:  "_com_*",
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize: 100,
		}

		err = l.Start()
		assert.ThatError(t, err).Nil()

		for range 5 {
			l.Publish(GetEvent())
		}

		time.Sleep(100 * time.Millisecond)
		assert.That(t, a.count).Equal(5)

		l.Stop()
		a.Stop()
	})
}
