package models

// Dimension represents a measurement with magnitude and unit
type Dimension struct {
	Magnitude float64 `json:"magnitude"`
	Unit      string  `json:"unit"`
}

// Size represents a size dimension
type Size struct {
	Height Dimension `json:"height"`
	Width  Dimension `json:"width"`
}
