package states

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aaronireland/state-server/pkg/api"
	"github.com/aaronireland/state-server/pkg/geospatial"
)

// Translates the [geospatial.State] object from the [geospatial] package into a GeoJSON feature
func NewStateResponse(state geospatial.State) api.Feature {
	return api.Feature{
		Type: "Feature",
		Geometry: api.Geometry{
			Type:        "Polygon",
			Coordinates: [][]geospatial.Coordinate{state.Border},
		},
		Properties: api.Properties{State: state.Name},
	}
}

// Adds the Bind method to the [geospatial.State] object to hook into the go-chi renderer
type CreateStateRequest geospatial.State

func (csr *CreateStateRequest) UnmarshalJSON(data []byte) error {
	required := struct {
		Name   *string             `json:"state"`
		Border *geospatial.Polygon `json:"border"`
	}{}

	if err := json.Unmarshal(data, &required); err != nil {
		return err
	} else {
		var missing []string
		if required.Name == nil {
			missing = append(missing, "name is required")
		}
		if required.Border == nil {
			missing = append(missing, "border is required")
		}
		if len(missing) > 0 {
			return fmt.Errorf("invalid json: %s", strings.Join(missing, ", "))
		}
	}
	csr.Name = *required.Name
	csr.Border = *required.Border

	return nil
}

func (sr *CreateStateRequest) Bind(r *http.Request) error {
	return nil
}

// Translates an array of [geo.State] objects into a GeoJSON FeatureCollection
func NewStateCollectionResponse(features []api.Feature) api.FeatureCollection {
	return api.FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}
