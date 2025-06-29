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
	"cmp"
	"fmt"
	"slices"
)

// IntType is the type of int, int8, int16, int32, int64.
type IntType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// UintType is the type of uint, uint8, uint16, uint32, uint64.
type UintType interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// FloatType is the type of float32, float64.
type FloatType interface {
	~float32 | ~float64
}

// Ptr returns a pointer to the given value.
func Ptr[T any](i T) *T {
	return &i
}

// MapKeys returns the keys of the map m.
func MapKeys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

// OrderedMapKeys returns the keys of the map m in sorted order.
func OrderedMapKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	r := MapKeys(m)
	slices.Sort(r)
	return r
}

// WrapError wraps an existing error, creating a new error with hierarchical relationships.
func WrapError(err error, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s << %w", msg, err)
}
