module github.com/goblimey/go-tools

go 1.14

require (
	github.com/goblimey/go-tools/clock v0.0.0-00010101000000-000000000000 // indirect
	github.com/goblimey/go-tools/dailylogger v0.0.0-00010101000000-000000000000 // indirect
	github.com/goblimey/go-tools/statusreporter v0.0.0-00010101000000-000000000000 // indirect
	github.com/goblimey/go-tools/switchwriter v0.0.0-00010101000000-000000000000 // indirect
	github.com/goblimey/go-tools/logger v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	github.com/goblimey/go-tools/clock => ./pkg/clock
	github.com/goblimey/go-tools/dailylogger => ./pkg/dailylogger
	github.com/goblimey/go-tools/statusreporter => ./pkg/statusreporter
	github.com/goblimey/go-tools/switchwriter => ./pkg/switchwriter
	github.com/goblimey/go-tools/logger => ./pkg/logger
)
