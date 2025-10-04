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
	"reflect"
	"testing"

	"github.com/go-spring/spring-base/barky"
	"github.com/go-spring/spring-base/testing/assert"
)

func TestRegisterPlugin(t *testing.T) {
	assert.Panic(t, func() {
		RegisterPlugin[int]("DummyLayout")
	}, "T must be struct")
	assert.Panic(t, func() {
		RegisterPlugin[FileAppender]("File")
	}, "duplicate plugin name \"File\" in .*/plugin_appender.go:30 and .*/plugin_test.go:32")
}

func TestInjectAttribute(t *testing.T) {

	t.Run("no attribute - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Name string `PluginAttribute:""`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, "test", nil)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Name error << found no attribute")
	})

	t.Run("no attribute - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Value string `PluginAttribute:"value"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Value error << found no attribute")
	})

	t.Run("property not found error", func(t *testing.T) {
		type ErrorPlugin struct {
			Value string `PluginAttribute:"value"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.value", "${nonexistent_prop}", 0)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field Value error << property \${nonexistent_prop} not found`)
	})

	t.Run("converter error", func(t *testing.T) {
		type ErrorPlugin struct {
			Level Level `PluginAttribute:"level,default=NOT-EXIST-LEVEL"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Level error << invalid log level: \"NOT-EXIST-LEVEL\"")
	})

	t.Run("uint64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M uint64 `PluginAttribute:"m,default=111"`
			N uint64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field N error << strconv.ParseUint: parsing \"abc\": invalid syntax`)
	})

	t.Run("int64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M int64 `PluginAttribute:"m,default=111"`
			N int64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field N error << strconv.ParseInt: parsing \"abc\": invalid syntax`)
	})

	t.Run("float64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M float64 `PluginAttribute:"m,default=111"`
			N float64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field N error << strconv.ParseFloat: parsing \"abc\": invalid syntax`)
	})

	t.Run("boolean error", func(t *testing.T) {
		type ErrorPlugin struct {
			M bool `PluginAttribute:"m,default=true"`
			N bool `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field N error << strconv.ParseBool: parsing \"abc\": invalid syntax`)
	})

	t.Run("type error", func(t *testing.T) {
		type ErrorPlugin struct {
			M chan error `PluginAttribute:"m,default=true"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field M error << unsupported inject type chan error`)
	})

	t.Run("success with name attribute", func(t *testing.T) {
		type SuccessPlugin struct {
			Name string `PluginAttribute:"name"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		v, err := NewPlugin(typ, "test", barky.NewStorage())
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.String(t, p.Name).Equal("test")
	})

	t.Run("success with default value", func(t *testing.T) {
		type SuccessPlugin struct {
			Value string `PluginAttribute:"value,default=hello"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		v, err := NewPlugin(typ, "test", barky.NewStorage())
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.String(t, p.Value).Equal("hello")
	})

	t.Run("success with storage value", func(t *testing.T) {
		type SuccessPlugin struct {
			Value string `PluginAttribute:"value"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.value", "world", 0)
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.String(t, p.Value).Equal("world")
	})

	t.Run("success with property reference", func(t *testing.T) {
		type SuccessPlugin struct {
			Value string `PluginAttribute:"value"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		s := barky.NewStorage()
		_ = s.Set("prop.value", "property_value", 0)
		_ = s.Set("test.value", "${prop.value}", 0)
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.String(t, p.Value).Equal("property_value")
	})
}

func TestInjectElement(t *testing.T) {

	t.Run("no element", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:""`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, "test", barky.NewStorage())
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layout error << found no plugin element")
	})

	t.Run("unsupported inject type", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout map[string]Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.layout.type", "TextLayout", 0)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layout error << unsupported inject type map\\[string]log.Layout")
	})

	t.Run("no element - slice - default", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts []Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, "test", barky.NewStorage())
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layouts error << found no plugin element")
	})

	t.Run("no element - slice - len - 0", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts []Layout `PluginElement:"Layout,default=;;"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, "test", barky.NewStorage())
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layouts error << found no plugin element")
	})

	t.Run("plugin not found - slice - interface - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout []Layout `PluginElement:"Layout,default=NotExistElement"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layout error << plugin NotExistElement not found")
	})

	t.Run("plugin not found - slice - interface - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout []Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.layout.type", "NotExistElement", 0)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layout error << plugin NotExistElement not found")
	})

	t.Run("plugin not found - slice - struct - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			AppenderRefs []*AppenderRef `PluginElement:"NotExistElement"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.notExistElement[0].ref", "file", 0)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field AppenderRefs error << plugin NotExistElement not found")
	})

	t.Run("plugin not found - slice - struct - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			AppenderRefs []*AppenderRef `PluginElement:"NotExistElement"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.notExistElement.ref", "file", 0)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field AppenderRefs error << plugin NotExistElement not found")
	})

	t.Run("no element - single - default", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, "test", barky.NewStorage())
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layouts error << found no plugin element")
	})

	t.Run("no element - single - no - type", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.layout.dummy", "", 0)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layouts error << found no plugin element")
	})

	t.Run("plugin not found - single - interface - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout,default=NotExistElement"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layout error << plugin NotExistElement not found")
	})

	t.Run("plugin not found - single - interface - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.layout.type", "NotExistElement", 0)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Layout error << plugin NotExistElement not found")
	})

	t.Run("NewPlugin error - slice - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Appenders []Appender `PluginElement:"Appender,default=File"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, "test", barky.NewStorage())
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field Appenders error << create plugin log.FileAppender error << inject struct field Layout error << found no plugin element`)
	})

	t.Run("NewPlugin error - slice - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Appenders []Appender `PluginElement:"Appender"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.appender.type", "File", 0)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field Appenders error << create plugin log.FileAppender error << inject struct field Layout error << found no plugin element`)
	})

	t.Run("NewPlugin error - single", func(t *testing.T) {
		type ErrorPlugin struct {
			Appender Appender `PluginElement:"Appender,default=File"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, "test", barky.NewStorage())
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field Appender error << create plugin log.FileAppender error << inject struct field Layout error << found no plugin element`)
	})

	t.Run("success - slice - 1", func(t *testing.T) {
		type SuccessPlugin struct {
			Layouts []Layout `PluginElement:"Layout,default=TextLayout"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		s := barky.NewStorage()
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.That(t, p.Layouts).NotNil()
	})

	t.Run("success - slice -2", func(t *testing.T) {
		type SuccessPlugin struct {
			Layouts []Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.layout.type", "TextLayout", 0)
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.That(t, p.Layouts).NotNil()
	})

	t.Run("success - single", func(t *testing.T) {
		type SuccessPlugin struct {
			Layout Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		s := barky.NewStorage()
		_ = s.Set("test.layout.type", "TextLayout", 0)
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.That(t, p.Layout).NotNil()
	})

}
