package types

import "time"

type exchangeDataType struct {
	SourceSymbol string
	SourceId     int64
	SourceAmount float64
	TargetSymbol string
	TargetId     int64
	TargetAmount float64
	Timestamp    time.Time
}
