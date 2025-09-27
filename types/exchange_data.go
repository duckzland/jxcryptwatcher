package types

import (
	"math/big"
	"time"
)

type exchangeDataType struct {
	SourceSymbol string
	SourceId     int64
	SourceAmount float64
	TargetSymbol string
	TargetId     int64
	TargetAmount *big.Float
	Timestamp    time.Time
}
