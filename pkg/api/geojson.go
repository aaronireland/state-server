// Package api provides the common resources for go chi API handlers and routers
// for the State Server API, namely error responses and RFC 7946 GeoJSON schema
// structs for serializing and deserializing JSON
package api

import (
	"net/http"

	"github.com/aaronireland/state-server/pkg/geospatial"
)

// GeoJSON schema for the geometry object of a feature
type Geometry struct {
	Type        string                    `json:"type"`
	Coordinates [][]geospatial.Coordinate `json:"coordinates"`
}

// GeoJSON schema for the properties object of a feature
type Properties struct {
	State string `json:"state"`
}

// GeoJSON schema for a feature object
type Feature struct {
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
}

// GeoJSON schema for a FeatureCollection object to represent a collection of several states
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

// render method which hooks into the go-chi renderer
func (sr Feature) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (scr FeatureCollection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
