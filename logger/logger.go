package logger

import (
	"os"
	"regexp"
)

// LoggerT is the log data.
type LoggerT struct {
	level  uint8
	stream *os.File
}

var log LoggerT

const logLevelRequest = "/loglevel/"

var logLevelRequestRE = regexp.MustCompile(`^/loglevel/([0-9]+)$`)

// MakeLog creates and returns a log.
func MakeLogger() LoggerT {
	log := LoggerT{stream: os.Stderr}
	return log
}

// Log gets the log
func (l LoggerT) Log() LoggerT {
	return log
}

// Write to a logging stream.
func (l LoggerT) Write(buf []byte) (n int, err error) {
	if l.level == 0 {
		return
	}
	return l.stream.Write(buf)
}

// LogLevel returns the logging level.
func (l LoggerT) LogLevel() uint8 {
	return l.level
}

// SetLogLevel sets the log level.  0 is er
func (l *LoggerT) SetLogLevel(n uint8) {
	l.level = n
}