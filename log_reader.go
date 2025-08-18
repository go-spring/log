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
	RegisterReader(ReadXml, ".xml")
	RegisterReader(ReadJSON, ".json")
	RegisterReader(ReadYaml, ".yml", ".yaml")
	RegisterReader(ReadProperties, ".properties")
}

type Reader func([]byte) (map[string]any, error)

// RegisterReader registers a Reader for one or more file extensions.
// This allows dynamic selection of parsers based on file type.
func RegisterReader(r Reader, ext ...string) {
	for _, s := range ext {
		readers[s] = r
	}
}

// readConfigFromFile reads a file and returns a Storage.
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

// readConfigFromReader reads a file and returns a Storage.
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
	return toStorage(barky.FlattenMap(m))
}

// ReadProperties reads a properties file and returns a Storage.
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

// ReadJSON reads a json file and returns a Storage.
func ReadJSON(b []byte) (map[string]any, error) {
	var r map[string]any
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("ReadJSON error: %w", err)
	}
	return r, nil
}

// ReadYaml reads a yaml file and returns a Storage.
func ReadYaml(b []byte) (map[string]any, error) {
	var r map[string]any
	if err := yaml.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("ReadYaml error: %w", err)
	}
	return r, nil
}

// ReadXml reads an xml file and returns a Storage.
func ReadXml(b []byte) (map[string]any, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	m := make(map[string]any)
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ReadXml error: %w", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Configuration":
			case "Properties", "Appenders", "Loggers":
				s := make(map[string]any)
				m[t.Name.Local] = s
				if err = xmlToMap(s, d); err != nil {
					return nil, fmt.Errorf("ReadXml error: %w", err)
				}
			default:
				err = fmt.Errorf("unsupported xml tag %s", t.Name.Local)
				return nil, fmt.Errorf("ReadXml error: %w", err)
			}
		default: // for linter
		}
	}

	r := make(map[string]any)

	if p, ok := m["Properties"]; ok {
		for k, v := range p.(map[string]any) {
			r[k] = v.(map[string]any)["Text"]
		}
	}

	if a, ok := m["Appenders"]; !ok {
		err := fmt.Errorf("missing Appenders")
		return nil, fmt.Errorf("ReadXml error: %w", err)
	} else {
		r["Appender"] = a
	}

	if l, ok := m["Loggers"]; !ok {
		err := fmt.Errorf("missing Loggers")
		return nil, fmt.Errorf("ReadXml error: %w", err)
	} else {
		a := l.(map[string]any)
		root := a["Root"]
		asyncRoot := a["AsyncRoot"]

		var s map[string]any
		if root != nil {
			if asyncRoot != nil {
				err := errors.New("found multiple root loggers")
				return nil, fmt.Errorf("ReadXml error: %w", err)
			}
			s = root.(map[string]any)
		} else {
			if asyncRoot == nil {
				err := fmt.Errorf("missing Root or AsyncRoot")
				return nil, fmt.Errorf("ReadXml error: %w", err)
			}
			s = asyncRoot.(map[string]any)
		}
		delete(a, s["Type"].(string))
		r["rootLogger"] = s
		r["Logger"] = l
	}

	return r, nil
}

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

// toStorage converts a map to a Storage.
func toStorage(m map[string]string) (*barky.Storage, error) {
	s := barky.NewStorage()
	for k, v := range m {
		if err := s.Set(ToCamelKey(k), v, 0); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// ToCamelKey converts a string to camel case.
func ToCamelKey(key string) string {
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
				c = c - offset
			}
			upperNext = false
		}
		r = append(r, c)
	}
	return string(r)
}
