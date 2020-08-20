module github.com/goblimey/go-tools/proxy

go 1.14

require (
	github.com/goblimey/go-tools/logger v0.0.4
	github.com/goblimey/go-tools/proxy/reportfeed v0.0.4
	github.com/goblimey/go-tools/statusreporter v0.0.4
)

replace (
    github.com/goblimey/go-tools/logger => ../logger
    github.com/goblimey/go-tools/reportfeed => ../reportfeed
    github.com/goblimey/go-tools/statusreporter => ../statusreporter
)
