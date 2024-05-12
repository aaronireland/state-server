package geospatial

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/golang/geo/s2"
)

// Provides methods for validation and analysis of the
// geospatial shape represented by the array of coordinates
type Polygon []Coordinate

// Constructs a new instance of a Polygon struct with valid
// coordinates which satisfy the RFC 7946 GeoJSON specifications
// [See: 3.1.6 Polygon](https://datatracker.ietf.org/doc/html/rfc7946#section-3.1.6)
func NewPolygon(coords []Coordinate) (*Polygon, error) {

	var p Polygon = coords
	if err := p.Validate(); err != nil {
		var invalidGeometryError *InvalidGeometryError
		if errors.As(err, &invalidGeometryError) {
			return nil, err
		}
	}
	vertices := append([]Coordinate{}, coords...)
	if turningAngle(vertices) > 0 {
		slices.Reverse(vertices)
	}
	p = vertices

	return &p, nil
}

// Formats the coordinate array so that is marshals into
// valid JSON
func (p Polygon) String() string {
	var coords []string
	for _, coord := range p {
		coords = append(coords, coord.String())
	}

	return fmt.Sprintf("{%s}", strings.Join(coords, ", "))
}

// Implements the RFC 7946 specifications for a geospatial Polygon
// [See: 3.1.6 Polygon](https://datatracker.ietf.org/doc/html/rfc7946#section-3.1.6)
func (p Polygon) Validate() error {
	if len(p) < 4 {
		return &InvalidGeometryError{RingTooShort}
	} else if p[len(p)-1] != p[0] {
		return &InvalidGeometryError{RingUnclosed}
	}

	if angle := turningAngle(p); angle > 0 { // angle greater than 0 = counter-clockwise
		return &RighthandRuleError{angle}
	}
	return nil
}

func (p Polygon) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Coordinate(p))
}

func (p *Polygon) UnmarshalJSON(data []byte) error {
	var coordinates []Coordinate
	if err := json.Unmarshal(data, &coordinates); err != nil {
		return err
	}

	if polygon, err := NewPolygon(coordinates); err != nil {
		return err
	} else {
		*p = *polygon
	}

	return nil
}

// Uses the [Ray-casting algorithm] to efficiently identify
// if the given coordinate is contained within the geospatial
// polygon. In simple terms, this algorithm creates horizontal
// lines (rays) and counts the number of intersections with the
// polygon's boundaries, an odd number of intersections indicates
// that the coordinate lies within the boundaries of the polygon.
// [Ray-casting algortithm](https://rosettacode.org/wiki/Ray-casting_algorithm)
func (p Polygon) Contains(coord Coordinate) (contains bool) {
	for i := 1; i < len(p); i++ {
		if rayIntersectsEdge(coord, edge{p[i-1], p[i]}) {
			contains = !contains
		}
	}
	return
}

// Implementation of the Ray-casting algortithm
func rayIntersectsEdge(p Coordinate, e edge) bool {
	var a, b Coordinate
	if e.p1.Lat < e.p2.Lat {
		a, b = e.p1, e.p2
	} else {
		a, b = e.p2, e.p1
	}

	for p.Lat == a.Lat || p.Lat == b.Lat {
		p.Lat = math.Nextafter(p.Lat, math.Inf(1))
	}

	if p.Lat < a.Lat || p.Lat > b.Lat {
		return false
	}

	if a.Lng > b.Lng {
		if p.Lng > a.Lng {
			return false
		}

		if p.Lng < b.Lng {
			return true
		}
	} else {
		if p.Lng > b.Lng {
			return false
		}

		if p.Lng < a.Lng {
			return true
		}
	}

	return (p.Lat-a.Lat)/(p.Lng-a.Lng) >= (b.Lat-a.Lat)/(b.Lng-a.Lng)
}

// Determines the orientation of the edges created by vertices
// represented by the coordinate array (e.g. is the shape drawn clockwise?)
func turningAngle(vertices []Coordinate) float64 {
	points := make([]s2.Point, len(vertices)-1)
	for i := 0; i < len(vertices)-1; i++ {
		points[i] = s2.PointFromLatLng(s2.LatLngFromDegrees(vertices[i].Lat, vertices[i].Lng))
	}

	return s2.LoopFromPoints(points).TurningAngle()
}

type edge struct {
	p1 Coordinate
	p2 Coordinate
}
