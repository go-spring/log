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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-spring/spring-base/util"
)

// Stdout is the standard output stream used by appenders.
var Stdout io.Writer = os.Stdout

func init() {
	RegisterPlugin[DiscardAppender]("Discard", PluginTypeAppender)
	RegisterPlugin[ConsoleAppender]("Console", PluginTypeAppender)
	RegisterPlugin[FileAppender]("File", PluginTypeAppender)
	RegisterPlugin[RollingFileAppender]("RollingFile", PluginTypeAppender)
}

// Appender defines components that handle log output.
// 要求所有的 appender 实现都必须是并发安全的。
type Appender interface {
	Lifecycle        // Start/Stop methods for resource management
	GetName() string // Returns the appender's name
	Append(e *Event) // Handles writing a log event
	Write(b []byte)  // Directly writes a byte slice
}

// AppenderBase provides common configuration fields for all appenders.
type AppenderBase struct {
	Name string `PluginAttribute:"name"`
}

// GetName returns the appender's name.
func (c *AppenderBase) GetName() string { return c.Name }
func (c *AppenderBase) Start() error    { return nil }
func (c *AppenderBase) Stop()           {}

var (
	_ Appender = (*DiscardAppender)(nil)
	_ Appender = (*ConsoleAppender)(nil)
	_ Appender = (*FileAppender)(nil)
	_ Appender = (*RollingFileAppender)(nil)
)

// DiscardAppender ignores all log events (no-op).
type DiscardAppender struct {
	AppenderBase
}

func (c *DiscardAppender) Append(e *Event) {}
func (c *DiscardAppender) Write(b []byte)  {}

// ConsoleAppender writes formatted log events to standard output.
type ConsoleAppender struct {
	AppenderBase
	Layout Layout `PluginElement:"Layout,default=TextLayout"`
}

// Append formats the event and writes it to standard output.
func (c *ConsoleAppender) Append(e *Event) {
	c.Write(c.Layout.ToBytes(e))
}

// Write writes a byte slice directly to standard output.
func (c *ConsoleAppender) Write(b []byte) {
	_, _ = Stdout.Write(b)
}

// FileAppender writes formatted log events to a specified file.
type FileAppender struct {
	AppenderBase
	Layout   Layout `PluginElement:"Layout,default=TextLayout"`
	FileDir  string `PluginAttribute:"fileDir,default=./logs"`
	FileName string `PluginAttribute:"fileName"`

	file *os.File
}

// Start opens the log file for appending.
func (c *FileAppender) Start() error {
	const fileFlag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	fileName := filepath.Join(c.FileDir, c.FileName)
	f, err := os.OpenFile(fileName, fileFlag, 0644)
	if err != nil {
		return err
	}
	c.file = f
	return nil
}

// Append formats the log event and writes it to the file.
func (c *FileAppender) Append(e *Event) {
	c.Write(c.Layout.ToBytes(e))
}

// Write writes a byte slice directly to the file.
func (c *FileAppender) Write(b []byte) {
	_, _ = c.file.Write(b)
}

// Stop flushes and closes the file.
func (c *FileAppender) Stop() {
	if c.file != nil {
		_ = c.file.Sync()
		_ = c.file.Close()
	}
}

type RollingPolicy struct {
	Interval time.Duration
}

func (r RollingPolicy) Time(t time.Time) int64 {
	seconds := int64(r.Interval.Seconds())
	return (t.Unix() / seconds) * seconds
}

// Format formats the time into a string with the pattern "yyyyMMddHHmmss".
func (r RollingPolicy) Format(t time.Time) string {
	return t.Format("20060102150405")
}

// RollingFileAppender allows **multiple goroutines** to call Write()
// safely, at the cost of slightly higher overhead and potential
// (acceptable) log loss during rotation.不支持按文件大小轮转。
//
// Usage scenarios:
//   - High-concurrency applications where logs may be produced
//     from many goroutines.
//
// Risks:
//   - During rotation, a small number of writes may fail if they
//     occur after the old file is closed but before the new file is ready.
//   - During Stop(), concurrent writes may also be lost.
//   - If zero log loss is required, use AsyncRotateFileWriter
//     with a dedicated logging goroutine instead.
type RollingFileAppender struct {
	AppenderBase
	Layout        Layout `PluginElement:"Layout,default=TextLayout"`
	FileDir       string `PluginAttribute:"fileDir,default=./logs"`
	FileName      string `PluginAttribute:"fileName"`
	ClearHours    int32
	RollingPolicy RollingPolicy

	file     atomic.Pointer[os.File]
	oldFile  atomic.Pointer[os.File]
	currTime atomic.Int64
}

