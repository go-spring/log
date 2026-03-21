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
	"sync"

	"github.com/go-spring/log/expr"
	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/flatten"
)

// RootLoggerName defines the reserved name for the root logger.
// This is the default logger used when no specific logger is matched.
const RootLoggerName = "root"

// global holds all runtime logger and appender instances.
var global struct {
	mutex     sync.Mutex
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

// Refresh loads a logging configuration from a *flatten.Storage.
func Refresh(s flatten.Storage) error {

	global.mutex.Lock()
	defer global.mutex.Unlock()

	// Refresh can only happen once; prevent multiple refreshes
	if len(global.loggers) > 0 {
		return errutil.Explain(nil, "can only refresh once")
	}

	loggers := make(map[string]struct{})
	s.MapKeys("logger", loggers)

	appenders := make(map[string]struct{})
	s.MapKeys("appender", appenders)
	if len(appenders) == 0 {
		return errutil.Explain(nil, "appenders section not found")
	}

	newPlugin := func(typeKey string) (reflect.Value, error) {
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

	for name := range appenders {
		v, err := newPlugin("appender." + name + ".type")
		if err != nil {
			return errutil.Stack(err, "create appender %s error", name)
		}
		cAppenders[name] = v.Interface().(Appender)
	}

	initAppenderRefs := func(v reflect.Value) error {
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
			// If synchronization mode is enabled, the appender must be concurrent-safe
			if syncMode && !a.ConcurrentSafe() {
				return errutil.Explain(nil, "appender %s is not concurrent-safe", r.Ref)
			}
			r.Appender = a
		}
		return nil
	}

	cLoggers[RootLoggerName] = cRoot
	for name := range loggers {
		v, err := newPlugin("logger." + name + ".type")
		if err != nil {
			return errutil.Stack(err, "create logger %s error", name)
		}

		if err = initAppenderRefs(v); err != nil {
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

	for _, a := range cAppenders {
		if err := a.Start(); err != nil {
			return errutil.Stack(err, "appender %s start error", a.GetName())
		}
	}

	for _, l := range cLoggers {
		if err := l.Start(); err != nil {
			return errutil.Stack(err, "logger %s start error", l.GetName())
		}
	}

	for _, l := range loggerMap {
		v, ok := cLoggers[l.name]
		if !ok {
			return errutil.Explain(nil, "logger %s not found", l.name)
		}
		l.logger.Store(&loggerValue{v})
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
			tag = tag[:i-1] + "_*"
		}
	}

	for tag, obj := range tagRegistry {
		v := findLoggerForTag(tag)
		obj.logger.Store(&loggerValue{v})
	}

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
	global.mutex.Lock()
	defer global.mutex.Unlock()

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
}
