module github.com/goblimey/go-tools/proxy/reportfeed

go 1.14

require (
    github.com/goblimey/go-tools/dailylogger v0.0.4
    github.com/goblimey/go-tools/clock v0.0.4
    github.com/goblimey/go-tools/switchwriter v0.0.4
)

replace (
    github.com/goblimey/go-tools/clock => ../clock
    github.com/goblimey/go-tools/dailylogger => ../dailylogger
    github.com/goblimey/go-tools/switchwriter => ../switchwriter
)
