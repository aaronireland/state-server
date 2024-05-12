package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/aaronireland/state-server/pkg/api/location"
	"github.com/aaronireland/state-server/pkg/api/states"
)

func StateServerAPIRouter(store StateLocationDataProvider) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Mount("/", location.Router(store))
	router.Mount("/api/v1/state", states.Router(store))

	return router
}
