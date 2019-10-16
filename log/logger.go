package log

type Logger interface {
	Debug(msg ...interface{})
	Info(msg ...interface{})
	Warn(msg ...interface{})
	Error(msg ...interface{})
	WithFields(Fields) Logger
}

type Fields map[string]interface{}
