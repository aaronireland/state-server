// Package state-server is the State Server REST API HTTP server. It uses the go-chi framework to handle
// http requests to get the state (or states) in which a location is contained, render
// the state location data as JSON in the [RFC 7946 GeoJSON] format, render the full
// collection of states, create a state, or delete a state.
//
// # Example requests
//
// Get the state(s), if any, in which a location exists:
//
//	$ curl  -d "longitude=-77.036133&latitude=40.513799" http://localhost:8080/
//	["Pennsylvania"]
//
//	$ curl  -d "longitude=-77.036133&latitude=45" http://localhost:8080/
//	{"status":"Not Found","error":"[-77.036133, 45] not within any state"}
//
// Get the GeoJSON Feature object which contains the location data for Pennsylvania
//
//	$ curl http://localhost:8080/api/v1/state/pennsylvania
//	"type":"Feature","properties":{"state":"Pennsylvania"},"geometry":{"type":"Polygon","coordinates":[[[-77.475793,39.719623],..., ]]}}
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/aaronireland/state-server/pkg/geospatial"
)

const (
	Start    Action = "start"
	Shutdown Action = "shutdown"
)

var lockFile = "server.lock"

func init() {
	if binPath, err := os.Executable(); err == nil {
		lockFile = filepath.Join(filepath.Dir(binPath), lockFile)
	}
}

type Action string
type StateLocationDataProvider interface {
	GetAll() ([]geospatial.State, error)
	GetByName(name string) (geospatial.State, error)
	Create(geospatial.State) (geospatial.State, error)
	Delete(name string) error
}

type StateServer struct {
	router *chi.Mux
	config serverConfig
}

func NewStateServer(config serverConfig, store StateLocationDataProvider) *StateServer {
	return &StateServer{
		config: config,
		router: StateServerAPIRouter(store),
	}
}

func (s *StateServer) Start(ctx context.Context) error {
	lock()
	defer unlock()

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      s.router,
		IdleTimeout:  s.config.IdleTimeout,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	var shutdownErr error
	shutdownComplete := handleShutdown(func() {
		if err := server.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("server.Shutdown failed: %w", err)
		}
	})

	if err := server.ListenAndServe(); err == http.ErrServerClosed {
		<-shutdownComplete
	} else {
		shutdownErr = fmt.Errorf("http.ListenAndServe failed: %w", err)
	}

	if shutdownErr != nil {
		return shutdownErr
	}

	return nil
}

func Abort(action Action, msg string) {
	fmt.Fprintf(os.Stderr, "\nserver failed to %s: %s\n", action, msg)
	os.Exit(1)
}

func handleShutdown(onShutdownSignal func()) <-chan struct{} {
	shutdown := make(chan struct{})

	go func() {
		shutdownSignal := make(chan os.Signal, 1)
		signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

		<-shutdownSignal

		onShutdownSignal()
		close(shutdown)
	}()

	return shutdown
}

func lock() {
	if _, err := os.Stat(lockFile); err == nil {
		Abort(Start, "server process is locked")
	}

	file, err := os.Create(lockFile)
	if err != nil {
		errMsg := fmt.Sprintf("unable to create %s -> %s", lockFile, err)
		Abort(Start, errMsg)
	}
	file.Close()
}

func unlock() {
	err := os.Remove(lockFile)
	if err != nil {
		errMsg := fmt.Sprintf("unable to remove %s -> %s", lockFile, err)
		Abort(Shutdown, errMsg)
	}
}