// Start opens the initial log file.
func (c *RollingFileAppender) Start() error {
	now := time.Now()
	nowTime := c.RollingPolicy.Time(now)
	filePath, file, err := c.createFile(c.RollingPolicy.Format(now))
	if err != nil {
		return util.WrapError(err, "Failed to create log file %s", filePath)
	}
	c.file.Store(file)
	c.currTime.Store(nowTime)
	return nil
}

// Append formats the log event and writes it to the current file.
func (c *RollingFileAppender) Append(e *Event) {
	c.Write(c.Layout.ToBytes(e))
}

// Write writes bytes to the current log file.
// May lose a few writes during rotation or Stop().
// 由于高并发的原因，并没有在日志文件轮转时加锁等待，所以会出现新日志写到旧文件
// 情况，这种是符合预期的。尤其在整点轮转遭遇整点任务时，往往是出现高并发的场景。
func (c *RollingFileAppender) Write(b []byte) {
	c.rotate()
	if file := c.file.Load(); file != nil {
		_, _ = file.Write(b)
	}
}

// Stop flushes and closes the current file.
// 如果执行 stop 时，仍然有新日志写入，那么这些日志往往会丢失。
func (c *RollingFileAppender) Stop() {
	// 关闭上一个周期的文件
	if file := c.oldFile.Swap(nil); file != nil {
		_ = file.Sync()
		_ = file.Close()
	}
	// 关闭当前周期的文件
	if file := c.file.Swap(nil); file != nil {
		_ = file.Sync()
		_ = file.Close()
	}
}

// rotate checks if the current time has passed into a new rotation slot.
// If so, it closes the old file, opens a new one, and triggers cleanup.
// Risk: If file creation fails during rotation, new logs will be lost
// until the issue is resolved.
func (c *RollingFileAppender) rotate() {
	now := time.Now()
	nowTime := c.RollingPolicy.Time(now)
	oldTime := c.currTime.Load()
	// 说明正在进行或者已经完成轮转过程
	if nowTime <= oldTime {
		return
	}
	// 只能有一个并发抢到更新文件句柄的机会
	if !c.currTime.CompareAndSwap(oldTime, nowTime) {
		return
	}

	// 关闭上一个周期的文件
	if file := c.oldFile.Swap(nil); file != nil {
		_ = file.Sync()
		_ = file.Close()
	}

	// 可能有 rename 过程

	// 创建新文件，创建失败的时候尽量保持原状，等待下一个周期的轮转
	filePath, file, err := c.createFile(c.RollingPolicy.Format(now))
	if err != nil {
		err = util.WrapError(err, "Failed to create log file %s", filePath)
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}

	// 保留上一个周期的日志文件，因为还可能有并发写在进行
	oldFile := c.file.Load()
	c.oldFile.Store(oldFile)

	// 更新文件句柄
	c.file.Store(file)
	c.currTime.Store(nowTime)

	// trigger cleanup after each rotation for timely housekeeping
	go c.clearExpiredFiles()
}

// createFile creates or opens the current log file for appending.
// The application is responsible for ensuring the directory exists.
func (c *RollingFileAppender) createFile(formatTime string) (string, *os.File, error) {
	fileName := c.FileName + "." + formatTime
	filePath := filepath.Join(c.FileDir, fileName)
	const fileFlag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	file, err := os.OpenFile(filePath, fileFlag, 0644)
	if err != nil {
		return filePath, nil, err
	}
	return filePath, file, nil
}

// clearExpiredFiles removes expired log files.
func (c *RollingFileAppender) clearExpiredFiles() {
	expiration := time.Now().Add(-time.Duration(c.ClearHours) * time.Hour)
	entries, _ := os.ReadDir(c.FileDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), c.FileName+".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(expiration) {
			filePath := fmt.Sprintf("%s/%s", c.FileDir, entry.Name())
			_ = os.Remove(filePath)
		}
	}
}
