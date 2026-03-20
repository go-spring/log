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
	"time"

	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/flatten"
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

func init() {
	RegisterConverter(func(s string) (time.Duration, error) {
		return time.ParseDuration(s)
	})
}

// Lifecycle is an optional interface for plugin lifecycle hooks.
type Lifecycle interface {
	Start() error
	Stop()
}

// pluginRegistry maintains a registry of available plugin classes,
// organized by plugin type and plugin name.
var pluginRegistry = map[string]*Plugin{}

// Plugin represents metadata about a plugin type.
type Plugin struct {
	Name  string       // Name of the plugin
	Class reflect.Type // Underlying struct type
	File  string       // File where plugin was registered
	Line  int          // Line number where plugin was registered
}

// RegisterPlugin registers a plugin struct type with a given name and plugin type.
func RegisterPlugin[T any](name string) {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic("T must be struct")
	}
	_, file, line, _ := runtime.Caller(1)
	if p, ok := pluginRegistry[name]; ok {
		err := errutil.Explain(nil, "duplicate plugin name %q in %s:%d and %s:%d",
			name, p.File, p.Line, file, line)
		panic(err)
	}
	pluginRegistry[name] = &Plugin{
		Name:  name,
		Class: t,
		File:  file,
		Line:  line,
	}
}

// NewPlugin creates a new plugin instance and injects configuration values.
func NewPlugin(t reflect.Type, prefix string, s flatten.Storage) (reflect.Value, error) {
	v := reflect.New(t)
	if err := inject(v.Elem(), t, prefix, s); err != nil {
		return reflect.Value{}, errutil.Stack(err, "create plugin %s error", t.String())
	}
	return v, nil
}

// inject recursively sets struct fields based on `PluginAttribute` and `PluginElement` tags.
func inject(v reflect.Value, t reflect.Type, prefix string, s flatten.Storage) error {
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

func resolvePropertyRef(s flatten.Storage, val string) (string, error) {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		key := val[2 : len(val)-1]
		str, ok := s.Value(key)
		if !ok {
			return "", errutil.Explain(nil, "property %s not found", val)
		}
		return str, nil
	}
	return val, nil
}

func convertAttributeValue(t reflect.Type, val string) (reflect.Value, error) {
	if fn := typeConverters[t]; fn != nil {
		fnValue := reflect.ValueOf(fn)
		out := fnValue.Call([]reflect.Value{reflect.ValueOf(val)})
		if !out[1].IsNil() {
			return reflect.Value{}, out[1].Interface().(error)
		}
		return out[0], nil
	}

	v := reflect.New(t).Elem()
	switch t.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 0, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetUint(u)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 0, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetInt(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetBool(b)
	case reflect.String:
		v.SetString(val)
	default:
		return reflect.Value{}, errutil.Explain(nil, "unsupported inject type %s", t.String())
	}
	return v, nil
}

//func getAttributeRawValue(s *flatten.Storage, key string, tag PluginTag) (string, error) {
//	if v, ok := s.RawData()[key]; ok {
//		return v.Value, nil
//	}
//	val, ok := tag.Lookup("default")
//	if !ok {
//		return "", errutil.Explain(nil, "found no attribute")
//	}
//	return val, nil
//}

func getArrayValues(s flatten.Storage, key string, attrTag PluginTag) ([]string, error) {
	delimiter := attrTag.Get("delimiter")
	if delimiter == "" {
		delimiter = ","
	}
	var values []string
	if s.Exists(key + "[0]") {
		// Indexed keys win over a delimited scalar value.
		for i := 0; ; i++ {
			subKey := fmt.Sprintf("%s[%d]", key, i)
			if !s.Exists(subKey) {
				break
			}
			str, ok := s.Value(subKey)
			if !ok {
				return nil, errutil.Explain(nil, "attribute %s not found", subKey)
			}
			v, err := resolvePropertyRef(s, str)
			if err != nil {
				return nil, err
			}
			for str := range strings.SplitSeq(v, delimiter) {
				if str = strings.TrimSpace(str); str == "" {
					continue
				}
				values = append(values, str)
			}
		}
		return values, nil
	}

	var val string
	if str, ok := s.Value(key); ok {
		val = str
	} else {
		str, ok = attrTag.Lookup("default")
		if !ok {
			return nil, errutil.Explain(nil, "found no attribute")
		}
		val = str
	}
	v, err := resolvePropertyRef(s, val)
	if err != nil {
		return nil, err
	}
	for str := range strings.SplitSeq(v, delimiter) {
		if str = strings.TrimSpace(str); str == "" {
			continue
		}
		values = append(values, str)
	}
	return values, nil
}

