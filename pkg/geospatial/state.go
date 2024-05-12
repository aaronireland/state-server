// Package geospatial implements objects and functions useful for representing a geographic [State]
// as a geospatial entity and performing operations on that entity such as: visualization with
// mapping tools and efficiently determining if a geographic coordinate is contained within the
// entity's borders
package geospatial

import (
	"fmt"
)

// Represents a geographic state as a name and a
// geospatial polygon (i.e. a linear ring of boundary coordinates)
type State struct {
	Name   string  `json:"state"`
	Border Polygon `json:"border"`
}

// Constructor generates a new [State] object with the
// provided name and array of coordinates. Translates
// the coordinates into a [Polyon] object which ensures
// the state's border is representable as valid GeoJSON
func NewState(name string, coords []Coordinate) (state State, err error) {
	if len(name) < 2 {
		err = fmt.Errorf("invalid name for state, minimum length is 2: %s", name)
		return
	}

	border, err := NewPolygon(coords)
	if err != nil {
		return
	}
	state = State{Name: name, Border: *border}
	return
}

// Calls the Contains method for the [Polygon] object representing
// the state's border
func (s State) Contains(coordinate Coordinate) bool {
	return s.Border.Contains(coordinate)
}
