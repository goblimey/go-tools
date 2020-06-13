module example.org/sketch

go 1.14

require (
	example.org/logger v0.0.0
	example.org/switchwriter v0.0.0
)

replace (
	example.org/logger => ./pkg/logger
	example.org/switchwriter => ./pkg/switchwriter
)