func injectArrayAttribute(fv reflect.Value, ft reflect.StructField, key string, attrTag PluginTag, s flatten.Storage) error {
	values, err := getArrayValues(s, key, attrTag)
	if err != nil {
		return err
	}
	switch fv.Kind() {
	case reflect.Slice:
		out := reflect.MakeSlice(ft.Type, len(values), len(values))
		elemType := ft.Type.Elem()
		for i, raw := range values {
			elem, err := convertAttributeValue(elemType, raw)
			if err != nil {
				return errutil.Stack(err, "inject struct field %s[%d] error", ft.Name, i)
			}
			out.Index(i).Set(elem)
		}
		fv.Set(out)
	case reflect.Array:
		if len(values) > fv.Len() {
			return errutil.Explain(nil, "too many values for array %s", ft.Type.String())
		}
		elemType := ft.Type.Elem()
		for i, raw := range values {
			elem, err := convertAttributeValue(elemType, raw)
			if err != nil {
				return errutil.Stack(err, "inject struct field %s[%d] error", ft.Name, i)
			}
			fv.Index(i).Set(elem)
		}
	default: // for linter
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
		ss := strings.SplitN(kvs[i], "=", 2)
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
func injectAttribute(tag string, fv reflect.Value, ft reflect.StructField, prefix string, s flatten.Storage) error {

	attrTag := PluginTag(tag)
	attrName := attrTag.Get("")
	if attrName == "" {
		err := errutil.Explain(nil, "found no attribute")
		return errutil.Stack(err, "inject struct field %s error", ft.Name)
	}

	// Special handling for "name" attribute
	if attrName == "name" {
		name := prefix[strings.LastIndex(prefix, ".")+1:]
		fv.SetString(name)
		return nil
	}

	key := prefix + "." + attrName

	if fv.Kind() == reflect.Slice || fv.Kind() == reflect.Array {
		if err := injectArrayAttribute(fv, ft, key, attrTag, s); err != nil {
			return errutil.Stack(err, "inject struct field %s error", ft.Name)
		}
		return nil
	}

	var val string
	if str, ok := s.Value(key); ok {
		val = str
	} else {
		str, ok = attrTag.Lookup("default")
		if !ok {
			err := errutil.Explain(nil, "found no attribute")
			return errutil.Stack(err, "inject struct field %s error", ft.Name)
		}
		val = str
	}

	val, err := resolvePropertyRef(s, val)
	if err != nil {
		return errutil.Stack(err, "inject struct field %s error", ft.Name)
	}

	v, err := convertAttributeValue(ft.Type, val)
	if err != nil {
		return errutil.Stack(err, "inject struct field %s error", ft.Name)
	}
	fv.Set(v)
	return nil
}

// injectElement injects child plugin elements into a struct field.
func injectElement(tag string, fv reflect.Value, ft reflect.StructField, prefix string, s flatten.Storage) error {

	elemKey := PluginTag(tag).Get("")
	if elemKey == "" {
		err := errutil.Explain(nil, "found no plugin element")
		return errutil.Stack(err, "inject struct field %s error", ft.Name)
	}

	elemKey, nullable := strings.CutSuffix(elemKey, "?")
	elemKey = prefix + "." + elemKey

	switch fv.Kind() {
	case reflect.Slice:
		slice := reflect.MakeSlice(ft.Type, 0, 1)
		if s.Exists(elemKey + "[0]") { // Multiple elements
			for i := 0; ; i++ {
				subKey := elemKey + "[" + strconv.Itoa(i) + "]"
				if !s.Exists(subKey) {
					break
				}
				strType, _ := s.Value(subKey + ".type")
				// 这里允许 strType 为空，对于结构体指针类型而言
				v, err := createPlugin(ft, subKey, strType, s)
				if err != nil {
					return errutil.Stack(err, "inject struct field %s error", ft.Name)
				}
				slice = reflect.Append(slice, v)
			}
		} else if s.Exists(elemKey) { // Single element
			strType, _ := s.Value(elemKey + ".type")
			// 这里允许 strType 为空，对于结构体指针类型而言
			v, err := createPlugin(ft, elemKey, strType, s)
			if err != nil {
				return errutil.Stack(err, "inject struct field %s error", ft.Name)
			}
			slice = reflect.Append(slice, v)
		}

		if slice.Len() == 0 {
			elemDef, ok := PluginTag(tag).Lookup("default")
			if !ok {
				if nullable {
					return nil
				}
				err := errutil.Explain(nil, "found no plugin element")
				return errutil.Stack(err, "inject struct field %s error", ft.Name)
			}

			var defaultTypes [][]string

			// Parse default plugin type sequence (semicolon-separated)
			index := 0
			for typeClass := range strings.SplitSeq(elemDef, ";") {
				typeClass = strings.TrimSpace(typeClass)
				if typeClass == "" {
					continue
				}
				typeKey := elemKey + "[" + strconv.Itoa(index) + "]"
				defaultTypes = append(defaultTypes, []string{typeKey, typeClass})
				index++
			}
			if index == 0 {
				err := errutil.Explain(nil, "found no plugin element")
				return errutil.Stack(err, "inject struct field %s error", ft.Name)
			}

			for _, arr := range defaultTypes {
				v, err := createPlugin(ft, arr[0], arr[1], s)
				if err != nil {
					return errutil.Stack(err, "inject struct field %s error", ft.Name)
				}
				slice = reflect.Append(slice, v)
			}
		}

		fv.Set(slice)
		return nil

	case reflect.Interface:
		strType, ok := s.Value(elemKey + ".type")
		if !ok {
			typeClass, ok := PluginTag(tag).Lookup("default")
			if !ok {
				if nullable {
					return nil
				}
				err := errutil.Explain(nil, "found no plugin element")
				return errutil.Stack(err, "inject struct field %s error", ft.Name)
			}
			strType = typeClass
		}

		v, err := createPlugin(ft, elemKey, strType, s)
		if err != nil {
			return errutil.Stack(err, "inject struct field %s error", ft.Name)
		}
		fv.Set(v)
		return nil

	default:
		err := errutil.Explain(nil, "unsupported inject type %s", ft.Type.String())
		return errutil.Stack(err, "inject struct field %s error", ft.Name)
	}
}

// createPlugin creates a plugin instance.
func createPlugin(ft reflect.StructField, elemKey string, strType string, s flatten.Storage) (reflect.Value, error) {
	//strType, ok := s.Value(elemKey + ".type")
	//if !ok {
	//	return reflect.Value{}, errutil.Explain(nil, "found no plugin element")
	//}
	p, ok := pluginRegistry[strType]
	if !ok {
		if ft.Type.Kind() == reflect.Interface {
			if strType == "" {
				return reflect.Value{}, errutil.Explain(nil, "found no plugin element")
			}
			return reflect.Value{}, errutil.Explain(nil, "plugin %s not found", strType)
		}
		p = &Plugin{Class: ft.Type.Elem()}
		for p.Class.Kind() == reflect.Pointer {
			p.Class = p.Class.Elem()
		}
		if p.Class.Kind() != reflect.Struct {
			return reflect.Value{}, errutil.Explain(nil, "plugin %s not found", strType)
		}
	}
	return NewPlugin(p.Class, elemKey, s)
}
