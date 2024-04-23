module main.go

go 1.20

require (
	github.com/gelberg/calorie-deficit-tracker/common v0.0.0-00010101000000-000000000000
	github.com/gelberg/oauth1 v0.2.3
)

require (
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
)

replace github.com/gelberg/calorie-deficit-tracker/common => ../common
