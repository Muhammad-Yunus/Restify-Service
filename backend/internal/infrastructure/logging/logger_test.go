package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := New("info", "json", &buf)
	logger.Info(context.Background(), "hello", "key", "value")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON log output: %v", err)
	}

	if entry["msg"] != "hello" {
		t.Errorf("msg = %v, want hello", entry["msg"])
	}

	if entry["key"] != "value" {
		t.Errorf("key = %v, want value", entry["key"])
	}
}

func TestNewTextFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := New("info", "text", &buf)
	logger.Info(context.Background(), "hello")

	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("text output = %q, want it to contain hello", buf.String())
	}
}

func TestTintDefaultFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := New("debug", "default", &buf)
	logger.Debug(context.Background(), "tint message")

	if !strings.Contains(buf.String(), "tint message") {
		t.Errorf("tint output = %q, want it to contain the message", buf.String())
	}
}

func TestNewLevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	logger := New("warn", "json", &buf)
	logger.Info(context.Background(), "should be filtered")

	if buf.Len() != 0 {
		t.Errorf("info message was not filtered at warn level: %q", buf.String())
	}
}

func TestWithAddsContext(t *testing.T) {
	var buf bytes.Buffer

	logger := New("info", "json", &buf).With("service", "api")
	logger.Info(context.Background(), "hello")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON log output: %v", err)
	}

	if entry["service"] != "api" {
		t.Errorf("service = %v, want api", entry["service"])
	}
}

func TestErrorLevel(t *testing.T) {
	var buf bytes.Buffer

	logger := New("error", "json", &buf)
	logger.Error(context.Background(), "boom")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON log output: %v", err)
	}

	if entry["level"] != "ERROR" {
		t.Errorf("level = %v, want ERROR", entry["level"])
	}
}

func TestParseLevel(t *testing.T) {
	cases := map[string]string{
		"debug":   "DEBUG",
		"info":    "INFO",
		"warn":    "WARN",
		"error":   "ERROR",
		"unknown": "INFO",
	}

	for in, want := range cases {
		if got := parseLevel(in).String(); got != want {
			t.Errorf("parseLevel(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestLoggerReturnsSlogLogger(t *testing.T) {
	var buf bytes.Buffer

	logger := New("info", "json", &buf)
	if logger.Logger() == nil {
		t.Error("Logger() returned nil")
	}
}
