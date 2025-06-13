# Go-Spring :: Log

<div>
   <img src="https://img.shields.io/github/license/go-spring/log" alt="license"/>
   <img src="https://img.shields.io/github/go-mod/go-version/go-spring/log" alt="go-version"/>
   <img src="https://img.shields.io/github/v/release/go-spring/log?include_prereleases" alt="release"/>
   <img src="https://codecov.io/gh/go-spring/log/branch/main/graph/badge.svg" alt="test-coverage"/>
   <a href="https://deepwiki.com/go-spring/log"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki"></a>
</div>

[English](README.md) | [中文](README_CN.md)

**Go-Spring :: Log** is a high-performance and extensible logging library designed specifically for the Go programming
language. It offers flexible and structured logging capabilities, including context field extraction, multi-level
logging configuration, and multiple output options, making it ideal for a wide range of server-side applications.

## Features

* **Multi-Level Logging**: Supports standard log levels such as `Trace`, `Debug`, `Info`, `Warn`, `Error`, `Panic`, and
  `Fatal`, suitable for debugging and monitoring in various scenarios.
* **Structured Logging**: Records logs in a structured format with key fields like `trace_id` and `span_id`, making them
  easy to parse and analyze by log aggregation systems.
* **Context Integration**: Extracts additional information from `context.Context` (e.g., request ID, user ID) and
  automatically attaches them to log entries.
* **Tag-Based Logging**: Introduces a tag system to distinguish logs across different modules or business lines.
* **Plugin Architecture**:
    * **Appender**: Supports multiple output targets including console and file.
    * **Layout**: Provides both plain text and JSON formatting for log output.
    * **Logger**: Offers both synchronous and asynchronous loggers; asynchronous mode avoids blocking the main thread.
* **Performance Optimizations**: Utilizes buffer management and event pooling to minimize memory allocation overhead.
* **Dynamic Configuration Reload**: Supports runtime reloading of logging configurations from external files.
* **Well-Tested**: All core modules are covered with unit tests to ensure stability and reliability.

## Installation

```bash
go get github.com/go-spring/log
```

## Quick Start

Here's a simple example demonstrating how to use Go-Spring :: Log:

```go
package main

import (
	"context"

	"github.com/go-spring/log"
)

func main() {
	// Set a function to extract fields from context
	log.FieldsFromContext = func(ctx context.Context) []log.Field {
		return []log.Field{
			log.String("trace_id", "0a882193682db71edd48044db54cae88"),
			log.String("span_id", "50ef0724418c0a66"),
		}
	}

	// Load configuration file
	err := log.RefreshFile("log.xml")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Write logs
	log.Infof(ctx, log.TagDef, "This is an info message")
	log.Errorf(ctx, log.TagBiz, "This is an error message")
}
```

## Configuration

Go-Spring :: Log supports XML-based configuration for full control over logging behavior:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration>
    <Properties>
        <Property name="LayoutBufferSize">100KB</Property>
    </Properties>
    <Appenders>
        <Console name="console">
            <JSONLayout bufferSize="${LayoutBufferSize}"/>
        </Console>
    </Appenders>
    <Loggers>
        <Root level="warn">
            <AppenderRef ref="console"/>
        </Root>
        <Logger name="logger" level="info" tags="_com_request_*">
            <AppenderRef ref="console"/>
        </Logger>
    </Loggers>
</Configuration>
```

## Plugin Development

Go-Spring :: Log offers rich plugin interfaces for developers to easily implement custom `Appender`, `Layout`, and
`Logger` components.

## License

Go-Spring :: Log is licensed under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
