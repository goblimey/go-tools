module github.com/goblimey/go-tools/proxy

require (
	github.com/goblimey/go-tools/dailylogger v0.0.6
	github.com/goblimey/go-tools/proxy/reportfeed v0.0.6
	github.com/goblimey/go-tools/statusreporter v0.0.6
	github.com/goblimey/go-tools/testsupport v0.0.6
)

replace (
	github.com/goblimey/go-tools/clock => ../../pkg/clock
	github.com/goblimey/go-tools/dailylogger => ../../pkg/dailylogger
	github.com/goblimey/go-tools/proxy/reportfeed => ./reportfeed 
	github.com/goblimey/go-tools/statusreporter => ../../pkg/statusreporter
	github.com/goblimey/go-tools/switchwriter => ../../pkg/switchwriter
	github.com/goblimey/go-tools/testsupport => ../../pkg/testsupport
)

go 1.16
