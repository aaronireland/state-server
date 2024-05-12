package geospatial

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	t.Run("Should correctly identify points inside a shape", func(t *testing.T) {
		square, err := NewPolygon([]Coordinate{
			{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0},
		})

		assert.Nil(t, err, "sqaure should be a valid polygon")

		insideCoordinates := []Coordinate{
			{5, 5}, {5, 8}, {10, 5}, {8, 5}, {1, 2}, {2, 1},
		}
		for _, coord := range insideCoordinates {
			assert.Truef(t, square.Contains(coord), "%s should be inside square bounded by: %s", coord.String(), square.String())
		}

		outsideCoordinates := []Coordinate{
			{-10, 5}, {0, 5}, {10, 10}, {100, 5}, {100, 100}, {-100, -50}, {5, -11}, {5, 11},
		}
		for _, coord := range outsideCoordinates {
			assert.Falsef(t, square.Contains(coord), "%s should be outside square bounded by: %s", coord.String(), square.String())
		}

		pt := Coordinate{1, 5}
		e := edge{p1: Coordinate{15, 2}, p2: Coordinate{10, 10}}
		intersects := rayIntersectsEdge(pt, e)
		assert.True(t, intersects, "ray from point intersects edge with negative slope")

		pt = Coordinate{100, 5}
		intersects = rayIntersectsEdge(pt, e)
		assert.False(t, intersects, "ray from point does not intersect edge with negative slope")
	})
}

func TestPolygonValidate(t *testing.T) {
	validRing := []Coordinate{{0, 0}, {2, 2}, {1, 1}, {0, 0}}
	unclosedRing := []Coordinate{{0, 0}, {2, 2}, {1, 1}, {0.5, 0}}
	line := []Coordinate{{1.2345, -1.2345}, {2.5, 9.00001}}
	counterClockwise := []Coordinate{{0, 0}, {1, 1}, {2, 2}, {0, 0}}

	err := Polygon(validRing).Validate()
	assert.Nil(t, err, "expect valid polygon")
	_, err = NewPolygon(validRing)
	assert.Nil(t, err, "expect valid polygon")

	err = Polygon(unclosedRing).Validate()
	var invalidGeoErr *InvalidGeometryError
	assert.NotNil(t, err, "unclosed ring should not validate")
	assert.Equal(t, err.Error(), RingUnclosed, "expecting RingUnclosed validation error")
	assert.True(t, errors.As(err, &invalidGeoErr), "expecting error to be InvalidGeometryError")

	_, err = NewPolygon(unclosedRing)
	assert.NotNil(t, err, "unclosed ring should not validate")
	assert.Equal(t, err.Error(), RingUnclosed, "expecting RingUnclosed validation error")
	assert.True(t, errors.As(err, &invalidGeoErr), "expecting error to be InvalidGeometryError")

	err = Polygon(line).Validate()
	assert.NotNil(t, err, "line should not validate")
	assert.Equal(t, err.Error(), RingTooShort, "expecting RingTooShort validation error")
	assert.True(t, errors.As(err, &invalidGeoErr), "expecting error to be InvalidGeometryError")

	_, err = NewPolygon(line)
	assert.NotNil(t, err, "line should not validate")
	assert.Equal(t, err.Error(), RingTooShort, "expecting RingTooShort validation error")
	assert.True(t, errors.As(err, &invalidGeoErr), "expecting error to be InvalidGeometryError")

	err = Polygon(counterClockwise).Validate()
	var righthandErr *RighthandRuleError
	assert.NotNil(t, err, "counter-clockwise ring should not validate")
	assert.True(t, errors.As(err, &righthandErr), "expecting error to be RighthandRuleError")
	assert.Contains(t, err.Error(), RingCounterClockwise, "expecting RingCounterClockwise error message")

	got, err := NewPolygon(counterClockwise)
	assert.Nil(t, err, "constructor should reverse a counter-clockwise ring")
	assert.NotNil(t, got, "constructor should produce a valid polygon from a counter-clockwise ring")
}

func TestPolygonJSON(t *testing.T) {
	t.Run("should produce expected json and unmarshal back with same coordinates", func(t *testing.T) {
		expected := "[[0,0],[2,2],[1,1],[0,0]]"
		p := Polygon([]Coordinate{{0, 0}, {2, 2}, {1, 1}, {0, 0}})
		got, err := p.MarshalJSON()
		assert.Nil(t, err, "expect valid json from jsonMarshal")
		assert.Equal(t, expected, string(got), "expect json to marshal into array of coordinates")

		var gotP Polygon
		err = gotP.UnmarshalJSON(got)
		assert.Nil(t, err, "expect json to unmarshal back into polygon")
		assert.ElementsMatch(t, p, gotP, "unmarshalled json should contain the same coordinates")
	})

	t.Run("invalid json should produce expected errors", func(t *testing.T) {
		var p Polygon

		invalidCoordinatesJSON := `[[0,2,3],[0,2]]`
		err := p.UnmarshalJSON([]byte(invalidCoordinatesJSON))
		assert.NotNil(t, err, "expect only 2-dimensional coordinates in polygon")

		invalidRingJSON := `[[0,2],[0,2]]`
		err = p.UnmarshalJSON([]byte(invalidRingJSON))
		assert.NotNil(t, err, "expect a valid ring in polygon")

		invalidJSON := "ceci n'est pas un json"
		err = p.UnmarshalJSON([]byte(invalidJSON))
		assert.NotNil(t, err, "expect UnmarshalJSON to produce error for invalid json")
	})

}
