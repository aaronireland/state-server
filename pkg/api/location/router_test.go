package location

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

	t.Run("get not supported", func(t *testing.T) {
		testServer := httptest.NewServer(testRouter)
		defer testServer.Close()

		req, err := http.NewRequest("GET", testServer.URL+"/", nil)
		assert.Nil(t, err, "should be a valid request")
		resp, err := http.DefaultClient.Do(req)
		assert.Nil(t, err, "router should reject GETfor basepath")
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "router should give a 405 response")
	})

	t.Run("POST request is valid", func(t *testing.T) {
		testServer := httptest.NewServer(testRouter)
		defer testServer.Close()

		req, err := http.NewRequest("POST", testServer.URL+"/", strings.NewReader("latitude=2&longitude=3"))
		assert.Nil(t, err, "should be a valid request")
		resp, err := http.DefaultClient.Do(req)
		assert.Nil(t, err, "router should reject GETfor basepath")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "router should give a 200 response")
	})
}
