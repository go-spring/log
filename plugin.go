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
	"runtime"
	"strconv"
	"strings"

	"github.com/go-spring/spring-base/barky"
	"github.com/go-spring/spring-base/util"
)

// typeConverters holds user-defined type conversion functions
// registered via RegisterConverter. It maps a reflect.Type to a
// conversion function that transforms a string into that type.
//
// This allows dynamic parsing of configuration values (usually from
// text-based formats like YAML, or properties files) into native Go types.
var typeConverters = map[reflect.Type]any{}

// Converter defines a function that converts a string to type T.
type Converter[T any] func(string) (T, error)

// RegisterConverter registers a custom converter for type T.
func RegisterConverter[T any](fn Converter[T]) {
	t := reflect.TypeFor[T]()
	typeConverters[t] = fn
}

// propertyRegistry holds all registered property injection functions.
// Each entry associates a configuration key with a function that applies
// its value at runtime.
var propertyRegistry = make(map[string]func(string) error)

// RegisterProperty registers a property setter with a given key.
func RegisterProperty(key string, val func(string) error) {
	propertyRegistry[key] = val
}

// Lifecycle is an optional interface for plugin lifecycle hooks.
type Lifecycle interface {
	Start() error
	Stop()
}

// PluginType enumerates supported plugin categories.
type PluginType string

const (
	PluginTypeLogger      PluginType = "logger"
	PluginTypeAppender    PluginType = "appender"
	PluginTypeAppenderRef PluginType = "appenderRef"
	PluginTypeLayout      PluginType = "layout"
)

// pluginRegistry maintains a registry of available plugin classes,
// organized by plugin type and plugin name.
var pluginRegistry = map[PluginType]map[string]*Plugin{}

// Plugin represents metadata about a plugin type.
type Plugin struct {
	Name  string       // Name of the plugin
	Type  PluginType   // Type of the plugin
	Class reflect.Type // Underlying struct type
	File  string       // File where plugin was registered
	Line  int          // Line number where plugin was registered
}

// RegisterPlugin registers a plugin struct type with a given name and plugin type.
func RegisterPlugin[T any](name string, typ PluginType) {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic("T must be struct")
	}

	m := pluginRegistry[typ]
	if m == nil {
		m = make(map[string]*Plugin)
		pluginRegistry[typ] = m
	}

	_, file, line, _ := runtime.Caller(1)
	if p, ok := m[name]; ok {
		err := util.FormatError(nil, "duplicate plugin name %q in %s:%d and %s:%d",
			name, p.File, p.Line, file, line)
		panic(err)
	}

	m[name] = &Plugin{
		Name:  name,
		Type:  typ,
		Class: t,
		File:  file,
		Line:  line,
	}
}

// NewPlugin creates a new plugin instance and injects configuration values.
func NewPlugin(t reflect.Type, prefix string, s *barky.Storage) (reflect.Value, error) {
	v := reflect.New(t)
	if err := inject(v.Elem(), t, prefix, s); err != nil {
		return reflect.Value{}, util.WrapError(err, "create plugin %s error", t.String())
	}
	return v, nil
}

// inject recursively sets struct fields based on `PluginAttribute` and `PluginElement` tags.
func inject(v reflect.Value, t reflect.Type, prefix string, s *barky.Storage) error {
	for i := range v.NumField() {
		ft := t.Field(i)
		fv := v.Field(i)

		// Inject from `PluginAttribute` tag
		if tag, ok := ft.Tag.Lookup("PluginAttribute"); ok {
			if err := injectAttribute(tag, fv, ft, prefix, s); err != nil {
				return err
			}
			continue
		}

		// Inject from `PluginElement` tag
		if tag, ok := ft.Tag.Lookup("PluginElement"); ok {
			if err := injectElement(tag, fv, ft, prefix, s); err != nil {
				return err
			}
			continue
		}

		// Recursively inject anonymous embedded structs
		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := inject(fv, fv.Type(), prefix, s); err != nil {
				return err
			}
		}
	}
	return nil
}

// PluginTag is a wrapper for parsing struct field tags.
type PluginTag string

// Get returns the value for a key or the first unnamed value.
func (tag PluginTag) Get(key string) string {
	v, _ := tag.Lookup(key)
	return v
}

// Lookup returns the value of a key in a tag and a boolean indicating existence.
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

// injectAttribute injects a primitive field from configuration storage.
//
// Supported field types include string, bool, integer, unsigned integer,
// and floating-point numbers. Custom converters (registered via
// RegisterConverter) are used if available.
//
// It also supports placeholder syntax `${prop}` for property substitution.
func injectAttribute(tag string, fv reflect.Value, ft reflect.StructField, prefix string, s *barky.Storage) error {

	attrTag := PluginTag(tag)
	attrName := attrTag.Get("")
	if attrName == "" {
		err := util.FormatError(nil, "found no attribute")
		return util.WrapError(err, "inject struct field %s error", ft.Name)
	}

	// Special handling for "name" attribute
	if attrName == "name" {
		name := prefix[strings.LastIndex(prefix, ".")+1:]
		fv.SetString(name)
		return nil
	}

	var val string
	key := prefix + "." + toCamelKey(attrName)
	if v, ok := s.RawData()[key]; ok {
		val = v.Value
	} else {
		val, ok = attrTag.Lookup("default")
		if !ok {
			err := util.FormatError(nil, "found no attribute")
			return util.WrapError(err, "inject struct field %s error", ft.Name)
		}
	}

	// Handle properties in format ${prop}
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		v, ok := s.RawData()[toCamelKey(val[2:len(val)-1])]
		if !ok {
			err := util.FormatError(nil, "property %s not found", val)
			return util.WrapError(err, "inject struct field %s error", ft.Name)
		}
		val = v.Value
	}

	// Use custom converter if exists
	if fn := typeConverters[ft.Type]; fn != nil {
		fnValue := reflect.ValueOf(fn)
		out := fnValue.Call([]reflect.Value{reflect.ValueOf(val)})
		if !out[1].IsNil() {
			err := out[1].Interface().(error)
			return util.WrapError(err, "inject struct field %s error", ft.Name)
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
		return util.WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 0, 0)
		if err == nil {
			fv.SetInt(i)
			return nil
		}
		return util.WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			fv.SetFloat(f)
			return nil
		}
		return util.WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err == nil {
			fv.SetBool(b)
			return nil
		}
		return util.WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.String:
		fv.SetString(val)
		return nil
	default:
		err := util.FormatError(nil, "unsupported inject type %s", ft.Type.String())
		return util.WrapError(err, "inject struct field %s error", ft.Name)
	}
}

