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
	"io"
	"reflect"
	"strings"

	"github.com/go-spring/spring-base/barky"
	"github.com/lvan100/errutil"
)

// RootLoggerName defines the reserved name for the root logger.
const RootLoggerName = "root"

// global holds all runtime logger and appender instances.
var global struct {
	init      bool
	loggers   []Logger
	appenders []Appender
}

// RefreshFile loads a logging configuration from a file.
// The file format is automatically detected by its extension.
func RefreshFile(fileName string) error {
	s, err := readConfigFromFile(fileName)
	if err != nil {
		return errutil.Explain(err, "RefreshFile error")
	}
	return RefreshConfig(s)
}

// RefreshReader loads a logging configuration from an io.Reader.
// The `ext` specifies the configuration format (e.g., "yaml", "json").
func RefreshReader(r io.Reader, ext string) error {
	s, err := readConfigFromReader(r, ext)
	if err != nil {
		return errutil.Explain(err, "RefreshReader error")
	}
	return RefreshConfig(s)
}

// RefreshConfig loads a logging configuration from a *barky.Storage.
func RefreshConfig(s *barky.Storage) error {

	// Read appenders
	appenders, err := s.SubKeys("appender")
	if err != nil {
		return errutil.Stack(err, "read appenders section error")
	}
	if len(appenders) == 0 {
		return errutil.Explain(nil, "appenders section not found")
	}

	// Read loggers
	loggers, err := s.SubKeys("logger")
	if err != nil {
		return errutil.Stack(err, "read loggers section error")
	}

	// Ensure this refresh is executed only once
	if global.init {
		return errutil.Explain(nil, "log refresh already done")
	}
	global.init = true

	// Factory function to create plugin instances
	newPlugin := func(typ PluginType, typeKey string) (reflect.Value, error) {
		if !s.Has(typeKey) {
			return reflect.Value{}, errutil.Explain(nil, "attribute 'type' not found")
		}
		strType := s.Get(typeKey)
		p, ok := pluginRegistry[typ][strType]
		if !ok {
			return reflect.Value{}, errutil.Explain(nil, "plugin %s not found", strType)
		}
		return NewPlugin(p.Class, typeKey[:strings.LastIndex(typeKey, ".")], s)
	}

	// Initialize appender references in a logger
	initAppenderRefs := func(v reflect.Value, cAppenders map[string]Appender) (*LoggerBase, error) {
		var (
			base *LoggerBase
			ref  *AppenderRefs
		)
		switch config := v.Interface().(type) {
		case *SyncLogger:
			base = &config.LoggerBase
			ref = &config.AppenderRefs
		case *AsyncLogger:
			base = &config.LoggerBase
			ref = &config.AppenderRefs
		default: // for linter
		}
		for _, r := range ref.AppenderRefs {
			appender, ok := cAppenders[r.Ref]
			if !ok {
				return nil, errutil.Explain(nil, "appender %s not found", r.Ref)
			}
			r.Appender = appender
		}
		ref.sortByLevel()
		return base, nil
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
	for _, name := range appenders {
		v, err := newPlugin(PluginTypeAppender, "appender."+name+".type")
		if err != nil {
			return errutil.Stack(err, "create appender %s error", name)
		}
		cAppenders[name] = v.Interface().(Appender)
	}

	// Initialize all other loggers
	for _, name := range loggers {
		v, err := newPlugin(PluginTypeLogger, "logger."+name+".type")
		if err != nil {
			return errutil.Stack(err, "create logger %s error", name)
		}
		base, err := initAppenderRefs(v, cAppenders)
		if err != nil {
			return errutil.Stack(err, "init appender refs for logger %s error", name)
		}
		logger := v.Interface().(Logger)
		cLoggers[name] = logger

		// Skip the root logger
		if name == RootLoggerName {
			if base.Tags != "" {
				err = errutil.Explain(nil, "root logger must not define any tags")
				return errutil.Stack(err, "create logger %s error", name)
			}
			cRoot = logger
			continue
		}

		// Parse and validate tag list
		var tags []string
		for tag := range strings.SplitSeq(base.Tags, ",") {
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
			err = errutil.Explain(nil, "logger must have attribute 'tags'")
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
	var findLoggerForTag func(tag string) Logger
	findLoggerForTag = func(tag string) Logger {
		if l, ok := cTags[tag]; ok {
			return l
		}
		tag, _ = strings.CutSuffix(tag, "_*")
		i := strings.LastIndex(tag, "_")
		if i <= 0 {
			return cRoot
		}
		tag = strings.TrimSuffix(tag[:i], "_") + "_*"
		return findLoggerForTag(tag)
	}

	// Update `tagMap` with corresponding loggers
	for tag, obj := range tagRegistry {
		obj.logger = findLoggerForTag(tag)
	}

	// Inject properties
	for k, f := range propertyRegistry {
		if v := s.Get(toCamelKey(k)); v == "" {
			continue
		} else if err = f(v); err != nil {
			return errutil.Stack(err, "inject property %s error", k)
		}
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
	if !global.init {
		return
	}
	for _, l := range global.loggers {
		l.Stop()
	}
	for _, a := range global.appenders {
		a.Stop()
	}
	global.loggers = nil
	global.appenders = nil
	global.init = false
}
