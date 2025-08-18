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
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-spring/barky"
)

var converters = map[reflect.Type]any{}

// Converter function type that converts string to a specific type T.
type Converter[T any] func(string) (T, error)

// RegisterConverter Registers a converter for a specific type T.
func RegisterConverter[T any](fn Converter[T]) {
	t := reflect.TypeFor[T]()
	converters[t] = fn
}

var propertyMap = make(map[string]func(string) error)

// RegisterProperty Registers a property with a given key.
func RegisterProperty(key string, val func(string) error) {
	propertyMap[key] = val
}

// Lifecycle Optional lifecycle interface for plugin instances.
type Lifecycle interface {
	Start() error
	Stop()
}

// PluginType Defines types of plugins supported by the logging system.
type PluginType string

const (
	PluginTypeProperty    PluginType = "Property"
	PluginTypeAppender    PluginType = "Appender"
	PluginTypeLayout      PluginType = "Layout"
	PluginTypeAppenderRef PluginType = "AppenderRef"
	PluginTypeRoot        PluginType = "Root"
	PluginTypeAsyncRoot   PluginType = "AsyncRoot"
	PluginTypeLogger      PluginType = "Logger"
	PluginTypeAsyncLogger PluginType = "AsyncLogger"
)

var plugins = map[string]*Plugin{}

func init() {
	RegisterPlugin[struct{}]("Property", PluginTypeProperty)
}

// Plugin metadata structure
type Plugin struct {
	Name  string       // Name of plugin
	Type  PluginType   // Type of plugin
	Class reflect.Type // Underlying struct type
	File  string       // Source file of registration
	Line  int          // Line number of registration
}

// RegisterPlugin Registers a plugin with a given name and type.
func RegisterPlugin[T any](name string, typ PluginType) {
	_, file, line, _ := runtime.Caller(1)
	if p, ok := plugins[name]; ok {
		panic(fmt.Errorf("duplicate plugin %s in %s:%d and %s:%d", typ, p.File, p.Line, file, line))
	}
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic("T must be struct")
	}
	plugins[name] = &Plugin{
		Name:  name,
		Type:  typ,
		Class: t,
		File:  file,
		Line:  line,
	}
}

// NewPlugin Creates and initializes a plugin instance.
func NewPlugin(t reflect.Type, prefix string, s *barky.Storage) (reflect.Value, error) {
	v := reflect.New(t)
	if err := inject(v.Elem(), t, prefix, s); err != nil {
		return reflect.Value{}, WrapError(err, "create plugin %s error", t.String())
	}
	return v, nil
}

// inject Recursively injects values into struct fields based on tags.
func inject(v reflect.Value, t reflect.Type, prefix string, s *barky.Storage) error {
	for i := range v.NumField() {
		ft := t.Field(i)
		fv := v.Field(i)
		if tag, ok := ft.Tag.Lookup("PluginAttribute"); ok {
			if err := injectAttribute(tag, fv, ft, prefix, s); err != nil {
				return err
			}
			continue
		}
		if tag, ok := ft.Tag.Lookup("PluginElement"); ok {
			if err := injectElement(tag, fv, ft, prefix, s); err != nil {
				return err
			}
			continue
		}
		// Recursively process anonymous embedded structs
		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := inject(fv, fv.Type(), prefix, s); err != nil {
				return err
			}
		}
	}
	return nil
}

type PluginTag string

// Get Gets the value of a key or the first unnamed value.
func (tag PluginTag) Get(key string) string {
	v, _ := tag.Lookup(key)
	return v
}

// Lookup Looks up a key-value pair in the tag.
func (tag PluginTag) Lookup(key string) (value string, ok bool) {
	kvs := strings.Split(string(tag), ",")
	if key == "" {
		return kvs[0], true
	}
	for i := 1; i < len(kvs); i++ {
		ss := strings.Split(kvs[i], "=")
		if len(ss) != 2 {
			return "", false
		} else if ss[0] == key {
			return ss[1], true
		}
	}
	return "", false
}

