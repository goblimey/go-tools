// This program is a code sketch of how to use SwitchWriter in a simplistic logging system.
//

package main

import (
	"fmt"
	"os"
	"time"
	"github.com/goblimey/go-tools/switchWriter"
)

// main() simulates getting some data, logging it, and rotating the log files.
//
func main() {

	logSetup()
	go logData()
	go rotateLogs()

	// Only give enough time for 7 log files to be created.
	time.Sleep(7 * time.Second)
}

// Log Plumbing
//
var logDir = "logs/"
var logFilePrefix = "log-number-"
var logFile = switchWriter.NewWriter()

// Make sure the log directory exists and is empty
//
func logSetup() {
	err := os.RemoveAll(logDir)
	check(err)

	err = os.Mkdir(logDir, 0777)
	check(err)
}

// logData() simluates getting message data from the device
// and writing it to the logger.
//
// A simulated message is generated every 10 milliseconds.
// We generate simple integer values as the simulated messages.
//
func logData() {
	for msg := 1; true; msg++ {
		fmt.Fprintf(logFile, "%d\n", msg)

		time.Sleep(10 * time.Millisecond)
	}
}

// rotateLogs() creates and rotates up to 10 log files.
// Each logfile will log for one second.
//
func rotateLogs() {
	for i := 1; i < 10; i++ {

		// Create new logfile.
		logName := fmt.Sprintf("%s/%s%d", logDir, logFilePrefix, i)
		f, err := os.Create(logName)
		check(err)

		// Switch to new logfile, and log for one second.
		logFile.SwitchTo(f)
		time.Sleep(1000 * time.Millisecond)
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