// injectElement injects child plugin elements into a struct field.
func injectElement(tag string, fv reflect.Value, ft reflect.StructField, prefix string, s *barky.Storage) error {

	elemTag := PluginTag(tag)
	elemType := elemTag.Get("")
	if elemType == "" {
		err := util.FormatError(nil, "found no plugin element")
		return util.WrapError(err, "inject struct field %s error", ft.Name)
	}

	elemType, nullable := strings.CutSuffix(elemType, "?")
	elemKey := prefix + "." + toCamelKey(elemType)

	switch fv.Kind() {
	case reflect.Slice:
		if !s.Has(elemKey) && !s.Has(elemKey+"[0]") {
			elemDef, ok := elemTag.Lookup("default")
			if !ok {
				if nullable {
					return nil
				}
				err := util.FormatError(nil, "found no plugin element")
				return util.WrapError(err, "inject struct field %s error", ft.Name)
			}

			// Parse default plugin type sequence (semicolon-separated)
			index := 0
			for typeClass := range strings.SplitSeq(elemDef, ";") {
				typeClass = strings.TrimSpace(typeClass)
				if typeClass == "" {
					continue
				}
				typeKey := elemKey + "[" + strconv.Itoa(index) + "].type"
				if err := s.Set(typeKey, typeClass, 0); err != nil {
					return err // Should never fail
				}
				index++
			}
			if index == 0 {
				err := util.FormatError(nil, "found no plugin element")
				return util.WrapError(err, "inject struct field %s error", ft.Name)
			}
		}

		slice := reflect.MakeSlice(ft.Type, 0, 1)
		if s.Has(elemKey + "[0]") { // Multiple elements
			for i := 0; ; i++ {
				subKey := elemKey + "[" + strconv.Itoa(i) + "]"
				if !s.Has(subKey) {
					break
				}

				const def = ":def:"
				strType := s.Get(subKey+".type", def)

				var (
					p  *Plugin
					ok bool
				)
				if strType != def {
					if p, ok = pluginRegistry[PluginType(toCamelKey(elemType))][strType]; !ok {
						err := util.FormatError(nil, "plugin %s not found", strType)
						return util.WrapError(err, "inject struct field %s error", ft.Name)
					}
				} else {
					if p, ok = pluginRegistry[PluginType(toCamelKey(elemType))][elemType]; !ok {
						err := util.FormatError(nil, "plugin %s not found", elemType)
						return util.WrapError(err, "inject struct field %s error", ft.Name)
					}
				}
				v, err := NewPlugin(p.Class, subKey, s)
				if err != nil {
					return util.WrapError(err, "inject struct field %s error", ft.Name)
				}
				slice = reflect.Append(slice, v)
			}

		} else if s.Has(elemKey) { // Single element
			const def = ":def:"
			strType := s.Get(elemKey+".type", def)

			var (
				p  *Plugin
				ok bool
			)
			if strType != def {
				if p, ok = pluginRegistry[PluginType(toCamelKey(elemType))][strType]; !ok {
					err := util.FormatError(nil, "plugin %s not found", strType)
					return util.WrapError(err, "inject struct field %s error", ft.Name)
				}
			} else {
				if p, ok = pluginRegistry[PluginType(toCamelKey(elemType))][elemType]; !ok {
					err := util.FormatError(nil, "plugin %s not found", elemType)
					return util.WrapError(err, "inject struct field %s error", ft.Name)
				}
			}
			v, err := NewPlugin(p.Class, elemKey, s)
			if err != nil {
				return util.WrapError(err, "inject struct field %s error", ft.Name)
			}
			slice = reflect.Append(slice, v)
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
				err := util.FormatError(nil, "found no plugin element")
				return util.WrapError(err, "inject struct field %s error", ft.Name)
			}
		} else {
			elemLabel, ok := elemTag.Lookup("default")
			if !ok {
				if nullable {
					return nil
				}
				err := util.FormatError(nil, "found no plugin element")
				return util.WrapError(err, "inject struct field %s error", ft.Name)
			}
			strType = elemLabel
		}

		p, ok := pluginRegistry[PluginType(toCamelKey(elemType))][strType]
		if !ok {
			err := util.FormatError(nil, "plugin %s not found", strType)
			return util.WrapError(err, "inject struct field %s error", ft.Name)
		}

		v, err := NewPlugin(p.Class, elemKey, s)
		if err != nil {
			return util.WrapError(err, "inject struct field %s error", ft.Name)
		}
		fv.Set(v)
		return nil

	default:
		err := util.FormatError(nil, "unsupported inject type %s", ft.Type.String())
		return util.WrapError(err, "inject struct field %s error", ft.Name)
	}
}
