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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-spring/barky"
	"github.com/magiconair/properties"
	"gopkg.in/yaml.v2"
)

var readers = map[string]Reader{}

func init() {
	RegisterReader(ReadXML, ".xml")
	RegisterReader(ReadJSON, ".json")
	RegisterReader(ReadYAML, ".yml", ".yaml")
	RegisterReader(ReadProperties, ".properties")
}

// Reader defines a function that converts a byte slice into a map.
type Reader func([]byte) (map[string]any, error)

// RegisterReader registers a Reader function for one or more file extensions.
func RegisterReader(r Reader, ext ...string) {
	for _, s := range ext {
		readers[s] = r
	}
}

// readConfigFromFile reads a configuration file and returns a *barky.Storage.
func readConfigFromFile(fileName string) (*barky.Storage, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	// nolint: errcheck
	defer file.Close()
	ext := filepath.Ext(fileName)
	return readConfigFromReader(file, ext)
}

// readConfigFromReader reads configuration from an io.Reader given a file extension.
func readConfigFromReader(reader io.Reader, ext string) (*barky.Storage, error) {
	r, ok := readers[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported file type %s", ext)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	m, err := r(data)
	if err != nil {
		return nil, err
	}

	// Flatten nested maps and convert to *barky.Storage
	return toStorage(barky.FlattenMap(m))
}

// ReadProperties parses a properties file into a map.
func ReadProperties(b []byte) (map[string]any, error) {
	p := properties.NewProperties()
	p.DisableExpansion = true
	if err := p.Load(b, properties.UTF8); err != nil {
		return nil, fmt.Errorf("ReadProperties error: %w", err)
	}
	r := make(map[string]any)
	for k, v := range p.Map() {
		r[k] = v
	}
	return r, nil
}

// ReadJSON parses a JSON file into a map.
func ReadJSON(b []byte) (map[string]any, error) {
	var r map[string]any
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("ReadJSON error: %w", err)
	}
	return r, nil
}

// ReadYAML parses a YAML file into a map.
func ReadYAML(b []byte) (map[string]any, error) {
	var r map[string]any
	if err := yaml.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("ReadYAML error: %w", err)
	}
	return r, nil
}

// ReadXML parses an XML configuration file into a map.
func ReadXML(b []byte) (map[string]any, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	m := make(map[string]any)
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ReadXML error: %w", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Configuration":
				// root element, skip
			case "Properties", "Appenders", "Loggers":
				s := make(map[string]any)
				if err = xmlToMap(s, d); err != nil {
					return nil, fmt.Errorf("ReadXML error: %w", err)
				}
				m[t.Name.Local] = s
			default:
				err = fmt.Errorf("unsupported xml tag %s", t.Name.Local)
				return nil, fmt.Errorf("ReadXML error: %w", err)
			}
		default: // for linter
		}
	}

	r := make(map[string]any)

	// flatten Properties
	if p, ok := m["Properties"]; ok {
		for k, v := range p.(map[string]any) {
			r[k], _ = v.(map[string]any)["Text"]
		}
	}

	// validate Appenders
	if a, ok := m["Appenders"]; !ok {
		return nil, fmt.Errorf("missing Appenders")
	} else {
		r["Appender"] = a
	}

	// validate Loggers
	if l, ok := m["Loggers"]; !ok {
		return nil, fmt.Errorf("missing Loggers")
	} else {
		a := l.(map[string]any)
		root, _ := a["Root"]
		asyncRoot, _ := a["AsyncRoot"]

		var s map[string]any
		if root != nil {
			if asyncRoot != nil {
				return nil, errors.New("found multiple root loggers")
			}
			s = root.(map[string]any)
		} else {
			if asyncRoot == nil {
				return nil, fmt.Errorf("missing Root or AsyncRoot")
			}
			s = asyncRoot.(map[string]any)
		}

		// remove root from Loggers map
		delete(a, s["Type"].(string))

		r["rootLogger"] = s
		r["Logger"] = l
	}

	return r, nil
}

// xmlToMap recursively converts XML elements into map[string]any
func xmlToMap(m map[string]any, d *xml.Decoder) error {
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}
		switch t := token.(type) {
		case xml.StartElement:
			p, ok := plugins[t.Name.Local]
			if !ok {
				return fmt.Errorf("unsupported xml tag %s", t.Name.Local)
			}

			s := map[string]any{
				"Type": t.Name.Local,
			}

			var name string
			for _, attr := range t.Attr {
				if attr.Name.Local == "name" {
					name = attr.Value
					continue
				}
				s[attr.Name.Local] = attr.Value
			}

			if name == "" {
				strType := string(p.Type)
				v, ok := m[strType]
				if ok {
					switch v.(type) {
					case []any:
						m[strType] = append(v.([]any), s)
					default:
						m[strType] = []any{v, s}
					}
				} else {
					m[strType] = s
				}
			} else {
				m[name] = s
			}

			if err = xmlToMap(s, d); err != nil {
				return err
			}

		case xml.CharData:
			if text := strings.TrimSpace(string(t)); text != "" {
				s, _ := m["Text"].(string)
				m["Text"] = s + text
			}
		case xml.EndElement:
			return nil
		default: // for linter
		}
	}
}

// toStorage converts a flattened map into a *barky.Storage instance.
func toStorage(m map[string]string) (*barky.Storage, error) {
	s := barky.NewStorage()
	for k, v := range m {
		if err := s.Set(toCamelKey(k), v, 0); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// toCamelKey converts a string like "foo_bar-baz" into "fooBarBaz".
func toCamelKey(key string) string {
	if key == "" {
		return ""
	}

	const offset = 'a' - 'A'

	b := []byte(key)
	r := make([]byte, 0, len(b))

	c := b[0]
	if c >= 'A' && c <= 'Z' {
		c += offset
	}
	r = append(r, c)

	lowerNext := false
	upperNext := false
	for i := 1; i < len(b); i++ {
		c = b[i]
		if c == '.' {
			lowerNext = true
			r = append(r, c)
			continue
		} else if c == '-' || c == '_' {
			upperNext = true
			continue
		}
		if lowerNext {
			if c >= 'A' && c <= 'Z' {
				c += offset
			}
			lowerNext = false
		} else if upperNext {
			if c >= 'a' && c <= 'z' {
				c -= offset
			}
			upperNext = false
		}
		r = append(r, c)
	}
	return string(r)
}
