package dailylogger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/goblimey/go-tools/clock"
	"github.com/goblimey/go-tools/switchwriter"
)

// Writer satisfies the io.Writer interface and writes data to a log file.  The
// name of the logfile contains a datestamp in yyyy-mm-dd format. When the logger
// is created the caller can specify leading and trailing text.  For example the
// name for a logfile created on the 5th October 2020 with leader "data" and
// trailer "log" would be "data.20201005.log".
//
// The Write rolls the log over at midnight at the start of each day - it closes
// yesterday's log and creates today's.
//
// On start up, the first call of Writer creates today's log file if it doesn't
// already exist.  If the files has already been created, the Writer appends to
// the existing contents.
//
// To allow for good unit testing the Writer's notion of the time of day is
// controlled by a Clock object.  The New factory function creates a Clock which
// is a thin wrapper around the standard time service.  That should be used by
// production code.  Tests can use a clock object that supplies predefined values.
//
// The Writer contains a mutex.  It's dangerous to copy an object that contain a
// mutex, so you should always call its methods via a pointer.  The New function
// returns a pointer, so that's a good way to create a DailyLogger.
//
type Writer struct {
	logMutex     sync.Mutex
	clock        clock.Clock          // The system clock in production, a fake in testing.
	startOfToday time.Time            // The current datestamp for the log.
	logDir       string               // Create logfiles in this directory.
	leader       string               // The leading part of the logfile name
	trailer      string               // The trailing part of the logfile name
	switchwriter *switchwriter.Writer // The connection to the log file.
}

// This is a compile-time check that Writer implements the io.Writer interface.
var _ io.Writer = (*Writer)(nil)

// New creates a Writer, starts the log rotator and returns the write as an
// io.Writer.  It uses a SystemClock.  Production code should call this to
// geta daily logger.
//
func New(logDir, leader, trailer string) io.Writer {

	writer := newDailyWriterWithClock(clock.NewSystemClock(), logDir, leader, trailer)
	// Start a goroutine to roll the log over at midnight every night.
	go writer.logRotator()
	return writer
}

// newDailyWriterWithClock creates a Writer with a supplied clock and returns a
// pointer to it. (This is used by New and by unit tests.)
//
func newDailyWriterWithClock(clock clock.Clock, logDir, leader, trailer string) *Writer {

	// By default, the logfile is of the form "daily.yyyy-mm-dd.log" and in the current directory.
	const defaultLeader = "daily."
	const defaultTrailer = ".log"
	const defaultLogDir = "."

	if leader == "" {
		leader = defaultLeader
	}

	if trailer == "" {
		trailer = defaultTrailer
	}

	if logDir == "" {
		logDir = defaultLogDir
	}

	// Create the log directory if it doesn't already exist.
	err := os.Mkdir(logDir, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatalf("cannot create log directory " + logDir + " - " + err.Error())
	}

	switchwriter := switchwriter.New()
	writer := Writer{clock: clock, switchwriter: switchwriter,
		logDir: logDir, leader: leader, trailer: trailer}
	// Create today's log file.
	writer.startOfToday = getLastMidnight(clock.Now())
	writer.openLog(writer.startOfToday)
	return &writer
}

// Write writes the buffer to the daily log file, creating the file at the start of each day.
//
func (lw *Writer) Write(buffer []byte) (n int, err error) {

	// Avoid a race with rotateLogs.
	lw.logMutex.Lock()
	defer lw.logMutex.Unlock()

	// Write to the log.
	n, err = lw.switchwriter.Write(buffer)

	return n, err
}

// GetTimestampForLog gets the timestamp for a log entry, for example
// "2020-02-14 15:42:11.789 UTC".  This uses the real time, not the supplied clock.
//
func (lw *Writer) GetTimestampForLog() string {
	now := time.Now()
	return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d.%03d %s",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1000000,
		now.Location().String())
}

