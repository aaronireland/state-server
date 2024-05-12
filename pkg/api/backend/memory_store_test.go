package backend

import (
	"errors"
	"strings"
	"testing"

	"github.com/aaronireland/state-server/pkg/geospatial"
	"github.com/stretchr/testify/assert"
)

func TestStatesMemoryStore(t *testing.T) {

	validState, err := geospatial.NewState(
		"Valid",
		[]geospatial.Coordinate{
			{Lng: float64(-122.402015), Lat: float64(48.225216)},
			{Lng: float64(-117.032049), Lat: float64(48.999931)},
			{Lng: float64(-116.919132), Lat: float64(45.995175)},
			{Lng: float64(-124.079107), Lat: float64(46.267259)},
			{Lng: float64(-124.717175), Lat: float64(48.377557)},
			{Lng: float64(-122.92315), Lat: float64(47.047963)},
			{Lng: float64(-122.402015), Lat: float64(48.225216)},
		},
	)

	assert.Nil(t, err, "given coordinates should produce a valid state")

	invalidState := geospatial.State{
		Name: "Unclosed Ring",
		Border: []geospatial.Coordinate{
			{Lng: float64(-122.402015), Lat: float64(48.225216)},
			{Lng: float64(-117.032049), Lat: float64(48.999931)},
			{Lng: float64(-116.919132), Lat: float64(45.995175)},
			{Lng: float64(-124.079107), Lat: float64(46.267259)},
			{Lng: float64(-124.717175), Lat: float64(48.377557)},
			{Lng: float64(-122.92315), Lat: float64(47.047963)},
		},
	}

	t.Run("should create valid geospatial.State object that is accessible through StateLocationMemoryStore functions", func(t *testing.T) {
		s := NewMemoryStore()
		assert.Equal(t, 0, len(s.states), "data store should contain no states")
		created, err := s.Create(validState)
		assert.Nil(t, err, "data store should add valid state with no errors")
		assert.Equal(t, validState.Name, created.Name, "title-cased state name should match the state object appended")
		assert.ElementsMatch(t, validState.Border, created.Border, "the coordinates for the state in the data store should match what was given")
		assert.Equal(t, 1, len(s.states), "data should should contain one state")

		states, err := s.GetAll()
		assert.Nil(t, err, "data store GetAll function should not produce any errors")
		assert.Equal(t, 1, len(states), "data store GetAll function should return the correct number of states")

		for _, name := range []string{"valid", "VALID", "Valid", "vALiD"} {
			got, err := s.GetByName(name)
			assert.Nil(t, err, "GetByName should be case-insensitive")
			assert.Equal(t, created.Name, got.Name, "GetByName should return title-cased name")
		}

		for _, name := range []string{"", ".valid", "Valid ", "not here"} {
			_, err := s.GetByName(name)
			assert.NotNil(t, err, "GetByName should produce an error if state name not in data store")
			var notFoundError *StateNotFoundError
			assert.True(t, errors.As(err, &notFoundError), "GetByName should return StateNotFound error if state name not in data store")
			assert.Contains(t, err.Error(), "no state found", "GetByName should produce descriptive error message")
		}

		err = s.Delete(strings.ToLower(created.Name))
		assert.Nil(t, err, "Delete should not produce an error")
		assert.Equal(t, 0, len(s.states), "data store should contain no states")
		states, err = s.GetAll()
		assert.Nil(t, err, "GetAll should not produce any errors if data store is empty")
		assert.Equal(t, 0, len(states), "GetAll should return empty array if data store is empty")
	})

	t.Run("should produce InvalidStateError for request to add invalid geospatial.State object", func(t *testing.T) {
		s := NewMemoryStore()

		assert.Equal(t, 0, len(s.states), "data store should contain no states")
		_, err := s.Create(invalidState)
		assert.NotNil(t, err, "Create should produce InvalidStateError if invalid geospatial.State is given")
		assert.Equal(t, 0, len(s.states), "data should should contain no states after attempt to add invalid geospatial.State")
	})

	t.Run("should produce InvalidStateError for request to add existing geospatial.State object", func(t *testing.T) {
		s := NewMemoryStore()

		assert.Equal(t, 0, len(s.states), "data store should contain no states")
		created, err := s.Create(validState)
		assert.Nil(t, err, "data store should add valid state with no errors")
		assert.Equal(t, 1, len(s.states), "data should should contain one state")

		_, err = s.Create(created)
		assert.NotNil(t, err, "Create should produce InvalidStateError for attempt to create existing state")
		assert.Contains(t, err.Error(), "duplicate", "attempt to add duplicate state should produce informative error message")
	})
}
