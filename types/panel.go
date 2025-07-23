package types

type PanelType struct {
	Source   int64   `json:"source"`
	Target   int64   `json:"target"`
	Value    float64 `json:"value"`
	Decimals int64   `json:"decimals"`
}
