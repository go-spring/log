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

//type Action interface {
//	Execute() error
//}
//
//type RenameAction struct {
//}
//
//type CompressAction struct {
//}
//
//type RollingPolicy interface {
//	Rollover(m RollingFileManager)
//}
//
//type SizeBasedRollingPolicy struct {
//	MaxFileSize int64
//}
//
//func (r *SizeBasedRollingPolicy) Rollover(m RollingFileManager) {
//	if m.Size() >= r.MaxFileSize {
//		m.Rollover()
//	}
//}
//
//type TimeBasedRollingPolicy struct {
//	Interval int64
//}
//
//func (r *TimeBasedRollingPolicy) Rollover(m RollingFileManager) {
//	nowTime := (time.Now().Unix() / r.Interval) * r.Interval
//	if m.NextTime() <= nowTime {
//		m.Rollover()
//	}
//}
//
//type SizeAndTimeBasedRollingPolicy struct {
//	SizeBasedRollingPolicy
//	TimeBasedRollingPolicy
//}
//
//func (r *SizeAndTimeBasedRollingPolicy) Rollover(m RollingFileManager) {
//	nowTime := (time.Now().Unix() / r.Interval) * r.Interval
//	if m.NextTime() <= nowTime || m.Size() >= r.MaxFileSize {
//		m.Rollover()
//	}
//}
//
//type RollingFileManager interface {
//	NextTime() int64
//	Size() int64
//	Rollover()
//	Write(p []byte) (n int, err error)
//	Close()
//}
//
////type RollingFileManager struct {
////	size int64
////	file *os.File
////}
////
////func OpenRollingFile(fileName string) (*RollingFileManager, error) {
////	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
////	if err != nil {
////		return nil, err
////	}
////	stat, err := file.Stat()
////	if err != nil {
////		return nil, err
////	}
////	return &RollingFileManager{
////		size: stat.Size(),
////		file: file,
////	}, nil
////}
////
////func (f *RollingFileManager) rollover() {
////
////}
////
////func (f *RollingFileManager) Write(p []byte) (n int, err error) {
////	return f.file.Write(p)
////}
////
////func (f *RollingFileManager) Close() {
////	f.file.Sync()
////	f.file.Close()
////}
