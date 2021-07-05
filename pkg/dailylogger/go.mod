module github.com/goblimey/go-tools/dailylogger

require (
	github.com/goblimey/go-tools/clock v0.0.6
	github.com/goblimey/go-tools/switchwriter v0.0.6
	github.com/goblimey/go-tools/testsupport v0.0.6
)

replace (
	github.com/goblimey/go-tools/clock => ../clock
	github.com/goblimey/go-tools/switchwriter => ../switchwriter
	github.com/goblimey/go-tools/testsupport => ../testsupport
)

go 1.16
