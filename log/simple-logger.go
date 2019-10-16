package log

import (
	"fmt"
	stdlog "log"
	"os"
)

type simple struct {
	fields map[string]interface{}
}

// NewSimple creates a basic logger that wraps the core log library.
func NewSimple() Logger {
	return &simple{}
}

// WithFields will return a new logger based on the original logger
// with the additional supplied fields
func (b *simple) WithFields(fields Fields) Logger {
	cp := &simple{}

	if b.fields == nil {
		cp.fields = fields
		return cp
	}

	cp.fields = make(map[string]interface{}, len(b.fields)+len(fields))
	for k, v := range b.fields {
		cp.fields[k] = v
	}

	for k, v := range fields {
		cp.fields[k] = v
	}

	return cp
}

// Debug log message
func (b *simple) Debug(msg ...interface{}) {
	stdlog.Printf("[DEBUG] %s %s", fmt.Sprint(msg...), pretty(b.fields))
}

// Info log message
func (b *simple) Info(msg ...interface{}) {
	stdlog.Printf("[INFO] %s %s", fmt.Sprint(msg...), pretty(b.fields))
}

// Warn log message
func (b *simple) Warn(msg ...interface{}) {
	stdlog.Printf("[WARN] %s %s", fmt.Sprint(msg...), pretty(b.fields))
}

// Error log message
func (b *simple) Error(msg ...interface{}) {
	stdlog.Printf("[ERROR] %s %s", fmt.Sprint(msg...), pretty(b.fields))
}

// Fatal log message (and exit)
func (b *simple) Fatal(msg ...interface{}) {
	stdlog.Printf("[FATAL] %s %s", fmt.Sprint(msg...), pretty(b.fields))
	os.Exit(1)
}

// helper for pretty printing of fields
func pretty(m map[string]interface{}) string {
	if len(m) < 1 {
		return ""
	}

	s := ""
	for k, v := range m {
		s += fmt.Sprintf("%s=%v ", k, v)
	}

	return s[:len(s)-1]
}
