package geospatial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordinate(t *testing.T) {
	expected := Coordinate{-75.1, 40.2}
	got := LatLng(40.2, -75.1)

	assert.Equal(t, expected.Lat, got.Lat, "both latitude values should match")
	assert.Equal(t, expected.Lng, got.Lng, "both longitude values should match")

	assert.Equal(t, "[-75.1, 40.2]", got.String(), "expect proper string formatting for coordinate")
}

func TestCoordinateJSON(t *testing.T) {
	expected := "[-75.1,40.2]"
	test := LatLng(40.2, -75.1)
	gotBytes, err := test.MarshalJSON()
	assert.Nil(t, err, "expect MarshalJSON to produce valid JSON byte array")
	got := string(gotBytes)
	assert.Equalf(t, expected, got, "expect %s = %s", got, expected)

	var test2 Coordinate
	err = test2.UnmarshalJSON(gotBytes)
	assert.Nil(t, err, "marshalled JSON should unmarshal back to same coordinate without errors")
	assert.Equal(t, test2.Lat, test.Lat, "marshalled JSON should unmarshal back to same coordinate")
	assert.Equal(t, test2.Lng, test.Lng, "marshalled JSON should unmarshal back to same coordinate")

	invalidCoordinateJSON := `[1.1,2.2,3.3]`
	var test3 Coordinate
	err = test3.UnmarshalJSON([]byte(invalidCoordinateJSON))
	assert.NotNil(t, err, "only 2-dimensional coordinates are valid")

	invalidJSON := "ceci n'est pas un json"
	var test4 Coordinate
	err = test4.UnmarshalJSON([]byte(invalidJSON))
	assert.NotNil(t, err, "expect invalid JSON to produce and error")
}
