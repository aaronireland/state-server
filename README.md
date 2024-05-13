# State Server
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Coverage](https://github.com/aaronireland/state-server/wiki/coverage.svg?branch=main)](https://raw.githack.com/wiki/aaronireland/state-server/coverage.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/aaronireland/state-server?branch=main)](https://goreportcard.com/report/github.com/aaronireland/state-server)

A simple REST API written in Go which maintains a collection of geographic representations of several US States (very roughly approximated) with the ability to
determine if a geographic point (consisting of a latitude and longitude coordinate) is contained with one (or more) states. This logic is implemented using the [Ray-casting algorithm](https://rosettacode.org/wiki/Ray-casting_algorithm)

The API server maintains an in-memory data store of state location data and contains RESTful endpoints for CRUD operations. The location data renders as geospatial JSON objects formatted according to [RFC 7946 - GeoJSON](https://datatracker.ietf.org/doc/html/rfc7946). The borders of a given state are represented as a spherical polygon, and the accompanying state data (e.g. name) is stored in the feature's properties object. Each state is a `Feature` and the total collection of all the represented states is a `FeatureCollection`.

To visualize an approximated state border or a collection of state borders, you can use a web tool like [geojson.tools/](https://geojson.tools/)


## Requirements

1. [Go](https://go.dev/doc/install)
2. [Mage](https://go.dev/doc/install)


```shell
go version
```

```shell
mage -version
```


## Usage

_Note: This project was developed on a machine running `macOS 12.7.4 21H1123 x86_64`_

This project was built with Go version 1.22. The binary builds with Mage but is simple enough to 
build directly with `go` commands and/or can be run directly with `go run`, however the magefile will handle starting and stopping the server as a background process along with seeding the API data store, so using mage is recommended.

```shell
mage build
```

The binary will be build to the `bin/` directory at the project root and can be run like this:

```shell
./bin/state-server
```
... or just use mage:
```
mage server:start
```

Mage calls the binary which starts the webserver with the `start` arg which triggers the binary to check for an existing state-server process (using the [ps](https://en.wikipedia.org/wiki/Ps_(Unix)) utility) and terminate it before starting as a subprocess.

The webserver creates a lock file in the directory where the running binary exists (e.g. the bin directory). The `Start` method checks for an existing lock file and fails if one exists. The lockfile is deleted when the webserver terminates. The `clean` target of the magefile uses the lockfile to know if it should stop a running webserver process before removing the bin directory.


To seed the API data store with the example states collection provided in [assets/states.json](assets/states.json):
```shell
mage server:seed states.json
```

To stop the server using mage:
```shell
mage server:stop
```

### Example Requests

 Get the state(s), if any, in which a location exists:

```shell
curl  -d "longitude=-77.036133&latitude=40.513799" http://localhost:8080/
```
outputs: `["Pennsylvania"]`

```
curl  -d "longitude=-77.036133&latitude=45" http://localhost:8080/

```
outputs: `{"status":"Not Found","error":"[-77.036133, 45] not within any state"}`

Get the GeoJSON Feature object which contains the location data for Pennsylvania

```shell
curl http://localhost:8080/api/v1/state/pennsylvania
```
outputs (truncated): `"type":"Feature","properties":{"state":"Pennsylvania"},"geometry":{"type":"Polygon","coordinates":[[[-77.475793,39.719623],..., ]]}}`



## Testing

To run tests with mage: 

```shell
mage test
```

or with the html coverage:

```shell
mage coverage
```

## Documentation


* [geospatial](https://aaronireland.github.io/state-server/pkg/geospatial/doc.html)

* [state-server](https://aaronireland.github.io/state-server/pkg/server/doc.html)
* [state-server location api](https://aaronireland.github.io/state-server/pkg/api/location/doc.html)
* [state-server states api](https://aaronireland.github.io/state-server/pkg/api/states/doc.html)
* [state-server api backend](https://aaronireland.github.io/state-server/pkg/api/backend/doc.html)
