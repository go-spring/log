/*
 * Copyright 2025 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 *
 * You may not use this file except in compliance with the License.
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
	"github.com/go-spring/stdlib/errutil"
)

// parseExpr expands inline map expressions embedded in map values.
//
// A key ending with "!" indicates that the corresponding value is a
// map expression. The trailing "!" is removed and the parsed entries
// are merged into the result using the "<key>.<subKey>" format.
//
// Example:
//
//	input:
//	  {
//	    "app.name": "MyApp",
//	    "db!": "{host: localhost, port: 5432}",
//	  }
//
//	output:
//	  {
//	    "app.name": "MyApp",
//	    "db.host":  "localhost",
//	    "db.port":  "5432",
//	  }
func parseExpr(m map[string]string) (map[string]string, error) {
	ret := make(map[string]string)
	for k, v := range m {
		var ok bool
		k, ok = strings.CutSuffix(k, "!")
		if !ok {
			ret[k] = v
			continue
		}
		// Parse the inline map expression
		subMap, err := expr.Parse(v)
		if err != nil {
			return nil, errutil.Explain(err, "parseExpr error")
		}
		for k2, v2 := range subMap {
			ret[k+"."+k2] = v2
		}
	}
	return ret, nil
}
