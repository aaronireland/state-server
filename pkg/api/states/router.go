// Package states provides the REST API endpoints needed for CRUD operations on
// geospatial state objects in the backend data store
package states

import (
	"github.com/go-chi/chi/v5"

	"github.com/aaronireland/state-server/pkg/geospatial"
)

// Injects the dependcies required by the handler for the
// backend data store.
type DataProvider interface {
	GetAll() ([]geospatial.State, error)
	GetByName(name string) (geospatial.State, error)
	Create(geospatial.State) (geospatial.State, error)
	Delete(name string) error
}

// Maps the handler to the REST API endpoints for the state API
func Router(store DataProvider) chi.Router {
	router := chi.NewRouter()
	handler := RouteHandler{store}

	router.Get("/", handler.ListStates)
	router.Post("/", handler.CreateState)
	router.Get("/{name}", handler.GetState)
	router.Delete("/{name}", handler.DeleteState)

	return router
}
