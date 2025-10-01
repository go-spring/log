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

	"github.com/go-spring/spring-base/testing/assert"
)

func TestParseBufferFullPolicy(t *testing.T) {
	_, err := ParseBufferFullPolicy("block")
	assert.Error(t, err).Matches("invalid BufferFullPolicy block")

	p, err := ParseBufferFullPolicy("Block")
	assert.Error(t, err).Nil()
	assert.That(t, p).Equal(BufferFullPolicyBlock)

	p, err = ParseBufferFullPolicy("Discard")
	assert.Error(t, err).Nil()
	assert.That(t, p).Equal(BufferFullPolicyDiscard)

	p, err = ParseBufferFullPolicy("DiscardOldest")
	assert.Error(t, err).Nil()
	assert.That(t, p).Equal(BufferFullPolicyDiscardOldest)
}

type CountAppender struct {
	Appender
	count int
}

func (c *CountAppender) Append(e *Event) {
	c.count++
	c.Appender.Append(e)
}

func TestLoggerConfig(t *testing.T) {

	t.Run("write", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Error(t, err).Nil()

		l := &SyncLogger{LoggerBase{
			AppenderRefs: []*AppenderRef{
				{appender: a},
			},
		}}

		n, err := l.Write([]byte("test"))
		assert.Error(t, err).Nil()
		assert.That(t, n).Equal(4)
		assert.That(t, a.count).Equal(0)

		l.Stop()
		a.Stop()
	})

	t.Run("success", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Error(t, err).Nil()

		l := &SyncLogger{LoggerBase{
			Level: InfoLevel,
			Tags:  "_com_*",
			AppenderRefs: []*AppenderRef{
				{appender: a},
			},
		}}

		err = l.Start()
		assert.Error(t, err).Nil()

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
			LoggerBase: LoggerBase{
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
			LoggerBase: LoggerBase{
				Name: "file",
			},
			BufferSize: 10,
		}

		err := l.Start()
		assert.Error(t, err).Matches("bufferSize is too small")
	})

	t.Run("buffer full - discard", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Error(t, err).Nil()

		l := &AsyncLogger{
			LoggerBase: LoggerBase{
				Level: InfoLevel,
				Tags:  "_com_*",
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize:       100,
			BufferFullPolicy: BufferFullPolicyDiscard,
		}

		err = l.Start()
		assert.Error(t, err).Nil()

		go func() {
			for range 100 {
				_, _ = l.Write([]byte("hello"))
			}
		}()

		for range 5000 {
			l.Publish(GetEvent())
		}

		time.Sleep(200 * time.Millisecond)

		l.Stop()
		a.Stop()

		assert.That(t, l.GetDiscardCounter() > 0).True()
	})

	t.Run("buffer full - discard oldest", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Error(t, err).Nil()

		l := &AsyncLogger{
			LoggerBase: LoggerBase{
				Level: InfoLevel,
				Tags:  "_com_*",
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize:       100,
			BufferFullPolicy: BufferFullPolicyDiscardOldest,
		}

		err = l.Start()
		assert.Error(t, err).Nil()

		go func() {
			for range 100 {
				_, _ = l.Write([]byte("hello"))
			}
		}()

		for range 5000 {
			l.Publish(GetEvent())
		}

		time.Sleep(200 * time.Millisecond)

		l.Stop()
		a.Stop()

		assert.That(t, l.GetDiscardCounter() > 0).True()
	})

	t.Run("buffer full - block", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Error(t, err).Nil()

		l := &AsyncLogger{
			LoggerBase: LoggerBase{
				Level: InfoLevel,
				Tags:  "_com_*",
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize:       100,
			BufferFullPolicy: BufferFullPolicyBlock,
		}

		err = l.Start()
		assert.Error(t, err).Nil()

		go func() {
			for range 100 {
				_, _ = l.Write([]byte("hello"))
			}
		}()

		for range 5000 {
			l.Publish(GetEvent())
		}

		time.Sleep(200 * time.Millisecond)

		l.Stop()
		a.Stop()

		assert.That(t, l.GetDiscardCounter() == 0).True()
	})

	t.Run("success", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Error(t, err).Nil()

		l := &AsyncLogger{
			LoggerBase: LoggerBase{
				Level: InfoLevel,
				Tags:  "_com_*",
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize: 100,
		}

		err = l.Start()
		assert.Error(t, err).Nil()

		for range 5 {
			l.Publish(GetEvent())
		}

		time.Sleep(100 * time.Millisecond)
		assert.That(t, a.count).Equal(5)

		l.Stop()
		a.Stop()
	})

	t.Run("write with discard policy", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Error(t, err).Nil()

		l := &AsyncLogger{
			LoggerBase: LoggerBase{
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize:       100,
			BufferFullPolicy: BufferFullPolicyDiscard,
		}

		err = l.Start()
		assert.Error(t, err).Nil()

		// Rapidly write large amount of data to fill the buffer
		for range 500 {
			_, _ = l.Write([]byte("test data"))
		}

		time.Sleep(100 * time.Millisecond)
		l.Stop()
		a.Stop()

		// Some data should be discarded
		assert.That(t, l.GetDiscardCounter() > 0).True()
	})
}
