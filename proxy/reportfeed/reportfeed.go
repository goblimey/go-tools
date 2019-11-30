package reportfeed

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/goblimey/go-tools/logger"
)

// Buffer contains an input or output buffer.
type Buffer struct {
	Timestamp     time.Time
	Source        uint64
	Content       *[]byte
	ContentLength int
}

const reportFormat = `
<h3>Last Client Buffer</h3>
<span id='clienttimestamp'>%s</span>
<div class="preformatted" id='clientbuffer'>
%s
</div>
<h3>Last Server Buffer</h3>
<span id='servertimestamp'>%s</span>
<div class="preformatted" id='serverbuffer'>
%s
</div>
`

// ReportFeed satisfies the status-reporter ReportFeedT interface.
type ReportFeed struct {
	// LogLevel is the log level - 0 is errors only.
	logger *logger.LoggerT
	lastClientBuffer *Buffer
	lastServerBuffer *Buffer
	mutex            sync.Mutex
}

// MakeReportFeed creates and returns a new ReportFeed object
func MakeReportFeed(logger *logger.LoggerT) *ReportFeed {
	var reportFeed ReportFeed
	reportFeed.SetLogger(logger)
	return &reportFeed
}

//SetLogLevel satisfies the ReportFeedT interface.
func (rf *ReportFeed) SetLogLevel(level uint8) {
	rf.logger.SetLogLevel(level)
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

func (rf *ReportFeed) SetLogger(logger *logger.LoggerT) {
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