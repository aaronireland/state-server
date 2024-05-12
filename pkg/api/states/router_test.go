package states

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aaronireland/state-server/pkg/geospatial"
	"github.com/stretchr/testify/assert"
)

func TestStatesRouter(t *testing.T) {

	squareState, err := geospatial.NewState(
		"square",
		[]geospatial.Coordinate{
			{Lng: float64(0), Lat: float64(0)},
			{Lng: float64(10), Lat: float64(0)},
			{Lng: float64(10), Lat: float64(10)},
			{Lng: float64(0), Lat: float64(10)},
			{Lng: float64(0), Lat: float64(0)},
		},
	)

	assert.Nil(t, err, "given coordinates should produce a valid state")
	testStore := mockDataProvider{
		States: []geospatial.State{squareState},
	}

	testRouter := Router(testStore)

	t.Run("list states path", func(t *testing.T) {
		testServer := httptest.NewServer(testRouter)
		defer testServer.Close()

		req, err := http.NewRequest("GET", testServer.URL+"/", nil)
		assert.Nil(t, err, "should be a valid request")
		resp, err := http.DefaultClient.Do(req)
		assert.Nil(t, err, "router should handle base path for states api GET")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "router should give a 200 OK response")
	})

	t.Run("get state", func(t *testing.T) {
		testServer := httptest.NewServer(testRouter)
		defer testServer.Close()

		req, err := http.NewRequest("GET", testServer.URL+"/square", nil)
		assert.Nil(t, err, "should be a valid request")
		resp, err := http.DefaultClient.Do(req)
		assert.Nil(t, err, "router should handle GET for state")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "router should give a 200 OK response")
	})

	t.Run("invalid request", func(t *testing.T) {
		testServer := httptest.NewServer(testRouter)
		defer testServer.Close()

		req, err := http.NewRequest("POST", testServer.URL+"/", nil)
		assert.Nil(t, err, "should be a valid request")
		resp, err := http.DefaultClient.Do(req)
		assert.Nil(t, err, "router should reject POST for basepath")
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "router should give a 200 OK response")
	})
}
