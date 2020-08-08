package dailylogger

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/goblimey/go-tools/clock"
)

// TestGetDurationToMidnight tests the getDurationToMidnight method.
//
func TestGetDurationToMidnight(t *testing.T) {
	locationUTC, _ := time.LoadLocation("UTC")
	start := time.Date(2020, time.February, 14, 22, 59, 0, 0, locationUTC)
	expectedDurationNanoseconds := time.Hour + time.Minute
	duration := getDurationToMidnight(start)
	if duration.Nanoseconds() != int64(expectedDurationNanoseconds) {
		t.Fatalf("expected duration to be \"%d\" actually \"%d\"", expectedDurationNanoseconds, duration.Nanoseconds())
	}

	start = time.Date(2020, time.February, 14, 0, 30, 3, 4, locationUTC)
	expectedDurationNanoseconds = 23*time.Hour + 30*time.Minute - (3*time.Second + 4*time.Nanosecond)
	duration = getDurationToMidnight(start)
	if duration.Nanoseconds() != int64(expectedDurationNanoseconds) {
		t.Fatalf("expected duration to be \"%d\" actually \"%d\"", expectedDurationNanoseconds, duration.Nanoseconds())
	}

	// Test using a time that's not in UTC.
	locationParis, _ := time.LoadLocation("Europe/Paris")
	start = time.Date(2020, time.February, 14, 13, 5, 0, 0, locationParis)
	expectedDurationNanoseconds = 10*time.Hour + 55*time.Minute
	duration = getDurationToMidnight(start)
	if duration.Nanoseconds() != int64(expectedDurationNanoseconds) {
		t.Fatalf("expected duration to be \"%d\" actually \"%d\"", expectedDurationNanoseconds, duration.Nanoseconds())
	}
}

// TestLogging tests that logging works - creates a file of the right name with the
// right contents.
//
func TestLogging(t *testing.T) {

	// This test uses the filestore.  It creates a directory in /tmp containing
	// a plain file.  At the end it attempts to remove everything it created.

	directoryName, err := createWorkingDirectory()
	if err != nil {
		t.Fatalf("createWorkingDirectory failed - %v", err)
	}
	defer removeWorkingDirectory(directoryName)

	locationParis, _ := time.LoadLocation("Europe/Paris")
	/*
		days := []time.Time{
			time.Date(2020, time.February, 14, 1, 2, 3, 4, locationParis),
			time.Date(2020, time.February, 14, 2, 2, 3, 4, locationParis)}
	*/
	stoppedClock :=
		clock.NewStoppedClock(2020, time.February, 14, 1, 2, 3, 4, locationParis)

	writer := newDailyWriterWithClock(stoppedClock, ".", "", "")
	expectedFilename := "daily.2020-02-14.log"
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

	if files[0].Name() != expectedFilename {
		t.Fatalf("directory %s contains file \"%s\", expected \"%s\".", directoryName, files[0].Name(), expectedFilename)
	}

	// Check the contents.
	inputFile, err := os.OpenFile(expectedFilename, os.O_RDONLY, 0644)
	defer inputFile.Close()
	b := make([]byte, 8096)
	length, err := inputFile.Read(b)
	if err != nil {
		t.Fatalf("error reading logfile back - %v", err)
	}
	if length != len(buffer) {
		t.Fatalf("logfile contains %d bytes - expected %d", length, len(buffer))
	}

	contents := string(b[:length])

	if expectedMessage != contents {
		t.Fatalf("logfile contains \"%s\" - expected \"%s\"", contents, expectedMessage)
	}
}

