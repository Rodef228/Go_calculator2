package logger

// MockLogger implements Logger interface for testing
type MockLogger struct {
	Messages []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, "DEBUG: "+msg)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, "INFO: "+msg)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, "WARN: "+msg)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, "ERROR: "+msg)
}

func (m *MockLogger) Infow(msg string, keysAndValues ...interface{}) {
	m.Messages = append(m.Messages, "INFOW: "+msg)
}
