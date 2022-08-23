package statusreporter

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"

	htmlTemplate "html/template"
	textTemplate "text/template"
)

// DefaultServiceNamePriv is the default name of the web service.
const DefaultServiceNamePriv = "status"

//StylesheetRequestEnd defines the end of the HTTP request for the CSS stylesheet.
const StylesheetRequestEnd = "/stylesheet.css"

// StatusRequestEnd defines the end of the HTTP request for the status page.
const StatusRequestEnd = "/report"

// LogLevelRequestMiddle defines the middle part of the HTTP loglevel request.
const LogLevelRequestMiddle = "/loglevel/"

// ReportFeedT defines the methods of the control object that the caller must suppy.
type ReportFeedT interface {
	// SetLogLevel sets the log level - 0 disables logging, anything else enables it.
	SetLogLevel(n uint8)
	// Status gets the body of a status report
	Status() []byte
}

// StatusReport contains the data for the status report page.
type StatusReport struct {
	PageTitle   string
	ServiceName string
	Content     string
}

// Reporter provides the reporting API.
type Reporter struct {
	// ReportFeedPriv is a reference to a ReporterFeed instance
	ReportFeedPriv ReportFeedT
	// UseTextTemplates makes the status report use text templates rather than html templates.
	UseTextTemplates bool
	// ServiceNamePriv is the name of the web service.  All requests start with this name.  Default is "status"
	ServiceNamePriv string
	// ServiceHostPriv is the name of the http host.
	ServiceHostPriv string
	// ServicePortPriv is the name of the http port.
	ServicePortPriv int
	// StylesheetRequestPriv defines the HTTP request for the stylesheet
	StylesheetRequestPriv string
	// StatusRequestPriv defines the name of the status request, eg /status/report
	StatusRequestPriv string
	// LogLevelRequestPriv defines the start of the set log level request, eg "/status/loglevel/".
	LogLevelRequestPriv string
	// LogLevelRequestRE is the regular expression defining the log level request including the level number,
	// eg "^/status/loglevel/([0-9]+)$"
	LogLevelRequestRE *regexp.Regexp
	// TextReportTemplate is the text template for the status report page.
	TextReportTemplate *textTemplate.Template
	// HTMLReportTemplate is the html template for the status report page.
	HTMLReportTemplate *htmlTemplate.Template
	// ErrorTemplate is the html template for the error page.
	ErrorTemplate *htmlTemplate.Template
}

// MakeReporter creates and returns a reporter object
func MakeReporter(reportFeed ReportFeedT, host string, port int) Reporter {
	var reporter Reporter
	reporter.InitTemplates()
	reporter.SetReportFeed(reportFeed)
	reporter.SetServiceName(DefaultServiceNamePriv)
	reporter.SetServiceHost(host)
	reporter.SetServicePort(port)
	reporter.SetRequests() // set to default values - "/status/...."
	return reporter
}

// SetUseTextTemplates sets the UseTextTeplates flag.
func (r *Reporter) SetUseTextTemplates(useTextTemplates bool) {
	r.UseTextTemplates = useTextTemplates
}

// SetReportFeed sets the reporter feed.
func (r *Reporter) SetReportFeed(reportFeed ReportFeedT) {
	r.ReportFeedPriv = reportFeed
}

// SetServiceName sets the name of the the web service.  All requests start with this.
func (r *Reporter) SetServiceName(name string) {
	r.ServiceNamePriv = name
	r.SetRequests()
}

// SetServiceHost sets the hostname that web service answers to.
func (r *Reporter) SetServiceHost(host string) {
	r.ServiceHostPriv = host
}

// SetServicePort sets the port that the web service listens on.
func (r *Reporter) SetServicePort(port int) {
	r.ServicePortPriv = port
}

