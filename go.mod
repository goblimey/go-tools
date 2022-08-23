module github.com/goblimey/go-tools

go 1.14

replace (
	github.com/goblimey/go-tools/clock => ./clock
	github.com/goblimey/go-tools/dailylogger => ./dailylogger
	github.com/goblimey/go-tools/logger => ./logger
	github.com/goblimey/go-tools/switchwriter => ./switchwriter
	github.com/goblimey/go-tools/proxy/reportfeed => ./proxy/reportfeed
)
