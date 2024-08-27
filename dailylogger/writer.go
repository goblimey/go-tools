package dailylogger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/goblimey/go-tools/clock"
	"github.com/goblimey/go-tools/switchwriter"
)

// Writer satisfies the io.Writer interface and writes data to a log file.
// The name of the logfile contains a datestamp in yyyy-mm-dd format. When the
// logger is created the caller can specify leading and trailing text.  For
// example the name for a logfile created on the 5th October 2020 with leader
// "data" and trailer "log" would be "data.20201005.log".
//
// The Writer rolls the log over at midnight at the start of each day - it
// closes yesterday's log and creates today's.
//
// On start up, the first call of New creates today's log file if it doesn't
// already exist.  If the file has already been created, the Writer appends to
// the existing contents.
//
// The Writer contains a mutex.  It's dangerous to copy an object that contain a
// mutex, so you should always call its methods via a pointer.  The New function
// returns a pointer, so that's a good way to create a DailyLogger.
//
// To allow for good unit testing the Writer's notion of the time of day is
// controlled by a Clock object.  The New factory function creates a Clock which
// is a thin wrapper around the standard time service.  That should be used by
// production code.  Tests can use a clock object that supplies predefined values.
type Writer struct {
	logMutex        sync.Mutex
	loggingDisabled bool                 // True if logging is disable. (Logging is enabled by default.)
	clock           clock.Clock          // The system clock in production, a fake in testing.
	startOfToday    time.Time            // The current datestamp for the log.
	logDir          string               // The log directory.
	leader          string               // The leading part of he log file name.
	trailer         string               // The trailing part of the log file name.
	switchwriter    *switchwriter.Writer // The connection to the log file.
}

// This is a compile-time check that Writer implements the io.Writer interface.
var _ io.Writer = (*Writer)(nil)

// New creates a Writer, starts the log rotator and returns the writer.  Production
// code should call this to get a Writer.
func New(logDir, leader, trailer string) *Writer {

	dw := newWriter(nil, logDir, leader, trailer)

	// Start a goroutine to roll the log over at the end of each day.
	go dw.logRotator()
	return dw
}

// newWriter creates a daily writer with a supplied switchwriter and clock,
// and returns a pointer to it. This is called by New as a helper method and by
// unit tests.
func newWriter(cl clock.Clock, logDir, leader, trailer string) *Writer {

	// The logfile is of the form "logDir/leader.yyyy-mm-dd.trailer".  The default
	// is "./daily.yyyy-mm-dd.log".
	const defaultLeader = "daily."
	const defaultTrailer = ".log"
	const defaultLogDir = "."

	if logDir == "" {
		logDir = defaultLogDir
	}

	if leader == "" {
		leader = defaultLeader
	}

	if trailer == "" {
		trailer = defaultTrailer
	}

	if cl == nil {
		// If the supplied clock is nil use the real system clock.
		cl = clock.NewSystemClock()
	}

	sw := switchwriter.New()

	dw := Writer{clock: cl, switchwriter: sw,
		logDir: logDir, leader: leader, trailer: trailer}

	// Create the log directory if it doesn't already exist.
	createlogDirectory(logDir)

	// Create today's log file and switch the switchwriter to it.
	dw.startOfToday = getLastMidnight(cl.Now())
	dw.openLog(dw.startOfToday)

	return &dw
}

// Write writes the buffer to the daily log file, creating the file at the
// start of each day.
func (dw *Writer) Write(buffer []byte) (int, error) {
	if dw.loggingDisabled {
		return 0, nil
	} else {
		// Avoid a race with rotateLogs.
		dw.logMutex.Lock()
		defer dw.logMutex.Unlock()

		// Write to the log.
		n, err := dw.switchwriter.Write(buffer)
		return n, err
	}

}

// EnableLogging switches logging on.
func (dw *Writer) EnableLogging() {
	dw.loggingDisabled = false
}

// DisableLogging switches logging off - Write calls do nothing.
func (dw *Writer) DisableLogging() {
	dw.loggingDisabled = true
}

// GetTimestampForLog gets the timestamp for a log entry, for example
// "2020-02-14 15:42:11.789 UTC".  This uses the real time, not the supplied clock.
func (dw *Writer) GetTimestampForLog() string {
	now := time.Now()
	return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d.%03d %s",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1000000,
		now.Location().String())
}

