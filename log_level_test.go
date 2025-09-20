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

	"github.com/go-spring/spring-base/testing/assert"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		str     string
		want    Level
		wantErr error
	}{
		{
			str:  "none",
			want: NoneLevel,
		},
		{
			str:  "trace",
			want: TraceLevel,
		},
		{
			str:  "debug",
			want: DebugLevel,
		},
		{
			str:  "info",
			want: InfoLevel,
		},
		{
			str:  "warn",
			want: WarnLevel,
		},
		{
			str:  "error",
			want: ErrorLevel,
		},
		{
			str:  "panic",
			want: PanicLevel,
		},
		{
			str:  "fatal",
			want: FatalLevel,
		},
		{
			str:     "unknown",
			want:    NoneLevel,
			wantErr: FormatError(nil, "invalid log level: %q", "unknown"),
		},
	}
	for _, tt := range tests {
		got, err := ParseLevel(tt.str)
		assert.That(t, got).Equal(tt.want)
		assert.That(t, err).Equal(tt.wantErr)
		assert.ThatNumber(t, got.Code()).Equal(tt.want.Code())
	}

	// Test that levels are properly ordered by code
	assert.ThatNumber(t, NoneLevel.Code()).LessThan(TraceLevel.Code())
	assert.ThatNumber(t, TraceLevel.Code()).LessThan(DebugLevel.Code())
	assert.ThatNumber(t, DebugLevel.Code()).LessThan(InfoLevel.Code())
	assert.ThatNumber(t, InfoLevel.Code()).LessThan(WarnLevel.Code())
	assert.ThatNumber(t, WarnLevel.Code()).LessThan(ErrorLevel.Code())
	assert.ThatNumber(t, ErrorLevel.Code()).LessThan(PanicLevel.Code())
	assert.ThatNumber(t, PanicLevel.Code()).LessThan(FatalLevel.Code())
}

func TestRegisterLevel(t *testing.T) {

	customLevel := RegisterLevel(800, "custom")
	assert.ThatNumber(t, customLevel.Code()).Equal(int32(800))
	assert.ThatString(t, customLevel.String()).Equal("CUSTOM")

	parsed, err := ParseLevel("custom")
	assert.ThatError(t, err).Nil()
	assert.That(t, parsed).Equal(customLevel)

	parsed, err = ParseLevel("Custom")
	assert.ThatError(t, err).Nil()
	assert.That(t, parsed).Equal(customLevel)

	parsed, err = ParseLevel("CUSTOM")
	assert.ThatError(t, err).Nil()
	assert.That(t, parsed).Equal(customLevel)
}
