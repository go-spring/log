package expr

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "simple type with no fields",
			input: "Logger {}",
			want: map[string]string{
				"type": "Logger",
			},
		},
		{
			name:  "type with string field",
			input: `Logger { level = "info" }`,
			want: map[string]string{
				"type":  "Logger",
				"level": "info",
			},
		},
		{
			name:  "type with raw value field",
			input: "Logger { level = info }",
			want: map[string]string{
				"type":  "Logger",
				"level": "info",
			},
		},
		{
			name:  "type with multiple fields",
			input: `Logger { level = "info", output = "stdout" }`,
			want: map[string]string{
				"type":   "Logger",
				"level":  "info",
				"output": "stdout",
			},
		},
		{
			name:  "type with nested expression",
			input: `Logger { level = "info", file = FileAppender { path = "/tmp/app.log" } }`,
			want: map[string]string{
				"type":      "Logger",
				"level":     "info",
				"file.type": "FileAppender",
				"file.path": "/tmp/app.log",
			},
		},
		{
			name:  "complex nested structure",
			input: `Logger { level = "debug", file = RollingFileAppender { path = "/tmp/app.log", policy = SizeBasedTriggeringPolicy { maxFileSize = "10MB" } } }`,
			want: map[string]string{
				"type":                    "Logger",
				"level":                   "debug",
				"file.type":               "RollingFileAppender",
				"file.path":               "/tmp/app.log",
				"file.policy.type":        "SizeBasedTriggeringPolicy",
				"file.policy.maxFileSize": "10MB",
			},
		},
		{
			name:    "invalid syntax missing closing brace",
			input:   `Logger { level = "info" `,
			wantErr: true,
		},
		{
			name:    "invalid syntax missing equals",
			input:   `Logger { level "info" }`,
			wantErr: true,
		},
		{
			name:    "invalid syntax missing opening brace",
			input:   `Logger level = "info" }`,
			wantErr: true,
		},
		{
			name:  "fields with special characters in strings",
			input: `Logger { format = "time=\"${timestamp}\" level=${level}" }`,
			want: map[string]string{
				"type":   "Logger",
				"format": "time=\"${timestamp}\" level=${level}",
			},
		},
		{
			name:  "whitespace handling",
			input: `  Logger  {  level  =  "info"  }  `,
			want: map[string]string{
				"type":  "Logger",
				"level": "info",
			},
		},
		{
			name:  "field with array index access",
			input: `Logger { appender[0] = "stdout" }`,
			want: map[string]string{
				"type":        "Logger",
				"appender[0]": "stdout",
			},
		},
		{
			name:  "field with dot notation access",
			input: `Logger { appender.out = "stdout" }`,
			want: map[string]string{
				"type":         "Logger",
				"appender.out": "stdout",
			},
		},
		{
			name:  "field with complex access",
			input: `Logger { appender.out[0].name = "stdout" }`,
			want: map[string]string{
				"type":                 "Logger",
				"appender.out[0].name": "stdout",
			},
		},
		{
			name:  "single quoted string",
			input: `Logger { level = 'info' }`,
			want: map[string]string{
				"type":  "Logger",
				"level": "info",
			},
		},
		{
			name:  "string with escaped characters",
			input: `Logger { format = "time=\"${timestamp}\"\nlevel=${level}" }`,
			want: map[string]string{
				"type":   "Logger",
				"format": "time=\"${timestamp}\"\nlevel=${level}",
			},
		},
		{
			name:  "trailing comma in field list",
			input: `Logger { level = "info", output = "stdout", }`,
			want: map[string]string{
				"type":   "Logger",
				"level":  "info",
				"output": "stdout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
