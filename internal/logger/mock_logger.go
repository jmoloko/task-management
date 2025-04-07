package logger

// MockLogger реализует интерфейс Logger для тестов
type MockLogger struct{}

func (l *MockLogger) Debug(msg string, args ...interface{}) {}
func (l *MockLogger) Info(msg string, args ...interface{})  {}
func (l *MockLogger) Warn(msg string, args ...interface{})  {}
func (l *MockLogger) Error(msg string, args ...interface{}) {}
func (l *MockLogger) Fatal(msg string, args ...interface{}) {}
func (l *MockLogger) WithFields(fields map[string]interface{}) Logger {
	return l
}
func (l *MockLogger) Close() error {
	return nil
}
