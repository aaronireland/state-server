package states

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aaronireland/state-server/pkg/api"
	"github.com/aaronireland/state-server/pkg/api/backend"
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

func (m mockDataProvider) GetByName(name string) (geospatial.State, error) {
	if m.Err != nil {
		return geospatial.State{}, m.Err
	}
	return m.States[0], nil
}

func (m mockDataProvider) Create(state geospatial.State) (geospatial.State, error) {
	if m.Err != nil {
		return geospatial.State{}, m.Err
	}
	return state, nil
}

func (m mockDataProvider) Delete(name string) error {
	return m.Err
}

func TestGetStateHandler(t *testing.T) {

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

	handler := http.HandlerFunc(RouteHandler{testStore}.GetState)

	t.Run("should render valid RFC 7946 Feature", func(t *testing.T) {

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/api/v1/state/square", nil)
		assert.Nil(t, err, "should generate valid http request")

		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, http.StatusOK, "request should respond with 200 OK")

		var feature api.Feature
		err = json.NewDecoder(rr.Body).Decode(&feature)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equal(t, "Feature", feature.Type, "response should be a valid RFC 7946 JSON type")
		assert.Equal(t, "square", feature.Properties.State, "response should be valid RFC 7946 JSON with the state name in the properties object")
		assert.ElementsMatch(t, testStore.States[0].Border, feature.Geometry.Coordinates[0], "state border coordinates should match the first ring of the RFC 7946 Polygon geometry")
	})

	t.Run("should return expected json for no state found", func(t *testing.T) {
		notFoundErrorStore := mockDataProvider{
			Err: &backend.StateNotFoundError{Name: "nunavut"},
		}
		handler := http.HandlerFunc(RouteHandler{notFoundErrorStore}.GetState)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/api/v1/state/nunavut", nil)
		assert.Nil(t, err, "should generate valid http request")

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code, "request should respond with 404 Not Found")

		var errResp api.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equalf(t, "Not Found", errResp.StatusText, "unexpected error response status: %s", errResp.StatusText)
		assert.Contains(t, errResp.ErrorText, "nunavut", "not found error should confirm the name it searched")
	})
}

func TestCreateStateHandler(t *testing.T) {

	testStatePayload := `{
		"state": "Washington",
		"border": [
			[-122.402015, 48.225216],
			[-117.032049, 48.999931],
			[-116.919132, 45.995175],
			[-124.079107, 46.267259],
			[-124.717175, 48.377557],
			[-122.92315, 47.047963],
			[-122.402015, 48.225216]
		]
	}`

	t.Run("should render expected json for state", func(t *testing.T) {
		testStore := mockDataProvider{}
		handler := http.HandlerFunc(RouteHandler{testStore}.CreateState)
		rr := httptest.NewRecorder()

		body := strings.NewReader(testStatePayload)
		req, err := http.NewRequest("POST", "/api/v1/state/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code, "request should respond with 201 Created")

		var feature api.Feature
		err = json.NewDecoder(rr.Body).Decode(&feature)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equal(t, "Washington", feature.Properties.State, "response should be valid RFC 7946 JSON with the state name in the properties object")
		assert.Equal(t, 1, len(feature.Geometry.Coordinates), "response object should have a valid polygon")
		assert.GreaterOrEqual(t, len(feature.Geometry.Coordinates[0]), 4, "response object should have a valid polygon")
	})
}

func TestCreateStateHandlerBadRequest(t *testing.T) {
	testStore := mockDataProvider{
		Err: &backend.InvalidStateError{Err: fmt.Errorf("test bad request")},
	}
	handler := http.HandlerFunc(RouteHandler{testStore}.CreateState)

	t.Run("incorrect json schema", func(t *testing.T) {
		testStatePayload := `{"key": "test", "value": "bad request"}`

		body := strings.NewReader(testStatePayload)
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/api/v1/state/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "request should respond with 400 Bad Request")

		var resp api.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&resp)
		assert.Nil(t, err, "bad request should produce valid error response json")
		assert.Contains(t, resp.ErrorText, "name is required", "error response should indicate which fields are missing")
		assert.Contains(t, resp.ErrorText, "border is required", "error response should indicate which fields are missing")

	})

	t.Run("invalid json data in valid schema", func(t *testing.T) {
		testStatePayload := `{"name": "test", "border": []}`

		body := strings.NewReader(testStatePayload)
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/api/v1/state/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "request should respond with 400 Bad Request")

		var resp api.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&resp)
		assert.Nil(t, err, "bad request should produce valid error response json")
		assert.Equal(t, resp.ErrorText, geospatial.RingTooShort, "error response should indicate invalid polygon coordinates given")
	})

	t.Run("incorrect backend schema", func(t *testing.T) {
		testStatePayload := `{
			"state": "Washington",
			"border": [
				[-122.402015, 48.225216],
				[-117.032049, 48.999931],
				[-116.919132, 45.995175],
				[-124.079107, 46.267259],
				[-124.717175, 48.377557],
				[-122.92315, 47.047963],
				[-122.402015, 48.225216]
			]
		}`

		body := strings.NewReader(testStatePayload)
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/api/v1/state/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "request should respond with 400 Bad Request")
	})
}

