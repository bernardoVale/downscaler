package types

// ScaleType defines all kinds of scaling
type ScaleType int

const (
	// ScaleUp action
	ScaleUp ScaleType = iota
	// ScaleDown action
	ScaleDown
)
