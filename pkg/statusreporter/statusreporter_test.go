package statusreporter

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"testing"
)

// The stylesheet link in the expected result depends upon the service name.  Define a Printf format and
// produce versions as required.
const format = `
<!DOCTYPE html>
<html lang="en">
 <head>
 <meta charset="UTF-8">
 <title>Status</title>
 <link href='/%s/stylesheet.css' rel='stylesheet'/>
 </head>
 <body>
 <h2>Status</h2>
 <section id="content">	
foo
 </section>
 </body>
</html>
`

// TestSetLoggingLevelRequest checks the SetLogLevel method with the default service name.
func TestSetLoggingLevelRequest(t *testing.T) {
	serviceName := "status"
	var url url.URL
	url.Opaque = "/" + serviceName + "/loglevel/0" // url.RequestURI() will return this URI
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "POST"
	var writer http.ResponseWriter
	var reportFeed = new(ReportFeedForTest)

	reporter := MakeReporter(reportFeed, "foo", 42)

	reporter.HandleLogLevelRequest(writer, &httpRequest)

	if reportFeed.GetLogLevel() != 0 {
		t.Errorf("Expected logLevel 0, got %v", reportFeed.GetLogLevel())
	}

	url.Opaque = "/status/loglevel/1"

	reporter.HandleLogLevelRequest(writer, &httpRequest)

	if reportFeed.GetLogLevel() != 1 {
		t.Errorf("Expected logLevel 1, got %v", reportFeed.GetLogLevel())
	}
}

// TestSetLoggingLevelRequestWithServiceName checks the SetLogLevel method with a specified service name.
func TestSetLoggingLevelRequestWithServiceName(t *testing.T) {
	serviceName := "someservicename"
	var url url.URL
	url.Opaque = "/" + serviceName + "/loglevel/0" // url.RequestURI() will return this URI
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "POST"
	var writer http.ResponseWriter
	reportFeed := new(ReportFeedForTest)
	reporter := MakeReporter(reportFeed, "foo", 42)
	reporter.SetServiceName(serviceName)

	reporter.HandleLogLevelRequest(writer, &httpRequest)

	if reportFeed.GetLogLevel() != 0 {
		t.Errorf("Expected logLevel 0, got %v", reportFeed.GetLogLevel())
	}

	url.Opaque = "/" + serviceName + "/loglevel/1" // url.RequestURI() will return this URI

	reporter.HandleLogLevelRequest(writer, &httpRequest)

	if reportFeed.GetLogLevel() != 1 {
		t.Errorf("Expected logLevel 1, got %v", reportFeed.GetLogLevel())
	}
}

// TestSetLoggingLevelRequestWithJunkLevel checks that SetLogLevel fails when the level is junk.
func TestSetLoggingLevelRequestWithJunkLevel(t *testing.T) {
	const expectedError = "illegal level in log level request - /status/loglevel/junk"
	serviceName := "status"
	var url url.URL
	url.Opaque = "/" + serviceName + "/loglevel/junk" // url.RequestURI() will return this URI
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "POST"
	responseWriterForTest := NewResponseWriterForTest()
	var reportFeed = new(ReportFeedForTest)

	reporter := MakeReporter(reportFeed, "foo", 42)

	reporter.HandleLogLevelRequest(responseWriterForTest, &httpRequest)

	if responseWriterForTest.HeaderValue() != 400 {
		t.Errorf("Expected logLevel 400, got %d", responseWriterForTest.HeaderValue())
	}
}

// TestStatusRequest checks the response to the report request with the default service name.
func TestStatusRequest(t *testing.T) {

	// Test using HTML templates.
	serviceName := "status"
	expectedHTML := reduceString(fmt.Sprintf(format, serviceName))
	var url url.URL
	url.Opaque = "/" + serviceName + "/report" // url.RequestURI() will return this URI
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "GET"
	var reportFeed = new(ReportFeedForTest)
	responseWriterForTest := NewResponseWriterForTest()

	reporter := MakeReporter(reportFeed, "foo", 42)
	reporter.HandleStatusRequest(responseWriterForTest, &httpRequest)

	body := string(responseWriterForTest.Body()[:responseWriterForTest.Length()])
	body = reduceString(body)
	if body != expectedHTML {
		fmt.Fprintf(os.Stdout, "expected |%s|\n", expectedHTML)
		fmt.Fprintf(os.Stdout, "actual   |%s|\n", body)
		t.Errorf("{html,%s} the HTML page does not contain the expected contents", serviceName)
	}

	// Test using text templates.
	url.Opaque = "/" + serviceName + "/report" // url.RequestURI() will return this URI
	reportFeed = new(ReportFeedForTest)
	responseWriterForTest = NewResponseWriterForTest()
	reporter = MakeReporter(reportFeed, "foo", 42)
	reporter.SetUseTextTemplates(true)
	reporter.HandleStatusRequest(responseWriterForTest, &httpRequest)

	body = string(responseWriterForTest.Body()[:responseWriterForTest.Length()])
	body = reduceString(body)
	if body != expectedHTML {
		fmt.Fprintf(os.Stdout, "expected |%s|\n", expectedHTML)
		fmt.Fprintf(os.Stdout, "actual   |%s|\n", body)
		t.Errorf("{html,%s} the HTML page does not contain the expected contents", serviceName)
	}
}

// TestStatusRequestWithServiceName checks the response to the report request with a given service name.
func TestStatusRequestWithServiceName(t *testing.T) {

	// Test using HTML templates.
	serviceName := "someservicename"
	expectedHTML := reduceString(fmt.Sprintf(format, serviceName))
	var url url.URL
	url.Opaque = "/" + serviceName + "/report" // url.RequestURI() will return this URI
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "GET"
	var reportFeed = new(ReportFeedForTest)
	responseWriterForTest := NewResponseWriterForTest()
	reporter := MakeReporter(reportFeed, "foo", 42)
	reporter.SetServiceName(serviceName)
	reporter.HandleStatusRequest(responseWriterForTest, &httpRequest)

	body := string(responseWriterForTest.Body()[:responseWriterForTest.Length()])
	body = reduceString(body)
	if body != expectedHTML {
		fmt.Fprintf(os.Stdout, "expected |%s|\n", expectedHTML)
		fmt.Fprintf(os.Stdout, "actual   |%s|\n", body)
		t.Errorf("{html,%s} the HTML page does not contain the expected contents", serviceName)
	}

	// Test using text templates.
	url.Opaque = "/" + serviceName + "/report" // url.RequestURI() will return this URI
	reportFeed = new(ReportFeedForTest)
	responseWriterForTest = NewResponseWriterForTest()
	reporter = MakeReporter(reportFeed, "foo", 42)
	reporter.SetServiceName(serviceName)
	reporter.SetUseTextTemplates(true)
	reporter.HandleStatusRequest(responseWriterForTest, &httpRequest)

	body = string(responseWriterForTest.Body()[:responseWriterForTest.Length()])
	body = reduceString(body)
	if body != expectedHTML {
		fmt.Fprintf(os.Stdout, "expected |%s|\n", expectedHTML)
		fmt.Fprintf(os.Stdout, "actual   |%s|\n", body)
		t.Errorf("{html,%s} the HTML page does not contain the expected contents", serviceName)
	}
}

// reduceString removes all newlines and reduces all other white space to a single space.
func reduceString(str string) string {
	re := regexp.MustCompile(`(?)\n+`)
	str = re.ReplaceAllString(str, "")
	re = regexp.MustCompile(`[ \t]+`)
	return re.ReplaceAllString(str, " ")
}
