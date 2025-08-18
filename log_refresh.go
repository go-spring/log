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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
)

var global struct {
	init      atomic.Bool
	loggers   []Logger
	appenders []Appender
}

const rootLoggerName = "::ROOT::"

// RefreshFile loads a logging configuration from a file by its name.
func RefreshFile(fileName string) error {

	s, err := readConfigFromFile(fileName)
	if err != nil {
		return fmt.Errorf("RefreshReader: %s", err)
	}

	if !s.Has("rootLogger") {
		return errors.New("RefreshReader: rootLogger not found")
	}

	appenders, err := s.SubKeys("appender")
	if err != nil {
		return fmt.Errorf("RefreshReader: read appenders error %w", err)
	}
	if len(appenders) == 0 {
		return errors.New("RefreshReader: Appenders section not found")
	}

	loggers, err := s.SubKeys("logger")
	if err != nil {
		return fmt.Errorf("RefreshReader: read loggers error %w", err)
	}
	if len(loggers) == 0 {
		return errors.New("RefreshReader: Loggers section not found")
	}

	if !global.init.CompareAndSwap(false, true) {
		return errors.New("RefreshReader: log refresh already done")
	}

	newPlugin := func(typeKey string) (reflect.Value, error) {
		if !s.Has(typeKey) {
			return reflect.Value{}, fmt.Errorf("RefreshReader: attribute 'type' not found")
		}
		strType := s.Get(typeKey)
		p, ok := plugins[strType]
		if !ok {
			return reflect.Value{}, fmt.Errorf("RefreshReader: plugin %s not found", strType)
		}
		return NewPlugin(p.Class, typeKey[:strings.LastIndex(typeKey, ".")], s)
	}

	initAppenderRefs := func(v reflect.Value, cAppenders map[string]Appender) (*BaseLogger, error) {
		var base *BaseLogger
		switch config := v.Interface().(type) {
		case *SyncLogger:
			base = &config.BaseLogger
		case *AsyncLogger:
			base = &config.BaseLogger
		default: // for linter
		}
		for _, r := range base.AppenderRefs {
			appender, ok := cAppenders[r.Ref]
			if !ok {
				return nil, fmt.Errorf("RefreshReader: appender %s not found", r.Ref)
			}
			r.appender = appender
		}
		return base, nil
	}

	var (
		cRoot      Logger
		cLoggers   = make(map[string]Logger)
		cAppenders = make(map[string]Appender)
		cTags      = make(map[string]Logger)
	)

	for _, name := range appenders {
		v, err := newPlugin("appender." + name + ".type")
		if err != nil {
			return err
		}
		cAppenders[name] = v.Interface().(Appender)
	}

	{
		v, err := newPlugin("rootLogger.type")
		if err != nil {
			return err
		}
		base, err := initAppenderRefs(v, cAppenders)
		if err != nil {
			return err
		}
		base.Name = rootLoggerName
		cRoot = v.Interface().(Logger)
		cLoggers[rootLoggerName] = cRoot
	}

	for _, name := range loggers {
		v, err := newPlugin("logger." + name + ".type")
		if err != nil {
			return err
		}
		base, err := initAppenderRefs(v, cAppenders)
		if err != nil {
			return err
		}
		logger := v.Interface().(Logger)
		cLoggers[name] = logger

		var tags []string
		for tag := range strings.SplitSeq(base.Tags, ",") {
			if tag = strings.TrimSpace(tag); tag == "" {
				continue
			}
			tags = append(tags, tag)
		}
		if len(tags) == 0 {
			return fmt.Errorf("RefreshReader: logger must have attribute 'tags'")
		}
		for _, strTag := range tags {
			cTags[strTag] = logger
		}
	}

	var (
		tagArray    []string
		tagExpArray []*regexp.Regexp
		loggerArray []Logger
	)

	for _, s := range OrderedMapKeys(cTags) {
		r, err := regexp.Compile(s)
		if err != nil {
			return WrapError(err, "RefreshReader: `%s` regexp compile error", s)
		}
		tagArray = append(tagArray, s)
		tagExpArray = append(tagExpArray, r)
		loggerArray = append(loggerArray, cTags[s])
	}

	// TODO(lvan100): Currently, there is only one refresh operation,
	// so exception handling is temporarily ignored.

	for _, a := range cAppenders {
		if err := a.Start(); err != nil {
			return WrapError(err, "RefreshReader: appender %s start error", a.GetName())
		}
	}
	for _, l := range cLoggers {
		if err := l.Start(); err != nil {
			return WrapError(err, "RefreshReader: logger %s start error", l.GetName())
		}
	}

	for _, l := range loggerMap {
		v, ok := cLoggers[l.name]
		if !ok {
			return fmt.Errorf("RefreshReader: logger %s not found", l.name)
		}
		l.setLogger(v)
	}

	for tag, obj := range tagMap {
		logger := cRoot
		for i := range len(tagArray) {
			s, r := tagArray[i], tagExpArray[i]
			if s == tag || r.MatchString(tag) {
				logger = loggerArray[i]
				break
			}
		}
		obj.setLogger(logger)
	}

	for k, f := range propertyMap {
		const defVal = "::def::"
		v := s.Get(ToCamelKey(k), defVal)
		if v == defVal {
			return fmt.Errorf("RefreshReader: property %s not found", k)
		}
		if err = f(v); err != nil {
			return WrapError(err, "RefreshReader: inject property %s error", k)
		}
	}

	for _, l := range cLoggers {
		global.loggers = append(global.loggers, l)
	}

	for _, a := range cAppenders {
		global.appenders = append(global.appenders, a)
	}

	return nil
}

// Destroy destroys all loggers.
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
	global.init.Store(false)
	global.appenders = nil
	global.loggers = nil
}
