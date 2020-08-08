module gtihub.com/goblimey/go-tools/dailylogger

go 1.14

require (
	github.com/goblimey/go-tools/clock v0.0.0-00010101000000-000000000000
	github.com/goblimey/go-tools/switchwriter v0.0.0-00010101000000-000000000000
)

replace (
	github.com/goblimey/go-tools/clock => ../clock
	github.com/goblimey/go-tools/switchwriter => ../switchwriter
)
