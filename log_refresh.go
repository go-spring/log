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
	"strings"

	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/flatten"
)

// RootLoggerName defines the reserved name for the root logger.
const RootLoggerName = "root"

// global holds all runtime logger and appender instances.
var global struct {
	refreshed bool // 目前暂不实现动态刷新的能力
	loggers   []Logger
	appenders []Appender
}

// RefreshConfig loads a logging configuration from a map[string]string.
func RefreshConfig(m map[string]string) error {
	m, err := parseExpr(m)
	if err != nil {
		return err
	}
	p := flatten.NewProperties(m)
	return Refresh(flatten.NewPropertiesStorage(p))
}

// Refresh loads a logging configuration from a *flatten.Storage.
func Refresh(s flatten.Storage) error {

	// Ensure this refresh is executed only once
	if global.refreshed {
		return errutil.Explain(nil, "log refresh already done")
	}
	global.refreshed = true

	appenders := make(map[string]struct{})
	s.MapKeys("appender", appenders)
	if len(appenders) == 0 {
		return errutil.Explain(nil, "appenders section not found")
	}

	// Read loggers
	loggers := make(map[string]struct{})
	s.MapKeys("logger", loggers)

	// Factory function to create plugin instances
	newPlugin := func(typeKey string) (reflect.Value, error) {
		if !s.Exists(typeKey) {
			return reflect.Value{}, errutil.Explain(nil, "attribute 'type' not found")
		}
		strType, ok := s.Value(typeKey)
		if !ok {
			return reflect.Value{}, errutil.Explain(nil, "attribute 'type' not found")
		}
		p, ok := pluginRegistry[strType]
		if !ok {
			return reflect.Value{}, errutil.Explain(nil, "plugin %s not found", strType)
		}
		return NewPlugin(p.Class, typeKey[:strings.LastIndex(typeKey, ".")], s)
	}

	var (
		cRoot      = defaultLogger
		cLoggers   = make(map[string]Logger)
		cAppenders = make(map[string]Appender)
		cTags      = make(map[string]Logger)
	)

	// Set root logger
	cLoggers[RootLoggerName] = cRoot

	// Initialize appenders
	for name := range appenders {
		v, err := newPlugin("appender." + name + ".type")
		if err != nil {
			return errutil.Stack(err, "create appender %s error", name)
		}
		cAppenders[name] = v.Interface().(Appender)
	}

	// Initialize all other loggers
	for name := range loggers {
		v, err := newPlugin("logger." + name + ".type")
		if err != nil {
			return errutil.Stack(err, "create logger %s error", name)
		}

		if err = initAppenderRefs(v, cAppenders); err != nil {
			return errutil.Stack(err, "init appender refs for logger %s error", name)
		}
		logger := v.Interface().(Logger)
		cLoggers[name] = logger

		// Skip the root logger
		if name == RootLoggerName {
			if len(logger.GetTags()) > 0 {
				err = errutil.Explain(nil, "root logger must not define any tags")
				return errutil.Stack(err, "create logger %s error", name)
			}
			cRoot = logger
			continue
		}

		// Parse and validate tag list
		var tags []string
		for _, tag := range logger.GetTags() {
			if tag = strings.TrimSpace(tag); tag == "" {
				continue
			}
			if strings.Contains(tag, "*") {
				if !strings.HasSuffix(tag, "_*") {
					err = errutil.Explain(nil, "tag '%s' is invalid", tag)
					return errutil.Stack(err, "create logger %s error", name)
				}
			}
			tags = append(tags, tag)
		}

		if len(tags) == 0 {
			err = errutil.Explain(nil, "logger must have attribute 'tag'")
			return errutil.Stack(err, "create logger %s error", name)
		}

		for _, strTag := range tags {
			if l, ok := cTags[strTag]; ok && l != logger {
				err = errutil.Explain(nil, "tag '%s' already config in logger %s", strTag, l)
				return errutil.Stack(err, "create logger %s error", name)
			}
			cTags[strTag] = logger
		}
	}

	// Start all appenders
	for _, a := range cAppenders {
		if err := a.Start(); err != nil {
			return errutil.Stack(err, "appender %s start error", a.GetName())
		}
	}

	// Start all loggers
	for _, l := range cLoggers {
		if err := l.Start(); err != nil {
			return errutil.Stack(err, "logger %s start error", l.GetName())
		}
	}

	// Update logger references in `loggerMap`
	for _, l := range loggerMap {
		v, ok := cLoggers[l.name]
		if !ok {
			return errutil.Explain(nil, "logger %s not found", l.name)
		}
		l.logger = v
	}

	// Helper to find the most appropriate logger for a given tag.
	findLoggerForTag := func(tag string) Logger {
		for {
			if l, ok := cTags[tag]; ok {
				return l
			}
			tag, _ = strings.CutSuffix(tag, "_*")
			i := strings.LastIndex(tag, "_")
			if i <= 0 {
				return cRoot
			}
			tag = tag[:i-1] + "_*" // 去掉最后一段
		}
	}

	// Update `tagMap` with corresponding loggers
	for tag, obj := range tagRegistry {
		obj.logger.Store(&loggerValue{findLoggerForTag(tag)})
	}

	// Update global loggers and appenders
	for _, l := range cLoggers {
		global.loggers = append(global.loggers, l)
	}
	for _, a := range cAppenders {
		global.appenders = append(global.appenders, a)
	}

	return nil
}

// Destroy gracefully shuts down all loggers and appenders,
// releasing resources and resetting the global state.
func Destroy() {
	if !global.refreshed {
		return
	}
	for _, obj := range tagRegistry {
		obj.reset()
	}
	for _, l := range global.loggers {
		l.Stop()
	}
	for _, a := range global.appenders {
		a.Stop()
	}
	global.loggers = nil
	global.appenders = nil
	global.refreshed = false
}

// initAppenderRefs Initialize appender references in a logger
func initAppenderRefs(v reflect.Value, cAppenders map[string]Appender) error {
	i, ok := v.Interface().(AppenderRefs)
	if !ok {
		return nil
	}
	syncMode, appenderRefs := i.GetAppenderRefs()
	for _, r := range appenderRefs {
		a, ok := cAppenders[r.Ref]
		if !ok {
			return errutil.Explain(nil, "appender %s not found", r.Ref)
		}
		// 同步 logger 必须搭配并发安全的 appender
		if syncMode && !a.ConcurrentSafe() {
			return errutil.Explain(nil, "appender %s is not concurrent-safe", r.Ref)
		}
		r.Appender = a
	}
	return nil
}
