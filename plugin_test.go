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

	"github.com/go-spring/stdlib/flatten"
	"github.com/go-spring/stdlib/testing/assert"
)

func TestRegisterPlugin(t *testing.T) {
	assert.Panic(t, func() {
		RegisterPlugin[int]("DummyLayout")
	}, "T must be struct")
	assert.Panic(t, func() {
		RegisterPlugin[FileAppender]("FileAppender")
	}, "duplicate plugin name \"FileAppender\" in .*/plugin_appender.go:.* and .*/plugin_test.go:.*")
}

func TestInjectAttribute(t *testing.T) {

	t.Run("no attribute - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Name string `PluginAttribute:""`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, "test", nil)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Name error >> found no attribute")
	})

	t.Run("no attribute - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Value string `PluginAttribute:"value"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Value error >> found no attribute")
	})

	t.Run("property not found error", func(t *testing.T) {
		type ErrorPlugin struct {
			Value string `PluginAttribute:"value"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.value", "${nonexistent_prop}")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field Value error >> property \${nonexistent_prop} not found`)
	})

	t.Run("converter error", func(t *testing.T) {
		type ErrorPlugin struct {
			Level LevelRange `PluginAttribute:"level,default=NOT-EXIST-LEVEL"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Level error >> invalid log level: \"NOT-EXIST-LEVEL\"")
	})

	t.Run("uint64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M uint64 `PluginAttribute:"m,default=111"`
			N uint64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field N error >> strconv.ParseUint: parsing \"abc\": invalid syntax`)
	})

	t.Run("int64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M int64 `PluginAttribute:"m,default=111"`
			N int64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field N error >> strconv.ParseInt: parsing \"abc\": invalid syntax`)
	})

	t.Run("float64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M float64 `PluginAttribute:"m,default=111"`
			N float64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field N error >> strconv.ParseFloat: parsing \"abc\": invalid syntax`)
	})

	t.Run("boolean error", func(t *testing.T) {
		type ErrorPlugin struct {
			M bool `PluginAttribute:"m,default=true"`
			N bool `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field N error >> strconv.ParseBool: parsing \"abc\": invalid syntax`)
	})

	t.Run("type error", func(t *testing.T) {
		type ErrorPlugin struct {
			M chan error `PluginAttribute:"m,default=true"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field M error >> unsupported inject type chan error`)
	})

	t.Run("success with name attribute", func(t *testing.T) {
		type SuccessPlugin struct {
			Name string `PluginAttribute:"name"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.String(t, p.Name).Equal("test")
	})

	t.Run("success with default value", func(t *testing.T) {
		type SuccessPlugin struct {
			Value string `PluginAttribute:"value,default=hello"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.String(t, p.Value).Equal("hello")
	})

	t.Run("success with storage value", func(t *testing.T) {
		type SuccessPlugin struct {
			Value string `PluginAttribute:"value"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.value", "world")
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
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("prop.value", "property_value")
		s.Set("test.value", "${prop.value}")
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.String(t, p.Value).Equal("property_value")
	})

	// Tests for array/slice injection
	t.Run("slice from comma separated value", func(t *testing.T) {
		type SlicePlugin struct {
			Values []string `PluginAttribute:"values"`
		}
		typ := reflect.TypeFor[SlicePlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.values", "apple,banana,cherry")
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SlicePlugin)
		assert.Slice(t, p.Values).Equal([]string{"apple", "banana", "cherry"})
	})

	t.Run("slice from indexed keys", func(t *testing.T) {
		type SlicePlugin struct {
			Numbers []int `PluginAttribute:"numbers"`
		}
		typ := reflect.TypeFor[SlicePlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.numbers[0]", "10")
		s.Set("test.numbers[1]", "20")
		s.Set("test.numbers[2]", "30")
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SlicePlugin)
		assert.Slice(t, p.Numbers).Equal([]int{10, 20, 30})
	})

	t.Run("slice with default and custom separator", func(t *testing.T) {
		type SlicePlugin struct {
			Values []string `PluginAttribute:"values,delimiter=;,default=a;b;c"`
		}
		typ := reflect.TypeFor[SlicePlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SlicePlugin)
		assert.Slice(t, p.Values).Equal([]string{"a", "b", "c"})
	})

	t.Run("slice with property reference", func(t *testing.T) {
		type SlicePlugin struct {
			Values []string `PluginAttribute:"values"`
		}
		typ := reflect.TypeFor[SlicePlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("prop.array", "property,value,array")
		s.Set("test.values", "${prop.array}")
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SlicePlugin)
		assert.Slice(t, p.Values).Equal([]string{"property", "value", "array"})
	})

	t.Run("slice conversion error", func(t *testing.T) {
		type ErrorPlugin struct {
			Numbers []int `PluginAttribute:"numbers"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.numbers", "1,abc,3")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field Numbers error >> inject struct field Numbers\[1] error >> strconv.ParseInt: parsing "abc": invalid syntax`)
	})

	t.Run("slice unsupported element type", func(t *testing.T) {
		type ErrorPlugin struct {
			Channels []chan error `PluginAttribute:"channels"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.channels", "test")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field Channels error >> inject struct field Channels\[0] error >> unsupported inject type chan error`)
	})

	t.Run("fixed array from indexed keys", func(t *testing.T) {
		type ArrayPlugin struct {
			Numbers [3]int `PluginAttribute:"numbers"`
		}
		typ := reflect.TypeFor[ArrayPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.numbers[0]", "7")
		s.Set("test.numbers[1]", "8")
		s.Set("test.numbers[2]", "9")
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*ArrayPlugin)
		assert.That(t, p.Numbers).Equal([3]int{7, 8, 9})
	})

	t.Run("fixed array overflow", func(t *testing.T) {
		type ArrayPlugin struct {
			Numbers [2]int `PluginAttribute:"numbers"`
		}
		typ := reflect.TypeFor[ArrayPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.numbers", "1,2,3")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ArrayPlugin error >> inject struct field Numbers error >> too many values for array \[2\]int`)
	})
}

func TestInjectElement(t *testing.T) {

	t.Run("no element", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:""`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layout error >> found no plugin element")
	})

	t.Run("unsupported inject type", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout map[string]Layout `PluginElement:"layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.layout.type", "TextLayout")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layout error >> unsupported inject type map\\[string]log.Layout")
	})

	t.Run("no element - slice - default", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts []Layout `PluginElement:"layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layouts error >> found no plugin element")
	})

	t.Run("no element - slice - len - 0", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts []Layout `PluginElement:"layout,default=;;"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layouts error >> found no plugin element")
	})

	t.Run("plugin not found - slice - interface - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout []Layout `PluginElement:"layout,default=NotExistElement"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layout error >> plugin NotExistElement not found")
	})

	t.Run("plugin not found - slice - interface - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout []Layout `PluginElement:"layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.layout.type", "NotExistElement")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layout error >> plugin NotExistElement not found")
	})

	t.Run("plugin not found - slice - struct - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			AppenderRefs []*AppenderRef `PluginElement:"appenderRef"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.appenderRef[0].ref", "file")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
	})

	t.Run("plugin not found - slice - struct - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			AppenderRefs []*AppenderRef `PluginElement:"appenderRef"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.appenderRef.ref", "file")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
	})

	t.Run("no element - single - default", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts Layout `PluginElement:"layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layouts error >> found no plugin element")
	})

	t.Run("no element - single - no - type", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts Layout `PluginElement:"layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.layout.dummy", "")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layouts error >> found no plugin element")
	})

	t.Run("plugin not found - single - interface - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"layout,default=NotExistElement"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layout error >> plugin NotExistElement not found")
	})

	t.Run("plugin not found - single - interface - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.layout.type", "NotExistElement")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches("create plugin log.ErrorPlugin error >> inject struct field Layout error >> plugin NotExistElement not found")
	})

	t.Run("NewPlugin error - slice - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Appenders []Appender `PluginElement:"appender,default=FileAppender"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field Appenders error >> create plugin log.FileAppender error >> inject struct field FileName error >> found no attribute`)
	})

	t.Run("NewPlugin error - slice - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Appenders []Appender `PluginElement:"appender"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		s.Set("test.appender.type", "FileAppender")
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field Appenders error >> create plugin log.FileAppender error >> inject struct field FileName error >> found no attribute`)
	})

	t.Run("NewPlugin error - single", func(t *testing.T) {
		type ErrorPlugin struct {
			Appender Appender `PluginElement:"appender,default=FileAppender"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		_, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Matches(`create plugin log.ErrorPlugin error >> inject struct field Appender error >> create plugin log.FileAppender error >> inject struct field FileName error >> found no attribute`)
	})

	t.Run("success - slice - 1", func(t *testing.T) {
		type SuccessPlugin struct {
			Layouts []Layout `PluginElement:"layout,default=TextLayout"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.That(t, p.Layouts).NotNil()
	})

	t.Run("success - slice -2", func(t *testing.T) {
		type SuccessPlugin struct {
			Layouts []Layout `PluginElement:"layout"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		ps.Set("test.layout.type", "TextLayout")
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.That(t, p.Layouts).NotNil()
	})

	t.Run("success - single", func(t *testing.T) {
		type SuccessPlugin struct {
			Layout Layout `PluginElement:"layout"`
		}
		typ := reflect.TypeFor[SuccessPlugin]()
		ps := flatten.NewProperties(nil)
		s := flatten.NewPropertiesStorage(ps)
		ps.Set("test.layout.type", "TextLayout")
		v, err := NewPlugin(typ, "test", s)
		assert.Error(t, err).Nil()
		p := v.Interface().(*SuccessPlugin)
		assert.That(t, p.Layout).NotNil()
	})

}