// TestRollover checks that the log rollover mecahanism creates a new file each day.
func TestRollover(t *testing.T) {

	// This test uses the filestore.

	const expectedMessage1 = "hello"
	const expectedFilename1 = "foo.2020-02-14.bar"
	buffer1 := []byte(expectedMessage1)
	const expectedMessage2 = "world"
	buffer2 := []byte(expectedMessage2)
	const expectedFilename2 = "foo.2020-02-15.bar"

	directoryName, err := createWorkingDirectory()
	if err != nil {
		t.Fatalf("createWorkingDirectory failed - %v", err)
	}
	defer removeWorkingDirectory(directoryName)

	locationParis, _ := time.LoadLocation("Europe/Paris")
	times := []time.Time{
		// 200 milliSeconds before midnight.
		time.Date(2020, time.February, 14, 23, 59, 59, int(time.Millisecond)*800, locationParis),
		// 00:01am The next day.
		time.Date(2020, time.February, 15, 0, 0, 0, 0, locationParis)}
	steppingClock := clock.NewSteppingClock(&times)

	writer := newDailyWriterWithClock(steppingClock, ".", "foo.", ".bar")

	// This should write to expectedFilename1.
	n, err := writer.Write(buffer1)
	if err != nil {
		t.Fatalf("Write failed - %v", err)
	}

	if n != len(buffer1) {
		t.Fatalf("Write returns %d - expected %d", n, len(buffer1))
	}

	// roll the log over.
	writer.rotateLogs()

	// This should write to expectedFilename2.
	n, err = writer.Write(buffer2)
	if err != nil {
		t.Fatalf("Write failed - %v", err)
	}

	if n != len(buffer2) {
		t.Fatalf("Write returns %d - expected %d", n, len(buffer2))
	}

	// The current directory should contain expectedLogfile1 and expectedLogfile2.
	// Check that tthe two files exist.
	files, err := ioutil.ReadDir(directoryName)
	if err != nil {
		t.Fatalf("error scanning directory %s - %s", directoryName, err.Error())
	}

	if len(files) != 2 {
		t.Fatalf("directory %s contains %d files.  Should contain just 2.",
			directoryName, len(files))
	}

	if files[0].Name() != expectedFilename1 &&
		files[0].Name() != expectedFilename2 {

		t.Fatalf("directory %s contains file \"%s\", expected \"%s\" or \"%s\".",
			directoryName, files[0].Name(), expectedFilename1, expectedFilename2)
	}

	if files[1].Name() != expectedFilename1 &&
		files[1].Name() != expectedFilename2 {

		t.Fatalf("directory %s contains file \"%s\", expected \"%s\" or \"%s\".",
			directoryName, files[1].Name(), expectedFilename1, expectedFilename2)
	}

	// Check the contents.
	expectedPathName := directoryName + "/" + expectedFilename1
	inputFile, err := os.OpenFile(expectedPathName, os.O_RDONLY, 0644)
	defer inputFile.Close()
	b := make([]byte, 8096)
	length, err := inputFile.Read(b)
	if err != nil {
		t.Fatalf("error reading logfile %s back - %v", expectedFilename1, err)
	}
	if length != len(expectedMessage1) {
		t.Fatalf("logfile contains %d bytes - expected %d", length, len(expectedMessage1))
	}
	result1 := string(b[:length])
	if result1 != expectedMessage1 {
		t.Fatalf("logfile contains \"%s\" - expected \"%s\"", result1, expectedMessage1)
	}

	expectedPathName = directoryName + "/" + expectedFilename2
	inputFile, err = os.OpenFile(expectedPathName, os.O_RDONLY, 0644)
	defer inputFile.Close()
	length, err = inputFile.Read(b)
	if err != nil {
		t.Fatalf("error reading logfile %s back - %v", expectedFilename2, err)
	}
	if length != len(buffer2) {
		t.Fatalf("logfile contains %d bytes - expected %d", length, len(buffer2))
	}
	result2 := string(b[:length])
	if result2 != expectedMessage2 {
		t.Fatalf("logfile contains \"%s\" - expected \"%s\"", result2, expectedMessage2)
	}
}

// TestRolloverWithLongDelay checks that the log rollover mechanism produces
// the correct datestamp even when works it's run very late and the day has
// moved on further.
func TestRolloverWithLongDelay(t *testing.T) {

	// This test uses the filestore.

	const message1 = "hello"
	buffer1 := []byte(message1)
	const expectedMessage = "world"
	buffer2 := []byte(expectedMessage)
	const expectedFilename = "foo.2020-02-16.bar"

	directoryName, err := createWorkingDirectory()
	if err != nil {
		t.Fatalf("createWorkingDirectory failed - %v", err)
	}
	defer removeWorkingDirectory(directoryName)

	locationLondon, _ := time.LoadLocation("Europe/London")
	times := []time.Time{
		// 200 milliSeconds before midnight.
		time.Date(2020, time.February, 14, 23, 59, 59, int(time.Millisecond)*800,
			locationLondon),
		// Simulate the effect of a very long delay running the log
		// rotation, talikng us into the day after next.
		time.Date(2020, time.February, 16, 0, 0, 0, 0, locationLondon)}
	steppingClock := clock.NewSteppingClock(&times)

	writer := newDailyWriterWithClock(steppingClock, ".", "foo.", ".bar")

	// Write to the log for the 14th.
	n, err := writer.Write(buffer1)
	if err != nil {
		t.Fatalf("Write failed - %v", err)
	}

	if n != len(buffer1) {
		t.Fatalf("Write returns %d - expected %d", n, len(buffer1))
	}

	// roll the log over.
	writer.rotateLogs()

	// This should write to expectedFilename.
	n, err = writer.Write(buffer2)
	if err != nil {
		t.Fatalf("Write failed - %v", err)
	}

	if n != len(buffer2) {
		t.Fatalf("Write returns %d - expected %d", n, len(buffer2))
	}

	// The current directory should contain expectedLogfile and one other file.
	// Check that thehe two files exist.
	files, err := ioutil.ReadDir(directoryName)
	if err != nil {
		t.Fatalf("error scanning directory %s - %s", directoryName, err.Error())
	}

	if len(files) != 2 {
		t.Fatalf("directory %s contains %d files.  Should contain just 2.",
			directoryName, len(files))
	}

	if files[0].Name() != expectedFilename &&
		files[1].Name() != expectedFilename {

		t.Fatalf("directory %s contains file \"%s\", expected \"%s\" or \"%s\".",
			directoryName, files[0].Name(), directoryName, expectedFilename)
	}

	// Check the contents.
	expectedPathName := directoryName + "/" + expectedFilename
	inputFile, err := os.OpenFile(expectedPathName, os.O_RDONLY, 0644)
	defer inputFile.Close()
	b := make([]byte, 8096)
	length, err := inputFile.Read(b)
	if err != nil {
		t.Fatalf("error reading logfile %s back - %v", expectedFilename, err)
	}
	if length != len(expectedMessage) {
		t.Fatalf("logfile contains %d bytes - expected %d", length, len(expectedMessage))
	}
	result1 := string(b[:length])
	if result1 != expectedMessage {
		t.Fatalf("logfile contains \"%s\" - expected \"%s\"", result1, expectedMessage)
	}
}

