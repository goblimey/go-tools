// The logger package provides a simple logger.  When logging is enabled, Write writes
// to the file "./log.txt", appending to anything already there.  SetLogLevel enables 
// and disables logging.
//	
package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/goblimey/go-tools/switchwriter"
)


// If the log level is not zero, log to this file.
var logFile = "./log.txt"

type LoggerT struct {
	level  uint8
	writer *switchwriter.Writer
}

// This is a compile-time check that LoggerT implements the io.Writer interface.
var _ io.Writer = (*LoggerT)(nil)

// New creates a LoggerT object.
func New() *LoggerT {
	logger := LoggerT{0, switchwriter.New()}
	return &logger
}

// SetLogLevel sets the LoggerT's log level.  Level 0 (or negative) disables logging.
// Level 1 or greater enables logging.
//
func (logger *LoggerT) SetLogLevel(level uint8) {
	logger.level = level
	if level <= 0 {
		logger.writer.SwitchTo(nil)
	} else {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("cannot open %s for append - %v", logFile, err)
			os.Exit(1)
		}
		logger.writer.SwitchTo(f)
	}
}

// Write writes the contents of p to the logger's writer.  If the
// log level is greater than zero, that will write to the log file,
// otherwise the byte are discarded.
func (logger LoggerT) Write(p []byte) (int, error) {
	n, err := logger.writer.Write(p)
	return n, err
}
