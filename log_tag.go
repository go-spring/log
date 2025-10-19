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
	"slices"
	"strings"

	"github.com/go-spring/spring-base/util"
)

// tagRegistry stores Tag instances keyed by their string names.
// Note: Not concurrency-safe; intended for use during initialization.
var tagRegistry = map[string]*Tag{}

// Tag represents a named log tag, used to categorize logs by
// subsystem, business domain, or RPC interaction, and so on.
// Each Tag holds a reference to a Logger.
type Tag struct {
	tag    string
	logger Logger
}

// GetAllTags returns the names of all registered tags.
func GetAllTags() []string {
	return util.OrderedMapKeys(tagRegistry)
}

// isValidTag validates a tag string according to the following rules:
//  1. Length must be between 3 and 36 characters.
//  2. Allowed characters: lowercase letters (a-z), digits (0-9), underscores (_).
//  3. Tag may start with an underscore.
//  4. Segments are separated by underscores, with 1 to 4 non-empty segments.
//  5. No consecutive or trailing underscores are allowed.
func isValidTag(tag string) bool {
	if len(tag) < 3 || len(tag) > 36 {
		return false
	}
	for i := range len(tag) {
		c := tag[i]
		// nolint: staticcheck
		if !(c >= 'a' && c <= 'z') && !(c >= '0' && c <= '9') && c != '_' {
			return false
		}
	}
	ss := strings.Split(strings.TrimPrefix(tag, "_"), "_")
	if len(ss) < 1 || len(ss) > 4 {
		return false
	}
	return !slices.Contains(ss, "")
}

// RegisterTag retrieves or creates a Tag by name.
// If the tag is not yet registered, it is associated with initLogger.
// Panics if called after global.init is set (i.e., after logging refresh).
//
// Normally, higher-level helpers like RegisterAppTag, RegisterBizTag,
// or RegisterRPCTag should be used to ensure semantic consistency.
func RegisterTag(tag string) *Tag {
	if global.init {
		panic("log refresh already done")
	}
	if !isValidTag(tag) {
		panic("invalid log tag")
	}
	m, ok := tagRegistry[tag]
	if !ok {
		m = &Tag{tag: tag}
		tagRegistry[tag] = m
	}
	return m
}

// BuildTag constructs a structured tag string from main type, sub type, and action.
// The format is:
//
//	_<mainType>_<subType>
//	_<mainType>_<subType>_<action>
//
// Example:
//
//	BuildTag("app", "startup", "init") -> "_app_startup_init"
func BuildTag(mainType, subType, action string) string {
	if subType == "" {
		panic("subType cannot be empty")
	}
	if action == "" {
		return "_" + mainType + "_" + subType
	}
	return "_" + mainType + "_" + subType + "_" + action
}

// RegisterAppTag registers or retrieves a Tag intended for application-layer logs.
//   - subType: component or module name
//   - action: lifecycle phase or behavior (optional)
func RegisterAppTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("app", subType, action))
}

// RegisterBizTag registers or retrieves a Tag intended for business-logic logs.
//   - subType: business domain or feature name
//   - action: operation being logged (optional)
func RegisterBizTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("biz", subType, action))
}

// RegisterRPCTag registers or retrieves a Tag intended for RPC logs,
// covering external/internal dependency interactions.
//   - subType: protocol or target system (e.g., http, grpc, redis)
//   - action: RPC phase (e.g., send, retry, fail)
func RegisterRPCTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("rpc", subType, action))
}
