// This program is a code sketch of how to use package switchwriter in a simplistic logging system.
// It simulates getting some data, logging it, and rotating the log files.

package logger

import (
	"fmt"
	"os"
	"time"

	"example.org/switchwriter"
)

var logDir = "logs/"
var logFilePrefix = "log-number-"
var logPeriod = time.Millisecond * 1000

var logFile = switchwriter.New()

// Main is just here for my module testing experiments.
//
func Main() {

	logSetup()
	go logData()
	go rotateLogs()

	// Only give enough time for 7 log files to be created.
	time.Sleep(logPeriod * 7)
}

// Make sure the log directory exists and is empty
//
func logSetup() {
	err := os.RemoveAll(logDir)
	check(err)

	err = os.Mkdir(logDir, 0777)
	check(err)
}

// logData() simulates getting message data (just integer values)
// from a slow device and writing them to the logger.
//
func logData() {
	for msg := 1; true; msg++ {
		fmt.Fprintf(logFile, "%d\n", msg)

		time.Sleep(logPeriod / 100)
	}
}

// rotateLogs() creates and rotates up to 10 log files.
//
func rotateLogs() {

	oldLog := (*os.File)(nil)
	for i := 1; i <= 10; i++ {

		// Switch to a new log.
		newLog, t := newLogFile(i)
		logFile.SwitchTo(newLog)

		// Mop up old log file.
		closeLog(oldLog) // A NOP first time round loop, as oldFile is nil.
		oldLog = newLog

		// Use new log file.
		time.Sleep(t)
	}
	closeLog(oldLog) // The last log file will not yet have been closed.
}

// newLogFile returns a newly created logfile and
// a duration for which the logFile is to be used.
//
func newLogFile(i int) (f *os.File, t time.Duration) {
	logName := fmt.Sprintf("%s/%s%d", logDir, logFilePrefix, i)
	f, err := os.Create(logName)
	check(err)
	return f, logPeriod
}

func closeLog(f *os.File) {
	if f != nil {
		f.Close()
	}
}

// Simplistic error handling.
//
func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