func TestDeleteStateHandler(t *testing.T) {
	testStore := mockDataProvider{}
	handler := http.HandlerFunc(RouteHandler{testStore}.DeleteState)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/api/v1/state/foo", nil)
	assert.Nil(t, err, "should generate valid http request")

	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK, "request should respond with 200 OK")
}

func TestListStatesHandler(t *testing.T) {
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

	t.Run("should render valid RFC 7946 FeatureCollection", func(t *testing.T) {

		handler := http.HandlerFunc(RouteHandler{testStore}.ListStates)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/api/v1/state/", nil)
		assert.Nil(t, err, "should generate valid http request")
		handler.ServeHTTP(rr, req)
		assert.Equal(t, rr.Code, http.StatusOK, "request should respond with 200 OK")

		var collection api.FeatureCollection
		err = json.NewDecoder(rr.Body).Decode(&collection)
		assert.Nilf(t, err, "should be able to decode response json, got error: %s", err)
		assert.Equal(t, "FeatureCollection", collection.Type, "response should be a valid RFC 7946 JSON type")
		assert.Equal(t, 1, len(collection.Features), "response should contain the state object in the mock data store")
		assert.Equal(t, "square", collection.Features[0].Properties.State, "response should be valid RFC 7946 JSON with the state name in the properties object")
		assert.ElementsMatch(t, testStore.States[0].Border, collection.Features[0].Geometry.Coordinates[0], "response should contain the state boundary coordinates")
	})
}

func TestStateHandlerInternalServerError(t *testing.T) {
	testStore := mockDataProvider{
		Err: fmt.Errorf("test internal server error"),
	}

	testStatePayload := `{
		"state": "Washington",
		"border": [
			[-122.402015, 48.225216],
			[-117.032049, 48.999931],
			[-116.919132, 45.995175],
			[-124.079107, 46.267259],
			[-124.717175, 48.377557],
			[-122.92315, 47.047963],
			[-122.402015, 48.225216]
		]
	}`

	t.Run("should render internal server error for backend error on GET", func(t *testing.T) {
		handler := http.HandlerFunc(RouteHandler{testStore}.GetState)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/api/v1/state/", nil)
		assert.Nil(t, err, "should generate valid http request")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code, "request should respond with 500 Internal Server Error")
	})

	t.Run("should render internal server error for backend error on POST", func(t *testing.T) {
		handler := http.HandlerFunc(RouteHandler{testStore}.CreateState)

		body := strings.NewReader(testStatePayload)
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/api/v1/state/", body)
		assert.Nil(t, err, "should generate valid http request")

		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code, "request should respond with 500 Internal Server Error")
	})

	t.Run("should render internal server error for backend error on DELETE", func(t *testing.T) {
		handler := http.HandlerFunc(RouteHandler{testStore}.DeleteState)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", "/api/v1/state/foo", nil)
		assert.Nil(t, err, "should generate valid http request")

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code, "request should respond with 500 Internal Server Error")
	})

	t.Run("should render internal server error for backend error on GET all states", func(t *testing.T) {
		handler := http.HandlerFunc(RouteHandler{testStore}.ListStates)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", "/api/v1/state/", nil)
		assert.Nil(t, err, "should generate valid http request")

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code, "request should respond with 500 Internal Server Error")
	})
}
