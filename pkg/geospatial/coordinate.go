package geospatial

import (
	"encoding/json"
	"fmt"
)

// Creates the expected [Coordinate] object
// with ordered arguments by latitude and longitude
func LatLng(lat, lng float64) Coordinate {
	return Coordinate{lng, lat}
}

// Represents a single point on a map or spherical geometry
type Coordinate struct {
	Lng, Lat float64
}

// Formats the coordinate as a string
func (c Coordinate) String() string {
	return fmt.Sprintf("[%G, %G]", c.Lng, c.Lat)
}

func (c Coordinate) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{c.Lng, c.Lat})
}

func (c *Coordinate) UnmarshalJSON(data []byte) error {
	var coord []float64
	if err := json.Unmarshal(data, &coord); err != nil {
		return err
	}

	if len(coord) != 2 {
		return fmt.Errorf("invalid coordinate: expecting latitude and longitude")
	}

	*c = LatLng(coord[1], coord[0])

	return nil
}
