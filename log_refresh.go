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
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/go-spring/spring-base/barky"
	"github.com/go-spring/spring-base/util"
)

// global holds the global state of loggers and appenders.
var global struct {
	init      atomic.Bool
	loggers   []Logger
	appenders []Appender
}

const RootLoggerName = "::ROOT::"

// RefreshFile loads a logging configuration from a file by its name.
func RefreshFile(fileName string) error {
	s, err := readConfigFromFile(fileName)
	if err != nil {
		return util.FormatError(err, "RefreshFile error")
	}
	return RefreshConfig(s)
}

// RefreshReader loads a logging configuration from an io.Reader.
func RefreshReader(r io.Reader, ext string) error {
	s, err := readConfigFromReader(r, ext)
	if err != nil {
		return util.FormatError(err, "RefreshReader error")
	}
	return RefreshConfig(s)
}

// RefreshConfig loads a logging configuration from a *barky.Storage object.
func RefreshConfig(s *barky.Storage) error {

	// Read appenders
	appenders, err := s.SubKeys("appender")
	if err != nil {
		return util.WrapError(err, "read appenders section error")
	}
	if len(appenders) == 0 {
		return util.FormatError(nil, "appenders section not found")
	}

	// Read loggers
	loggers, err := s.SubKeys("logger")
	if err != nil {
		return util.WrapError(err, "read loggers section error")
	}

	// Ensure this refresh is executed only once
	if !global.init.CompareAndSwap(false, true) {
		return util.FormatError(nil, "log refresh already done")
	}

	// Factory function to create plugin instances
	newPlugin := func(typeKey string) (reflect.Value, error) {
		if !s.Has(typeKey) {
			return reflect.Value{}, util.FormatError(nil, "attribute 'type' not found")
		}
		strType := s.Get(typeKey)
		p, ok := pluginRegistry[strType]
		if !ok {
			return reflect.Value{}, util.FormatError(nil, "plugin %s not found", strType)
		}
		return NewPlugin(p.Class, typeKey[:strings.LastIndex(typeKey, ".")], s)
	}

	// Initialize appender references in a logger
	initAppenderRefs := func(v reflect.Value, cAppenders map[string]Appender) (*LoggerBase, error) {
		var base *LoggerBase
		switch config := v.Interface().(type) {
		case *SyncLogger:
			base = &config.LoggerBase
		case *AsyncLogger:
			base = &config.LoggerBase
		default: // for linter
		}
		for _, r := range base.AppenderRefs {
			appender, ok := cAppenders[r.Ref]
			if !ok {
				return nil, util.FormatError(nil, "appender %s not found", r.Ref)
			}
			r.appender = appender
		}
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
		v, err := newPlugin("appender." + name + ".type")
		if err != nil {
			return util.WrapError(err, "create appender %s error", name)
		}
		cAppenders[name] = v.Interface().(Appender)
	}

	// Initialize root logger
	if s.Has("rootLogger") {
		v, err := newPlugin("rootLogger.type")
		if err != nil {
			return util.WrapError(err, "create root logger error")
		}
		base, err := initAppenderRefs(v, cAppenders)
		if err != nil {
			return util.WrapError(err, "init appender refs for root logger error")
		}
		base.Name = RootLoggerName
		cRoot = v.Interface().(Logger)
		cLoggers[RootLoggerName] = cRoot
	}

	// Initialize all other loggers
	for _, name := range loggers {
		v, err := newPlugin("logger." + name + ".type")
		if err != nil {
			return util.WrapError(err, "create logger %s error", name)
		}
		base, err := initAppenderRefs(v, cAppenders)
		if err != nil {
			return util.WrapError(err, "init appender refs for logger %s error", name)
		}
		logger := v.Interface().(Logger)
		cLoggers[name] = logger

		// Assign tags to logger
		var tags []string
		for tag := range strings.SplitSeq(base.Tags, ",") {
			if tag = strings.TrimSpace(tag); tag == "" {
				continue
			}
			tags = append(tags, tag)
		}
		if len(tags) == 0 {
			err = util.FormatError(nil, "logger must have attribute 'tags'")
			return util.WrapError(err, "create logger %s error", name)
		}
		for _, strTag := range tags {
			if l, ok := cTags[strTag]; ok && l != logger {
				err = util.FormatError(nil, "tag '%s' already config in logger %s", strTag, l)
				return util.WrapError(err, "create logger %s error", name)
			}
			cTags[strTag] = logger
		}
	}

	// Precompile regexp patterns for tag matching
	tagRegexpMap := map[string]*regexp.Regexp{}
	for tag := range cTags {
		r, err := regexp.Compile(tag)
		if err != nil {
			return util.FormatError(err, "`%s` regexp compile error", tag)
		}
		tagRegexpMap[tag] = r
	}

	// Start all appenders
	for _, a := range cAppenders {
		if err := a.Start(); err != nil {
			return util.WrapError(err, "appender %s start error", a)
		}
	}

	// Start all loggers
	for _, l := range cLoggers {
		if err := l.Start(); err != nil {
			return util.WrapError(err, "logger %s start error", l)
		}
	}

	// Update logger references in loggerMap
	for _, l := range loggerMap {
		v, ok := cLoggers[l.name]
		if !ok {
			return util.FormatError(nil, "logger %s not found", l.name)
		}
		l.setLogger(v)
	}

	// Update tagMap with corresponding loggers
	for tag, obj := range tagRegistry {
		if l, ok := cTags[tag]; ok {
			obj.setLogger(l)
			continue
		}
		found := false
		for k, r := range tagRegexpMap {
			if r.MatchString(tag) {
				obj.setLogger(cTags[k])
				found = true
				break
			}
		}
		if found {
			continue
		}
		obj.setLogger(cRoot)
	}

	// Inject properties
	for k, f := range propertyRegistry {
		if v := s.Get(toCamelKey(k)); v == "" {
			continue
		} else if err = f(v); err != nil {
			return util.WrapError(err, "inject property %s error", k)
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

// Destroy stops all loggers and appenders and resets global state.
func Destroy() {
	if !global.init.Load() {
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
	global.init.Store(false)
}
