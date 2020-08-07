package logger

import (
	"testing"
)

var dLog LoggerT

// TestSetLevel tests the SetLevel method.
func TestSetLevel(t *testing.T) {

	dLog.SetLogLevel(1)

	if dLog.LogLevel() != 1 {
		t.Errorf("Expected logLevel %v, got %v", 1, log.LogLevel())
	}

	dLog.SetLogLevel(0)

	if dLog.LogLevel() != 0 {
		t.Errorf("Expected logLevel %v, got %v", 0, log.LogLevel())
	}
}