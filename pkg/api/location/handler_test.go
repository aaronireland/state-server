package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aaronireland/state-server/pkg/api"
	"github.com/aaronireland/state-server/pkg/geospatial"
	"github.com/stretchr/testify/assert"
)

type mockDataProvider struct {
	Err    error
	States []geospatial.State
}

func (m mockDataProvider) GetAll() ([]geospatial.State, error) {
	return m.States, m.Err
}

func TestLocationHandler(t *testing.T) {

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

	handler := http.HandlerFunc(RouteHandler{testStore}.CheckLocationStates)

	t.Run("should return expected json for match", func(t *testing.T) {
		rr := httptest.NewRecorder()
		body := strings.NewReader("longitude=5&latitude=6")
		req, err := http.NewRequest("POST", "/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, http.StatusOK, "request should respond with 200 OK")

		var matches []string
		err = json.NewDecoder(rr.Body).Decode(&matches)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equalf(t, len(matches), 1, "expected one match, got %d", len(matches))
		assert.Equalf(t, testStore.States[0].Name, matches[0], "match in response should equal state name: %s != %s", testStore.States[0].Name, matches[0])
	})

	t.Run("should return expected json for no match", func(t *testing.T) {
		rr := httptest.NewRecorder()
		body := strings.NewReader("longitude=-5&latitude=6")
		req, err := http.NewRequest("POST", "/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code, "request should respond with 404 Not Found")

		var errResp api.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equalf(t, "Not Found", errResp.StatusText, "unexpected error response status: %s", errResp.StatusText)
	})
}

type badRequest int

func (badRequest) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("test a bad request")
}

func TestLocationHandlerBadRequest(t *testing.T) {
	testStore := mockDataProvider{}
	handler := http.HandlerFunc(RouteHandler{testStore}.CheckLocationStates)

	t.Run("should return BadRequestError for invalid request body", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", badRequest(0))
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "request should respond with 400 Bad Request")

		var errResp api.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equalf(t, "Bad Request", errResp.StatusText, "unexpected error response status: %s", errResp.StatusText)
		assert.Contains(t, errResp.ErrorText, "test a bad request", "error response missing expected error description")
	})

	t.Run("should return BadRequestError for invalid request parameters", func(t *testing.T) {
		rr := httptest.NewRecorder()
		body := strings.NewReader("longitude=5&latitude=6;")
		req, err := http.NewRequest("POST", "/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, http.StatusBadRequest, "request should respond with 400 Bad Request")

		var errResp api.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equalf(t, "Bad Request", errResp.StatusText, "unexpected error response status: %s", errResp.StatusText)
		assert.Contains(t, errResp.ErrorText, "invalid semicolon separator in query", "error response missing expected error description")
	})

	t.Run("should return BadRequestError for invalid longitude", func(t *testing.T) {
		rr := httptest.NewRecorder()
		body := strings.NewReader("longitude=oops&latitude=6")
		req, err := http.NewRequest("POST", "/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, http.StatusBadRequest, "request should respond with 400 Bad Request")

		var errResp api.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equalf(t, "Bad Request", errResp.StatusText, "unexpected error response status: %s", errResp.StatusText)
		assert.Contains(t, errResp.ErrorText, "invalid longitude", "error response missing expected error description")
	})

	t.Run("should return BadRequestError for invalid latitude", func(t *testing.T) {
		rr := httptest.NewRecorder()
		body := strings.NewReader("longitude=5&latitude=oops")
		req, err := http.NewRequest("POST", "/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "request should respond with 400 Bad Request")

		var errResp api.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equalf(t, "Bad Request", errResp.StatusText, "unexpected error response status: %s", errResp.StatusText)
		assert.Contains(t, errResp.ErrorText, "invalid latitude", "error response missing expected error description")
	})
}

func TestLocationHandlerBadBackend(t *testing.T) {

	testStore := mockDataProvider{
		Err: fmt.Errorf("uh-oh data store no good"),
	}
	handler := http.HandlerFunc(RouteHandler{testStore}.CheckLocationStates)
	rr := httptest.NewRecorder()

	body := strings.NewReader("longitude=5&latitude=6")
	req, err := http.NewRequest("POST", "/", body)
	assert.Nil(t, err, "should generate valid http request")

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusInternalServerError, "request should respond with 500 Internal Server Error")

	var errResp api.ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&errResp)
	assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
	assert.Equalf(t, "Internal Server Error", errResp.StatusText, "unexpected error response status: %s", errResp.StatusText)
	assert.Equal(t, errResp.ErrorText, "uh-oh data store no good", "error response missing expected error description")
}
