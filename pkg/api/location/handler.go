package location

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aaronireland/state-server/pkg/api"
	"github.com/aaronireland/state-server/pkg/geospatial"
	"github.com/go-chi/render"
)

type RouteHandler struct {
	store DataProvider
}

// checks each [geospatial.State] object to see if the given geographic coordinate is contained within
// its borders
func getStateForLocation(states []geospatial.State, coord geospatial.Coordinate) (inStates []string) {

	for _, state := range states {
		if state.Contains(coord) {
			inStates = append(inStates, state.Name)
		}
	}

	return
}

// HTTP Request handler for the POST / endpoint which returns a list of state names or HTTP 404 error response
// for the coordinate given in the request latitude and longitude form fields
func (h RouteHandler) CheckLocationStates(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		render.Render(w, r, api.BadRequestError(err))
		return
	}

	params, err := url.ParseQuery(string(body))
	if err != nil {
		render.Render(w, r, api.BadRequestError(err))
		return
	}

	var latitude, longitude float64
	if val, ok := params["latitude"]; ok {
		if lat, err := strconv.ParseFloat(val[0], 64); err != nil {
			render.Render(w, r, api.BadRequestError(fmt.Errorf("invalid latitude: %v", val[0])))
			return
		} else {
			latitude = lat
		}
	}

	if val, ok := params["longitude"]; ok {
		if lng, err := strconv.ParseFloat(val[0], 64); err != nil {
			render.Render(w, r, api.BadRequestError(fmt.Errorf("invalid longitude: %v", val[0])))
			return
		} else {
			longitude = lng
		}
	}

	states, err := h.store.GetAll()
	if err != nil {
		render.Render(w, r, api.InternalServerError(err))
		return
	}

	coord := geospatial.LatLng(latitude, longitude)
	locationInStates := getStateForLocation(states, coord)
	if len(locationInStates) == 0 {
		render.Render(w, r, api.NotFoundError(fmt.Errorf("%s not within any state", coord.String())))
		return
	}

	render.JSON(w, r, locationInStates)
}
