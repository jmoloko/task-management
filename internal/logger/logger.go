package logger

// Logger определяет интерфейс для операций логирования
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	WithFields(fields map[string]interface{}) Logger
	Close() error
}
