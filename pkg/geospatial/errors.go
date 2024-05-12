package geospatial

import "fmt"

const (
	RingTooShort         = "polygon ring too short, must contain at least 4 positions"
	RingUnclosed         = "polygon ring must be closed, first and last positions must be equal"
	RingCounterClockwise = "polygon exterior ring must be clockwise"
)

type InvalidGeometryError struct {
	msg string
}

func (e *InvalidGeometryError) Error() string {
	return e.msg
}

type RighthandRuleError struct {
	Angle float64
}

func (e *RighthandRuleError) Error() string {
	return RingCounterClockwise + fmt.Sprintf("(angle is %f)", e.Angle)
}