// setClock sets the clock.  This is used for unit testing.
func (dw *Writer) setClock(clock clock.Clock) {
	dw.clock = clock
}

// rotator() runs forever, rotating the log files at the end of each day.
func (dw *Writer) logRotator() {

	// This should be run in a goroutine.
	//
	// As it runs forever it can't be unit tested.  It uses the real time,
	// not the supplied clock, which is for testing.

	for {
		// Sleep until the end of day
		now := time.Now()
		waitTime := getDurationToMidnight(now)
		time.Sleep(waitTime)

		// Wake up and rotate the log file using the next
		// day as the timestamp.
		//
		// If the system is running properly, It could by now
		// be a fraction of a second before midnight or (more
		// likely) a fraction of a second after.  If the system
		// gets very slow for some reason, it could be much
		// later than that.  In the very worst case, a later
		// day altogether, but that's *very* unlikely.
		dw.rotateLogs()
	}
}

// rotateLogs() rotates the daily log files.
func (dw *Writer) rotateLogs() {
	// Avoid a race with Write.
	dw.logMutex.Lock()
	defer dw.logMutex.Unlock()
	currentTime := dw.clock.Now()
	dw.closeLog()

	// Advance the current day.
	// This should be happening just before or just after midnight
	// so the start of today should be the same as the stored
	// start of day or (more likely) that time plus one day.  If
	// the goroutine has drifted, start of today may be yet later
	// but this sequence will fix that.
	startOfToday := getLastMidnight(currentTime)

	if startOfToday.After(dw.startOfToday) {
		// Example: the stored start of day is midnight on the 1st.
		// It's now some day after that, probably the 2nd, so the
		// stored start of day should be 00:00:00 on the 2nd - the
		// last midnight.
		dw.startOfToday = getLastMidnight(startOfToday)
	} else {
		// Example, the stored start of day is the 1st.  We are
		// running just before midnight so the calculated start
		// of day is also the 1st, not the 2nd.  The stored start
		// of day should be 00:00:00 on the 2nd - the next midnight.
		dw.startOfToday = getNextMidnight(startOfToday)
	}

	// Open the logfile using start of today as the timestamp.

	dw.openLog(dw.startOfToday)
}

// CreateLogDirectory creates the log directory if it does not
// already exist.
func createlogDirectory(directory string) {
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		log.Fatalf("cannot create log directory " + directory + " - " + err.Error())
	}
}

// Create the log directory if it doesn't already exist.

// closeLog is a helper function that closes the log file (which
// also flushes any uncommitted writes).  It doesn't apply the
// lock so it should only be called by a function that does.
func (dw *Writer) closeLog() {
	dw.switchwriter.SwitchTo(nil)
}

// openLog is a helper function that opens today's log.  It doesn't
// apply the lock, so it should only be done by something that does.
func (dw *Writer) openLog(startOfToday time.Time) {

	// Create the log directory
	pathname := dw.getLogPathname(startOfToday)

	logFile, err := openFile(pathname)
	if err != nil {
		log.Printf("openLog: error creating log file %s - %s\n",
			pathname, err.Error())
		// Continue - file is now nil.
	}
	dw.switchwriter.SwitchTo(logFile)
}

// getLogPathname returns today's log filename, for example "data.2020-01-19.rtcm3".
// The time is supplied to aid unit testing.
func (dw *Writer) getLogPathname(now time.Time) string {

	return fmt.Sprintf("%s/%s%04d-%02d-%02d%s",
		dw.logDir, dw.leader, now.Year(), int(now.Month()), now.Day(), dw.trailer)
}

// openFile either creates and opens the file or, if it already exists, opens it
// in append mode.
func openFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Seek(0, 2)
	if err != nil {
		log.Fatal(err)
	}
	return file, nil
}

// getLastMidnight gets midnight at the beginning of the day given by the start time.
func getLastMidnight(start time.Time) time.Time {
	return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
}

// getNextMidnight gets midnight at the beginning of the day after the given start time.
func getNextMidnight(start time.Time) time.Time {
	nextDay := start.AddDate(0, 0, 1)
	return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, start.Location())
}

// getDurationToMidnight gets the duration between the start time
// and midnight at the beginning of the next day in the same timezone.
func getDurationToMidnight(start time.Time) time.Duration {
	nextMidnight := getNextMidnight(start)
	return nextMidnight.Sub(start)
}