// injectAttribute Injects a value into a struct field from plugin attribute.
func injectAttribute(tag string, fv reflect.Value, ft reflect.StructField, prefix string, s *barky.Storage) error {

	attrTag := PluginTag(tag)
	attrName := attrTag.Get("")
	if attrName == "" {
		return fmt.Errorf("found no attribute for struct field %s", ft.Name)
	}

	if attrName == "name" {
		name := prefix[strings.LastIndex(prefix, ".")+1:]
		fv.SetString(name)
		return nil
	}

	var val string
	key := prefix + "." + ToCamelKey(attrName)
	if v, ok := s.RawData()[key]; ok {
		val = v.Value
	} else {
		val, ok = attrTag.Lookup("default")
		if !ok {
			return fmt.Errorf("found no attribute for struct field %s", ft.Name)
		}
	}

	// Use a property if available
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		v, ok := s.RawData()[ToCamelKey(val[2:len(val)-1])]
		if !ok {
			return fmt.Errorf("property %s not found", val)
		}
		val = v.Value
	}

	// Use a custom converter if available
	if fn := converters[ft.Type]; fn != nil {
		fnValue := reflect.ValueOf(fn)
		out := fnValue.Call([]reflect.Value{reflect.ValueOf(val)})
		if !out[1].IsNil() {
			err := out[1].Interface().(error)
			return WrapError(err, "inject struct field %s error", ft.Name)
		}
		fv.Set(out[0])
		return nil
	}

	switch fv.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 0, 0)
		if err == nil {
			fv.SetUint(u)
			return nil
		}
		return WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 0, 0)
		if err == nil {
			fv.SetInt(i)
			return nil
		}
		return WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			fv.SetFloat(f)
			return nil
		}
		return WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err == nil {
			fv.SetBool(b)
			return nil
		}
		return WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.String:
		fv.SetString(val)
		return nil
	default:
		return fmt.Errorf("unsupported inject type %s for struct field %s", ft.Type.String(), ft.Name)
	}
}

// injectElement Injects plugin elements (child nodes) into struct fields.
func injectElement(tag string, fv reflect.Value, ft reflect.StructField, prefix string, s *barky.Storage) error {

	elemTag := PluginTag(tag)
	elemType := elemTag.Get("")
	if elemType == "" {
		return fmt.Errorf("found no element for struct field %s", ft.Name)
	}
	elemKey := prefix + "." + ToCamelKey(elemType)

	switch fv.Kind() {
	case reflect.Slice:

		p, ok := plugins[elemType]
		if !ok {
			return fmt.Errorf("plugin %s not found for struct field %s", elemType, ft.Name)
		}

		slice := reflect.MakeSlice(ft.Type, 0, 1)
		if s.Has(elemKey + "[0]") { // 多个配置
			for i := 0; ; i++ {
				subKey := elemKey + "[" + strconv.Itoa(i) + "]"
				if !s.Has(subKey) {
					break
				}
				v, err := NewPlugin(p.Class, subKey, s)
				if err != nil {
					return err
				}
				slice = reflect.Append(slice, v)
			}
		} else if s.Has(elemKey) { // 单个配置
			v, err := NewPlugin(p.Class, elemKey, s)
			if err != nil {
				return err
			}
			slice = reflect.Append(slice, v)
		}

		if slice.Len() == 0 {
			return fmt.Errorf("found no plugin elements for struct field %s", ft.Name)
		}

		fv.Set(slice)
		return nil

	case reflect.Interface:

		var strType string
		if s.Has(elemKey) {
			typeKey := elemKey + ".type"
			if v, ok := s.RawData()[typeKey]; ok {
				strType = v.Value
			} else {
				return fmt.Errorf("found no plugin elements for struct field %s", ft.Name)
			}
		} else {
			elemLabel, ok := elemTag.Lookup("default")
			if !ok {
				return fmt.Errorf("found no plugin elements for struct field %s", ft.Name)
			}
			strType = elemLabel
		}

		p, ok := plugins[strType]
		if !ok {
			return fmt.Errorf("plugin %s not found for struct field %s", strType, ft.Name)
		}

		v, err := NewPlugin(p.Class, elemKey, s)
		if err != nil {
			return err
		}
		fv.Set(v)
		return nil

	default:
		return fmt.Errorf("unsupported inject type %s for struct field %s", ft.Type.String(), ft.Name)
	}
}