// TestAppendOnRestart checks that if the program creates a log file for the day,
// then crashes and restarts, the Writer appends to the existing file rather than
// overwriting it.
//
func TestAppendOnRestart(t *testing.T) {

	// NOTE:  this test uses the filestore.

	const expectedMessage1 = "goodbye "
	buffer1 := []byte(expectedMessage1)
	const expectedMessage2 = "cruel world"
	buffer2 := []byte(expectedMessage2)
	const expectedFilename = "log.2020-02-14.txt"
	const expectedFirstContents = "goodbye "
	const expectedFinalContents = "goodbye cruel world"

	directoryName, err := createWorkingDirectory()
	if err != nil {
		t.Fatalf("createWorkingDirectory failed - %v", err)
	}
	defer removeWorkingDirectory(directoryName)

	locationUTC, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatalf("error while loading UTC timezone - %v", err)
	}

	// Write some text to the logger.
	// That should create a file for today.
	stoppedClock := clock.NewStoppedClock(2020, time.February, 14, 0, 1, 30, 0, locationUTC)
	writer1 := newDailyWriterWithClock(stoppedClock, ".", "log.", ".txt")
	n, err := writer1.Write(buffer1)
	if err != nil {
		t.Fatalf("Write failed - %v", err)
	}
	if n != len(expectedMessage1) {
		t.Fatalf("Write returns %d - expected %d", n, len(expectedMessage1))
	}

	inputFile, err := os.OpenFile(expectedFilename, os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file %s - %v", expectedFilename, err)
	}
	defer inputFile.Close()
	inputBuffer := make([]byte, 8096)
	n, err = inputFile.Read(inputBuffer)
	if err != nil {
		t.Fatalf("error reading logfile back - %v", err)
	}
	contents := string(inputBuffer[:n])
	if contents != expectedFirstContents {
		t.Fatalf("logfile contains \"%s\" - expected \"%s\"", contents, expectedFirstContents)
	}

	// Create a new writer.  On the first call it will behave as on system startup.  It should
	// append to the existing daily log.
	stoppedClock = clock.NewStoppedClock(2020, time.February, 14, 0, 2, 30, 0, locationUTC)
	writer2 := newDailyWriterWithClock(stoppedClock, ".", "log.", ".txt")
	n, err = writer2.Write(buffer2)
	if err != nil {
		t.Fatalf("Write failed - %v", err)
	}
	if n != len(buffer2) {
		t.Fatalf("Write returns %d - expected %d", n, len(expectedMessage2))
	}

	// Check that only one log file was created and contains the expected contents.
	files, err := ioutil.ReadDir(directoryName)
	if err != nil {
		t.Fatalf("error scanning directory %s - %s", directoryName, err.Error())
	}

	if len(files) != 1 {
		t.Fatalf("directory %s contains %d files.  Should contain exactly one.", directoryName, len(files))
	}

	if files[0].Name() != expectedFilename {
		t.Fatalf("directory %s contains file \"%s\", expected \"%s\".", directoryName, files[0].Name(), expectedFilename)
	}

	// Check the contents.  It should be the result of the two Write calls.

	inputFile, err = os.OpenFile(expectedFilename, os.O_RDONLY, 0644)
	defer inputFile.Close()
	inputBuffer = make([]byte, 8096)
	n, err = inputFile.Read(inputBuffer)
	if err != nil {
		t.Fatalf("error reading logfile back - %v", err)
	}

	contents = string(inputBuffer[:n])
	if contents != expectedFinalContents {
		t.Fatalf("logfile contains \"%s\" - expected \"%s\"", contents, expectedFinalContents)
	}
}

// makeUUID creates a UUID.  See https://yourbasic.org/golang/generate-uuid-guid/.
//
func makeUUID() string {
	// Produces something like "9e0825f2-e557-28df-93b7-a01c789f36a8".
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

// createWorkingDirectory create a working directory and makes it the current
// directory.
//
func createWorkingDirectory() (string, error) {
	directoryName := "/tmp/" + makeUUID()
	err := os.Mkdir(directoryName, os.ModePerm)
	if err != nil {
		return "", err
	}
	err = os.Chdir(directoryName)
	if err != nil {
		return "", err
	}
	return directoryName, nil
}

// removeWorkingDirectory removes the working directory and any files in it.
//
func removeWorkingDirectory(directoryName string) error {
	err := os.RemoveAll(directoryName)
	if err != nil {
		return err
	}
	return nil
}