// setClock sets the clock.  This is used for unit testing.
func (lw *Writer) setClock(clock clock.Clock) {
	lw.clock = clock
}

// rotator() runs forever, rotating the log files at the end of each day.
func (lw Writer) logRotator() {

	// This should be run in a goroutine.
	//
	// As it runs forever it can't be unit tested.  It uses the real time,
	// not the supplied clock, which is for testing.

	for {
		// Sleep until the end of day
		waitTime := getDurationToMidnight(time.Now())
		fmt.Printf("rotateLogs: sleeping for %d\n", waitTime)
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
		lw.rotateLogs()
	}
}

// rotateLogs() rotates the daily log files.
func (lw Writer) rotateLogs() {
	// Avoid a race with Write.
	lw.logMutex.Lock()
	defer lw.logMutex.Unlock()
	currentTime := lw.clock.Now()
	fmt.Printf("rotateLogs: end of day %v\n", currentTime)
	lw.closeLog()

	// Advance the current day.
	// This should be happening just before or just after midnight
	// so the start of today should be the same as the stored
	// start of day or (more likely) that time plus one day.  If
	// the goroutine has drifted, start of today may be yet later
	// but this sequence will fix that.
	startOfToday := getLastMidnight(currentTime)

	if startOfToday.After(lw.startOfToday) {
		// Example: the stored start of day is midnight on the 1st.
		// It's now some day after that, probably the 2nd, so the
		// stored start of day should be 00:00:00 on the 2nd - the
		// last midnight.
		lw.startOfToday = getLastMidnight(startOfToday)
	} else {
		// Example, the stored start of day is the 1st.  We are
		// running just before midnight so the calculated start
		// of day is also the 1st, not the 2nd.  The stored start
		// of day should be 00:00:00 on the 2nd - the next midnight.
		lw.startOfToday = getNextMidnight(startOfToday)
	}

	// Open the logfile using start of today as the timestamp.

	lw.openLog(lw.startOfToday)
}

// closeLog is a helper function that closes the log file (which
// also flushes any uncommitted writes).  It doesn't apply the
// lock so it should only be called by a function that does.
//
func (lw Writer) closeLog() {
	lw.switchwriter.SwitchTo(nil)
}

// openLog is a helper function that opens today's log.  It doesn't
// apply the lock, so it should only be done by something that does.
func (lw *Writer) openLog(startOfToday time.Time) {

	filename := lw.getFilename(startOfToday)
	pathname := lw.logDir + "/" + filename
	logFile, err := openFile(pathname)
	if err != nil {
		log.Printf("openLog: error creating log file %s - %s\n",
			pathname, err.Error())
		// Continue - file is now nil.
	}
	lw.switchwriter.SwitchTo(logFile)
}

// getFilename returns today's log filename, for example "data.2020-01-19.rtcm3".
// The time is supplied to aid unit testing.
//
func (lw Writer) getFilename(now time.Time) string {

	return fmt.Sprintf("%s%04d-%02d-%02d%s",
		lw.leader, now.Year(), int(now.Month()), now.Day(), lw.trailer)
}

//openFile either creates and opens the file or, if it already exists, opens it
// in append mode.
//
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

// getDurationToMidnight gets the duration between the start time and midnight at the
// beginning of the next day in the same timezone.
//
func getDurationToMidnight(start time.Time) time.Duration {
	nextMidnight := getNextMidnight(start)
	return nextMidnight.Sub(start)
}

// newLogFile returns a newly created logfile.
// It doesn't set the mutex so it must be called by something that does.
//
func (lw *Writer) newLogFile(startOfToday time.Time, logDir, leader, trailer string) (*os.File, error) {
	logName := lw.getFilename(startOfToday)
	newLog, err := openFile(logName)
	if err != nil {
		return nil, err
	}
	lw.switchwriter.SwitchTo(newLog)
	return newLog, nil
}

// check provides simple error handling.
//
func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