// HandleStatusRequest handles the request for a status report by displaying the last input and output buffers.
func (r *Reporter) HandleStatusRequest(writer http.ResponseWriter, request *http.Request) {
	if r.TextReportTemplate == nil {
		r.InitTemplates()
	}
	body := string(r.ReportFeedPriv.Status())
	statusReport := StatusReport{"Status", r.ServiceNamePriv, body}
	if r.UseTextTemplates {
		// The supplied r.ReportFeedPriv.Status() value is expected to contain
		// HTML tags so we need to use the less secure text template.  That
		// introduces the risk of HTML in the buffers being used for an injection
		// attack, so we disallow HTML there.
		err := r.TextReportTemplate.Execute(writer, statusReport)
		if err != nil {
			em := fmt.Sprintf("error displaying page - %s", err.Error())
			fmt.Fprintf(os.Stderr, "%s\n", em)
			err = r.ErrorTemplate.Execute(writer, em)
			if err != nil {
				writer.Write(internalErrorPage)
				return
			}
			return
		}
	} else {
		// The r.ReportFeedPriv.Status() value should not contains HTML tags so
		// we can use the more secure HTML template.
		err := r.HTMLReportTemplate.Execute(writer, statusReport)
		if err != nil {
			em := fmt.Sprintf("error displaying page - %s", err.Error())
			fmt.Fprintf(os.Stderr, "%s\n", em)
			err = r.ErrorTemplate.Execute(writer, em)
			if err != nil {
				writer.Write(internalErrorPage)
				return
			}
			return
		}
	}
	return
}

// HandleStylesheetRequest handles an HTTP request for the stylesheet.
func (r *Reporter) HandleStylesheetRequest(writer http.ResponseWriter, request *http.Request) {
	_, err := writer.Write(stylesheetPage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error displaying stylesheet - %s\n", err.Error())
		return
	}
	return
}

// HandleLogLevelRequest responds to a /{servicename}/Loglevel/{level} HTTP request.
func (r *Reporter) HandleLogLevelRequest(writer http.ResponseWriter, request *http.Request) {
	url := request.URL
	// The uri is something like "/{servicename}/Loglevel/42".  We want just the "42".
	part := r.LogLevelRequestRE.FindStringSubmatch(url.RequestURI())
	if len(part) <= 1 {
		fmt.Fprintf(os.Stderr, "illegal level in log level request - %s\n", url.RequestURI())
		writer.WriteHeader(400)
		return
	}
	level, err := strconv.ParseUint(part[1], 10, 8)
	if err != nil {
		em := fmt.Sprintf("invalid log level %s in request, must be an unsigned integer 0-255 - %s", part[1], err.Error())
		fmt.Fprintf(os.Stderr, "%s\n", em)
		err = r.ErrorTemplate.Execute(writer, em)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error displaying error report page - %s\n", err.Error())
			writer.Write(internalErrorPage)
			return
		}
		return
	}
	r.ReportFeedPriv.SetLogLevel(uint8(level))
}

// StartService starts the web service.
func (r Reporter) StartService() {
	if r.TextReportTemplate == nil {
		r.InitTemplates()
	}
	// Set the HTTP request handlers.
	http.HandleFunc(r.StatusRequestPriv, r.HandleStatusRequest)
	http.HandleFunc(r.StylesheetRequestPriv, r.HandleStylesheetRequest)
	http.HandleFunc(r.LogLevelRequestPriv, r.HandleLogLevelRequest)

	server := fmt.Sprintf("%s:%d", r.ServiceHostPriv, r.ServicePortPriv)
	fmt.Fprintf(os.Stderr, "listening for status requests on %s\n", server)
	err := http.ListenAndServe(server, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot start http server for status requests - %s\n"+err.Error())
	}
}

// SetRequests sets the names and expressions defining the HTTP requests.
func (r *Reporter) SetRequests() {
	if len(r.ServiceNamePriv) == 0 {
		r.ServiceNamePriv = DefaultServiceNamePriv
	}

	// eg "/status/stylesheet.css".
	r.StylesheetRequestPriv = "/" + r.ServiceNamePriv + StylesheetRequestEnd
	// eg "/status/report".
	r.StatusRequestPriv = "/" + r.ServiceNamePriv + StatusRequestEnd
	// eg "/status/loglevel/"
	r.LogLevelRequestPriv = "/" + r.ServiceNamePriv + LogLevelRequestMiddle
	// eg "^/status/loglevel/([0-9]+)$"
	str := "^" + r.LogLevelRequestPriv + "([0-9]+)$"
	r.LogLevelRequestRE = regexp.MustCompile(str)
}

// InitTemplates initialises the HTML and text templates.
func (r *Reporter) InitTemplates() {
	r.TextReportTemplate = textTemplate.New("report")
	r.TextReportTemplate = textTemplate.Must(r.TextReportTemplate.Parse(reportText + baseText))

	r.HTMLReportTemplate = htmlTemplate.New("report")
	r.HTMLReportTemplate = htmlTemplate.Must(r.HTMLReportTemplate.Parse(reportText + baseText))

	r.ErrorTemplate = htmlTemplate.New("error")
	r.ErrorTemplate = htmlTemplate.Must(r.ErrorTemplate.Parse(errorText + baseText))
}
