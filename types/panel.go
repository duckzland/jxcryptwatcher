package types

type panelType struct {
	Source       int64   `json:"source"`
	Target       int64   `json:"target"`
	Value        float64 `json:"value"`
	Decimals     int64   `json:"decimals"`
	SourceSymbol string  `json:"source_symbol"`
	TargetSymbol string  `json:"target_symbol"`
}
