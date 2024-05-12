package states

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/aaronireland/state-server/pkg/api"
	"github.com/aaronireland/state-server/pkg/api/backend"
	"github.com/aaronireland/state-server/pkg/geospatial"
)

type RouteHandler struct {
	store DataProvider
}

// HTTP request handler for the /api/v1/state/{name} endpoint
func (h RouteHandler) GetState(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	state, err := h.store.GetByName(name)
	if err != nil {
		var notFoundErr *backend.StateNotFoundError
		if errors.As(err, &notFoundErr) {
			render.Render(w, r, api.NotFoundError(err))
		} else {
			render.Render(w, r, api.InternalServerError(err))
		}
		return
	}

	render.Render(w, r, NewStateResponse(state))
}

// HTTP  request handler for the POST /api/v1/state endpoint creates the [geospatialspatial.State] object and
// adds it to the data store
func (h RouteHandler) CreateState(w http.ResponseWriter, r *http.Request) {
	state := &CreateStateRequest{}
	if err := render.Bind(r, state); err != nil {
		render.Render(w, r, api.BadRequestError(err))
		return
	}

	created, err := h.store.Create(geospatial.State(*state))
	if err != nil {
		var invalidStateErr *backend.InvalidStateError
		if errors.As(err, &invalidStateErr) {
			render.Render(w, r, api.BadRequestError(err))
		} else {
			render.Render(w, r, api.InternalServerError(err))
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Render(w, r, NewStateResponse(created))

}

// HTTP request handler for the DELETE /api/v1/state/{name} endpoint removes a given state from
// the data store
func (h RouteHandler) DeleteState(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if err := h.store.Delete(name); err != nil {
		render.Render(w, r, api.InternalServerError(err))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(nil)

}

// HTTP request handler for the GET /api/v1/state endpoint renders the entire list of states
// in the data store to a GeoJSON feature collection
func (h RouteHandler) ListStates(w http.ResponseWriter, r *http.Request) {
	states, err := h.store.GetAll()
	if err != nil {
		render.Render(w, r, api.InternalServerError(err))
		return
	}

	var features []api.Feature
	for _, state := range states {
		features = append(features, NewStateResponse(state))
	}

	render.Render(w, r, NewStateCollectionResponse(features))

}
