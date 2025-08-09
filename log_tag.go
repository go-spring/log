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
	"sync/atomic"
)

var tagMap = map[string]*Tag{}

var initLogger Logger = &SyncLogger{
	BaseLogger: BaseLogger{
		Level: InfoLevel,
		AppenderRefs: []*AppenderRef{
			{
				appender: &ConsoleAppender{
					BaseAppender: BaseAppender{
						Layout: &TextLayout{
							BaseLayout: BaseLayout{
								FileLineLength: 48,
							},
						},
					},
				},
			},
		},
	},
}

// LoggerHolder is a wrapper struct used to store a Logger in an atomic.Value,
// ensuring type consistency and safe concurrent access.
type LoggerHolder struct {
	Logger
}

// Tag is a struct representing a named logging tag.
// It holds a pointer to a Logger and a string identifier.
type Tag struct {
	logger atomic.Value
	name   string
}

// getLogger returns the Logger associated with this tag.
// It uses atomic loading to ensure safe concurrent access.
func (m *Tag) getLogger() Logger {
	return m.logger.Load().(LoggerHolder)
}

// setLogger sets or replaces the Logger associated with this tag.
// Uses atomic storing to ensure thread safety.
func (m *Tag) setLogger(logger Logger) {
	m.logger.Store(LoggerHolder{logger})
}

// GetAllTags returns all registered tags.
func GetAllTags() []string {
	return OrderedMapKeys(tagMap)
}

// isValidTag checks whether the tag is valid according to the following rules:
// 1. The length must be between 3 and 36 characters.
// 2. Only lowercase letters (a-z), digits (0-9), and underscores (_) are allowed.
// 3. The tag can start with an underscore.
// 4. Underscores separate the tag into 1 to 4 non-empty segments.
// 5. No empty segments are allowed (i.e., no consecutive or trailing underscores).
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

// RegisterTag creates or retrieves a Tag by name.
// If the tag does not exist in the global registry, it is created and associated with a default logger.
// Normally, you should use GetAppTag, GetBizTag, or GetRPCTag to create tags semantically.
func RegisterTag(tag string) *Tag {
	if global.init.Load() {
		panic("log refresh already done")
	}
	if !isValidTag(tag) {
		panic("invalid tag name")
	}
	m, ok := tagMap[tag]
	if !ok {
		m = &Tag{name: tag}
		m.setLogger(initLogger)
		tagMap[tag] = m
	}
	return m
}

// BuildTag constructs a structured tag string based on main type, sub type, and action.
// The format is: _<mainType>_<subType> or _<mainType>_<subType>_<action>.
func BuildTag(mainType, subType, action string) string {
	if subType == "" {
		panic("subType cannot be empty")
	}
	if action == "" {
		return "_" + mainType + "_" + subType
	}
	return "_" + mainType + "_" + subType + "_" + action
}

// RegisterAppTag returns a Tag used for application-layer logs (e.g., framework events, lifecycle).
// subType represents the component or module, action represents the lifecycle phase or behavior.
func RegisterAppTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("app", subType, action))
}

// RegisterBizTag returns a Tag used for business-logic logs (e.g., use cases, domain events).
// subType is the business domain or feature name, action is the operation being logged.
func RegisterBizTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("biz", subType, action))
}

// RegisterRPCTag returns a Tag used for RPC or external/internal dependency logs.
// subType is the protocol or target system, action is the RPC phase (e.g., sent, retry, fail).
func RegisterRPCTag(subType, action string) *Tag {
	return RegisterTag(BuildTag("rpc", subType, action))
}
