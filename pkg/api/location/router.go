// Package location provides the router and handler for the
// State Server endpoint which finds the state(s) in which
// a given location is contained
package location

import (
	"github.com/go-chi/chi/v5"

	"github.com/aaronireland/state-server/pkg/geospatial"
)

// Injects the dependcies required by the handler for the
// backend data store
type DataProvider interface {
	GetAll() ([]geospatial.State, error)
}

// Maps the handler to required REST API endpoint
func Router(store DataProvider) chi.Router {
	router := chi.NewRouter()
	handler := RouteHandler{store}

	router.Post("/", handler.CheckLocationStates)

	return router
}
