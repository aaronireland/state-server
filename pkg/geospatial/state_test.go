package geospatial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateObject(t *testing.T) {
	t.Run("clockwise border coordinates should remain unchanged", func(t *testing.T) {
		ring := []Coordinate{{-77.475793, 39.719623}, {-80.524269, 39.721209}, {-80.520592, 41.986872}, {-74.705273, 41.375059}, {-75.142901, 39.881602}, {-77.475793, 39.719623}}
		expected := State{Name: "foo", Border: ring}
		got, err := NewState("foo", ring)

		assert.Nil(t, err, "given ring should produce a valid State")
		assert.Equal(t, expected.Name, got.Name, "constructor should not modify name")
		assert.Equal(t, len(expected.Border), len(got.Border), "constructor should not modify ring")

		for i := 0; i < len(ring); i++ {
			assert.Equal(t, expected.Border[i].Lng, got.Border[i].Lng, "coordinates should match in both State objects")
			assert.Equal(t, expected.Border[i].Lat, got.Border[i].Lat, "coordinates should match in both State objects")
		}
	})

	t.Run("counter-clockwise border coordinates should be reversed", func(t *testing.T) {

		ring := []Coordinate{{0, 0}, {1, 1}, {2, 2}, {0, 0}}
		expected := State{Name: "foo", Border: ring}
		got, err := NewState("foo", ring)

		assert.Nil(t, err, "given ring should produce a valid State")
		assert.Equal(t, expected.Name, got.Name, "constructor should not modify name")
		assert.Equal(t, len(expected.Border), len(got.Border), "constructor should not modify ring")

		for i := 0; i < len(ring); i++ {
			assert.Equalf(t, expected.Border[len(ring)-(i+1)].Lng, got.Border[i].Lng, "coordinates should be reversed: %s", got.Border.String())
			assert.Equalf(t, expected.Border[len(ring)-(i+1)].Lat, got.Border[i].Lat, "coordinates should be reversed %s", got.Border.String())
		}
	})
}

func TestInvalidStateObject(t *testing.T) {
	validRing := []Coordinate{{0, 0}, {2, 2}, {1, 1}, {0, 0}}
	unclosedRing := []Coordinate{{0, 0}, {1, 1}, {2, 2}, {1, 1}}
	line := []Coordinate{{1.2345, -1.2345}, {2.5, 9.00001}}

	t.Run("invalid state name given", func(t *testing.T) {
		invalid := []struct {
			Name   string
			Border []Coordinate
		}{
			{Name: "", Border: validRing},
			{Name: "A", Border: validRing},
		}

		for _, s := range invalid {
			_, err := NewState(s.Name, s.Border)
			assert.NotNilf(t, err, "constructor should return error if invalid name given: %s", s.Name)
		}
	})

	t.Run("invalid border coordinates given", func(t *testing.T) {
		invalid := []struct {
			Name   string
			Border []Coordinate
		}{
			{Name: "Line", Border: line},
			{Name: "Unclosed Ring", Border: unclosedRing},
		}

		for _, s := range invalid {
			_, err := NewState(s.Name, s.Border)
			assert.NotNilf(t, err, "constructor should return error if invalid name given: %s", s.Name)
		}

	})
}

func TestStateContains(t *testing.T) {
	preciousPetals := LatLng(40.152404, -75.039853)
	usSupplyCompany := LatLng(40.162555, -75.062416)

	pa, err := NewState("Pennsylvania", []Coordinate{
		{-77.475793, 39.719623}, {-80.524269, 39.721209}, {-80.520592, 41.986872},
		{-74.705273, 41.375059}, {-75.142901, 39.881602}, {-77.475793, 39.719623},
	})

	assert.Nil(t, err, "given ring should produce a valid State")
	assert.False(t, pa.Contains(preciousPetals), "location not contained in given state border")
	assert.True(t, pa.Contains(usSupplyCompany), "location is contained in given state border")

}
