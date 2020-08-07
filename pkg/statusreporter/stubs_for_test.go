package statusreporter

import (
	"errors"
	"net/http"
)

const maxBufferLength = 1024

// ReportFeedForTest respects the status-reporter ReportFeedT interface.
type ReportFeedForTest struct {
	// LogLevel is the log level - 0 is errors only.
	LogLevel uint8
}

//SetLogLevel satisfies the ReportFeedT interface.
func (trf *ReportFeedForTest) SetLogLevel(level uint8) {
	trf.LogLevel = level
}

//Status satisfies the ReportFeedT interface.
func (trf *ReportFeedForTest) Status() []byte {
	reportBody := "foo"
	return []byte(reportBody)
}

//GetLogLevel returns the logging level.
func (trf ReportFeedForTest) GetLogLevel() uint8 {
	return trf.LogLevel
}

// ResponseWriterForTest satisfies the http.ResponseWriter interface and logs what is written.
type ResponseWriterForTest struct {

	// the body records anything written by Write.
	body *[]byte

	// Length of body used.
	length *int
	// HeaderValue records anything written by WriteHeader.
	headerValue *int
}

func MakeResponseWriterForTest() ResponseWriterForTest {
	writer := new(ResponseWriterForTest)
	writer.Init()

	return *writer
}

// Header satisfies the http.ResponseWriter interface
func (trw ResponseWriterForTest) Header() http.Header {
	return *new(http.Header)
}

// Write satisfies the http.ResponseWriter interface
func (trw ResponseWriterForTest) Write(buf []byte) (int, error) {
	if trw.Length()+len(buf) < maxBufferLength {
		for i := 0; i < len(buf); i++ {
			(*trw.body)[*trw.length] = buf[i]
			*trw.length++
		}
		return len(buf), nil
	} else {
		return 0, errors.New("buffer overflow")
	}
}

// WriteHeader satisfies the http.ResponseWriter interface
func (trw ResponseWriterForTest) WriteHeader(statusCode int) {
	*trw.headerValue = statusCode
}

// CreateBody allocates space for the body
func (trw *ResponseWriterForTest) Init() {
	b := make([]byte, 1024, 1024)
	trw.body = &b
	var i int = 0
	trw.length = &i
	headerValue := -1
	trw.headerValue = &headerValue
}

// Body gets the body
func (trw ResponseWriterForTest) Body() []byte {
	return *trw.body
}

//Length gets the length used of the body.
func (trw ResponseWriterForTest) Length() int {
	return *trw.length
}

// Headervalue gets the last written header value.
func (trw ResponseWriterForTest) HeaderValue() int {
	return *trw.headerValue
}

