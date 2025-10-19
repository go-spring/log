# Go-Spring :: Log

<div>
   <img src="https://img.shields.io/github/license/go-spring/log" alt="license"/>
   <img src="https://img.shields.io/github/go-mod/go-version/go-spring/log" alt="go-version"/>
   <img src="https://img.shields.io/github/v/release/go-spring/log?include_prereleases" alt="release"/>
   <a href="https://codecov.io/gh/go-spring/log" > 
      <img src="https://codecov.io/gh/go-spring/log/graph/badge.svg?token=QBCHVEK97Q" alt="test-coverage"/> 
   </a>
   <a href="https://deepwiki.com/go-spring/log"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki"></a>
</div>

[English](README.md) | [中文](README_CN.md)

> 该项目已经正式发布，欢迎使用！

Go-Spring :: Log 是一个高性能、可扩展的日志处理库，专为 Go 语言设计。它提供了灵活的日志记录功能，支持结构化日志、
上下文字段提取、多级日志配置以及多种输出方式，适用于各种服务端应用场景。

## 特性

- **多级日志支持**：提供 Trace, Debug, Info, Warn, Error, Panic, Fatal 等日志级别，满足不同场景下的调试和监控需求。
- **结构化日志**：支持以结构化的方式记录日志字段（如 `trace_id`, `span_id`），便于日志分析系统解析。
- **上下文支持**：通过 `context.Context` 提取额外信息（如请求 ID、用户 ID）并自动附加到日志中。
- **标签系统**：使用标签（Tag）机制区分不同模块或业务线的日志。
- **插件机制**：
  - **Appender**：支持多种日志输出方式，包括控制台输出（Console）、文件写入（File）等。
  - **Layout**：提供文本格式（Text）和 JSON 格式（JSON）两种日志布局方式。
  - **Logger**：支持同步和异步日志记录器，异步记录器可防止日志写入阻塞主线程。
- **性能优化**：提供缓冲区管理、事件池化（Event Pooling）以减少内存分配开销。
- **日志刷新**：支持从配置文件动态刷新日志设置，方便运行时调整日志行为。
- **测试完备**：所有核心模块均配有单元测试，确保稳定性和可靠性。

## 核心概念

### 标签（Tag）

标签是日志包中的核心概念，用于对日志进行分类。可以通过 `RegisterTag` 函数注册标签，然后使用正则表达式匹配这些标签。
这种方法允许统一的日志 API，无需显式创建日志记录器实例，即使是第三方库也可以在不设置日志实例的情况下规范写入日志。

### 日志记录器（Logger）

`Logger` 是实际处理日志记录的对象。可以使用 `GetLogger`函数获取日志记录器实例，主要用于与旧项目的兼容性。
这允许您直接按名称检索日志记录器，并使用 `Write` 函数记录预格式化的消息。

### 上下文字段提取

可以通过可配置的函数从上下文中提取上下文数据并包含在日志条目中：

- `StringFromContext`：从上下文中提取字符串值（例如请求 ID）。
- `FieldsFromContext`：从上下文中返回结构化字段列表，如 trace ID 或用户 ID。

## 安装

```bash
go get github.com/go-spring/log
```

## 快速开始

以下是一个简单的示例，展示如何使用 Go-Spring :: Log 记录日志：

```go
package main

import (
	"context"

	"github.com/go-spring/log"
)

func main() {
	// 设置上下文字段提取函数，还可以使用 StringFromContext 函数，按需选择。
	log.FieldsFromContext = func(ctx context.Context) []log.Field {
		return []log.Field{
			log.String("trace_id", "0a882193682db71edd48044db54cae88"),
			log.String("span_id", "50ef0724418c0a66"),
		}
	}

	// 加载配置文件
	err := log.RefreshFile("log.properties")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 记录日志
	log.Infof(ctx, log.TagAppDef, "This is an info message")
	log.Errorf(ctx, log.TagBizDef, "This is an error message")

    // 使用结构化字段记录日志
    log.Info(ctx, log.TagAppDef,
        log.String("key1", "value1"),
        log.Int("key2", 123),
        log.Msg("structured log message"),
    )
}
```

## 配置说明

Go-Spring :: Log 支持通过 JSON 或 YAML 配置文件定义日志行为，例如：

```properties
bufferCap=1KB
bufferSize=1000

appender.file.type=File
appender.file.fileName=log.txt
appender.file.layout.type=JSONLayout

appender.console.type=Console
appender.console.layout.type=TextLayout

logger.root.type=Logger
logger.root.level=warn
logger.root.appenderRef.ref=console

logger.myLogger.type=AsyncLogger
logger.myLogger.level=trace
logger.myLogger.tags=_com_request_in,_com_request_*
logger.myLogger.bufferSize=${bufferSize}
logger.myLogger.appenderRef[0].ref=file
```

## 插件开发

Go-Spring :: Log 提供了丰富的插件接口，开发者可以轻松实现自定义的 Appender、Layout 和 Logger。

## 许可证

Go-Spring :: Log 使用 [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0) 许可证发布。
