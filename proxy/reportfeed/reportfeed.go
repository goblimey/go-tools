package reportfeed

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/goblimey/go-tools/dailylogger"
)

// Buffer contains an input or output buffer.
type Buffer struct {
	Timestamp     time.Time
	Source        uint64
	Content       *[]byte
	ContentLength int
}

// ReportFeed satisfies the status-reporter ReportFeedT interface.
type ReportFeed struct {
	logger           *dailylogger.DailyWriter
	lastClientBuffer *Buffer
	lastServerBuffer *Buffer
	mutex            sync.Mutex
}

// MakeReportFeed creates and returns a new ReportFeed object
func MakeReportFeed(logger *dailylogger.DailyWriter) *ReportFeed {
	var reportFeed ReportFeed
	reportFeed.SetLogger(logger)
	return &reportFeed
}

//SetLogLevel satisfies the ReportFeedT interface.
func (rf *ReportFeed) SetLogLevel(level uint8) {
	if level == 0 {
		rf.logger.DisableLogging()
	} else {
		rf.logger.EnableLogging()
	}
}

//Status satisfies the ReportFeedT interface.
func (rf *ReportFeed) Status() []byte {
	clientLeader := "no input buffer"
	clientHexDump := ""
	serverLeader := "no output buffer"
	serverHexDump := ""
	rf.mutex.Lock()
	defer rf.mutex.Unlock()
	if rf.lastClientBuffer != nil && rf.lastClientBuffer.Content != nil {
		fmt.Fprintf(os.Stderr, "client buffer")
		clientLeader = fmt.Sprintf("From Client [%d]:\n%s\n", rf.lastClientBuffer.Source,
			rf.lastClientBuffer.Timestamp.Format("Mon Jan _2 15:04:05 2006"))

		clientHexDump =
			Sanitise(hex.Dump((*rf.lastClientBuffer.Content)[:rf.lastClientBuffer.ContentLength]))
	}
	if rf.lastServerBuffer != nil && rf.lastServerBuffer.Content != nil {
		fmt.Fprintf(os.Stderr, "server buffer")
		serverLeader = fmt.Sprintf("To Server [%d]:\n%s\n", rf.lastServerBuffer.Source,
			rf.lastServerBuffer.Timestamp.Format("Mon Jan _2 15:04:05 2006"))
		serverHexDump =
			Sanitise(hex.Dump((*rf.lastServerBuffer.Content)[:rf.lastServerBuffer.ContentLength]))
	}

	reportBody := fmt.Sprintf(reportFormat,
		clientLeader,
		clientHexDump,
		serverLeader,
		serverHexDump)

	return []byte(reportBody)
}

// SetLogger sets the logger.
func (rf *ReportFeed) SetLogger(logger *dailylogger.DailyWriter) {
	rf.logger = logger
}

// RecordClientBuffer takes a timestamped copy of a client buffer.
func (rf *ReportFeed) RecordClientBuffer(buffer *[]byte, source uint64, length int) {
	rf.mutex.Lock()
	defer rf.mutex.Unlock()
	rf.lastClientBuffer = &Buffer{time.Now(), uint64(source), buffer, length}
}

// RecordServerBuffer takes a timestamped copy of a server buffer.
func (rf *ReportFeed) RecordServerBuffer(buffer *[]byte, source uint64, length int) {
	rf.mutex.Lock()
	defer rf.mutex.Unlock()
	rf.lastServerBuffer = &Buffer{time.Now(), uint64(source), buffer, length}
}

// Sanitise edits a string, replacing some dangerous HTML characters.
func Sanitise(s string) string {
	s = strings.Replace(s, "<", "&lt;", -1)
	s = strings.Replace(s, ">", "&gt;", -1)
	return s
}
