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
	"strings"

	"github.com/go-spring/log/expr"
	"github.com/lvan100/golib/errutil"
	"github.com/lvan100/golib/flatten"
)

// toStorage converts a flattened string map into a *flatten.Storage instance.
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
//	  flatten.Storage{
//	    "app.name": "MyApp",
//	    "db.host":  "localhost",
//	    "db.port":  "5432",
//	  }
func toStorage(m map[string]string) (*flatten.Storage, error) {
	s := flatten.NewStorage()
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
