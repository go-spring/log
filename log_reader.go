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
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-spring/log/expr"
	"github.com/go-spring/spring-base/barky"
	"github.com/lvan100/errutil"
	"github.com/magiconair/properties"
	"gopkg.in/yaml.v2"
)

var fileReaders = map[string]Reader{}

func init() {
	RegisterReader(ReadJSON, ".json")
	RegisterReader(ReadYAML, ".yml", ".yaml")
	RegisterReader(ReadProperties, ".properties")
}

// Reader defines a function that converts a byte slice into a map.
type Reader func([]byte) (map[string]any, error)

// RegisterReader registers a Reader function for one or more file extensions.
func RegisterReader(r Reader, ext ...string) {
	for _, s := range ext {
		fileReaders[strings.ToLower(s)] = r
	}
}

// readConfigFromFile reads a configuration file and returns a *barky.Storage.
func readConfigFromFile(fileName string) (*barky.Storage, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	ext := strings.ToLower(filepath.Ext(fileName))
	return readConfigFromReader(file, ext)
}

// readConfigFromReader reads configuration from an io.Reader given a file extension.
func readConfigFromReader(reader io.Reader, ext string) (*barky.Storage, error) {
	r, ok := fileReaders[ext]
	if !ok {
		return nil, errutil.Explain(nil, "unsupported file type %s", ext)
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
		return nil, errutil.Explain(err, "ReadProperties error")
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
		return nil, errutil.Explain(err, "ReadJSON error")
	}
	return r, nil
}

// ReadYAML parses a YAML file into a map.
func ReadYAML(b []byte) (map[string]any, error) {
	var r map[string]any
	if err := yaml.Unmarshal(b, &r); err != nil {
		return nil, errutil.Explain(err, "ReadYAML error")
	}
	return r, nil
}

// toStorage converts a flattened string map into a *barky.Storage instance.
// It interprets keys with a "!" suffix as nested map structures, parsing their
// values using expr.Parse() and inserting them under a composite key.
//
// Example:
//
//	Input:
//	  map[string]string{
//	    "app.name": "MyApp",
//	    "db!": "{host: localhost, port: 5432}",
//	  }
//	Output:
//	  barky.Storage{
//	    "app.name": "MyApp",
//	    "db.host":  "localhost",
//	    "db.port":  "5432",
//	  }
func toStorage(m map[string]string) (*barky.Storage, error) {
	s := barky.NewStorage()
	for k, v := range m {
		var ok bool
		// Normalize key to CamelCase and check for "!" suffix,
		// which indicates that the value represents a nested map expression.
		if k, ok = strings.CutSuffix(toCamelKey(k), "!"); ok {
			subMap, err := expr.Parse(v)
			if err != nil {
				return nil, errutil.Explain(err, "toStorage error")
			}
			for k2, v2 := range subMap {
				if err = s.Set(k+"."+toCamelKey(k2), v2, 0); err != nil {
					return nil, errutil.Explain(err, "toStorage error")
				}
			}
		} else {
			if err := s.Set(k, v, 0); err != nil {
				return nil, errutil.Explain(err, "toStorage error")
			}
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
