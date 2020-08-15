package dailylogger

import (
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	ts "github.com/goblimey/go-tools/testsupport"
)

// TestDailyLoggerIntegration is an integration test of the daily logger.
//
func TestDailyLoggerIntegration(t *testing.T) {

	// This test uses the filestore.  It creates a directory in /tmp containing
	// a plain file.  At the end it attempts to remove everything it created.
	//
	// It creates a production version of the daily logger.  It's expected to
	// produce one log file but it will start the log rollover goroutine.  If
	// it's run just before midnight that could result in two log files, which
	// would make the test fail.

	directoryName, err := ts.CreateWorkingDirectory()
	if err != nil {
		t.Fatalf("createWorkingDirectory failed - %v", err)
	}
	defer ts.RemoveWorkingDirectory(directoryName)

	writer := New(".", "", "")

	expectedFilenamePattern := "daily.[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9].log"

	expectedMessage := "hello world"
	buffer := []byte(expectedMessage)

	n, err := writer.Write(buffer)

	if err != nil {
		t.Fatalf("Write failed - %v", err)
	}

	if n != len(buffer) {
		t.Fatalf("Write returned %d - expected %d", n, len(buffer))
	}

	// Check that one log file was created and contains the expected contents.
	files, err := ioutil.ReadDir(directoryName)
	if err != nil {
		t.Fatalf("error scanning directory %s - %s", directoryName, err.Error())
	}

	if len(files) != 1 {
		t.Fatalf("directory %s contains %d files.  Should contain exactly one.", directoryName, len(files))
	}

	match, err := regexp.MatchString(expectedFilenamePattern, files[0].Name())
	if err != nil {
		t.Fatalf("error matching log file name %s - %s", files[0].Name(), err.Error())
	}
	if !match {
		t.Fatalf("directory %s contains file \"%s\", incorrect name format",
			directoryName, files[0].Name())
	}

	// Check the contents.
	inputFile, err := os.OpenFile(files[0].Name(), os.O_RDONLY, 0644)
	defer inputFile.Close()
	b := make([]byte, 8096)
	length, err := inputFile.Read(b)
	if err != nil {
		t.Fatalf("error reading logfile back - %v", err)
	}
	if length != len(buffer) {
		t.Fatalf("logfile %s contains %d bytes - expected %d",
			files[0].Name(), length, len(buffer))
	}

	contents := string(b[:length])

	if expectedMessage != contents {
		t.Fatalf("logfile %s contains \"%s\" - expected \"%s\"",
			files[0].Name(), contents, expectedMessage)
	}
}
