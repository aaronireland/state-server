// Package backend provides an in-memory thread-safe data store for creating and accessing [geospatial.State] objects.
// This data store implements the methods that satisfy the [server.StateLocationDataProvider] interface used by
// the state-server API webserver
package backend

import (
	"fmt"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/aaronireland/state-server/pkg/geospatial"
)

type StateLocationMemoryStore struct {
	states    map[string]geospatial.State
	formatter cases.Caser
	mu        sync.RWMutex
}

// Constructor for the [StateLocationMemoryStore] struct instaniates the data store
// with a string formatter for title-casing state names in a standardized manner
func NewMemoryStore() *StateLocationMemoryStore {
	formatter := cases.Title(language.AmericanEnglish)

	return &StateLocationMemoryStore{
		states:    map[string]geospatial.State{},
		formatter: formatter,
	}
}

// Gets the entire collection of [geospatial.State] objects from the data store
func (s *StateLocationMemoryStore) GetAll() ([]geospatial.State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var states []geospatial.State
	for _, state := range s.states {
		states = append(states, state)
	}
	return states, nil
}

// Gets a single [geospatial.State] object from the data store.
// Returns [StateNotFoundError] if no state exists for the given name
func (s *StateLocationMemoryStore) GetByName(name string) (geospatial.State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if state, ok := s.states[s.formatter.String(name)]; ok {
		return state, nil
	} else {
		return geospatial.State{}, &StateNotFoundError{name}
	}
}

// Validates a provided [geospatial.State] object and adds it to the data store.
// Returns [InvalidStateError] is the [geospatial.State] provided is invalid.
func (s *StateLocationMemoryStore) Create(state geospatial.State) (geospatial.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state.Name = s.formatter.String(state.Name)

	created, err := geospatial.NewState(state.Name, state.Border)
	if err != nil {
		return geospatial.State{}, &InvalidStateError{err}
	}

	if _, ok := s.states[created.Name]; ok {
		return geospatial.State{}, &InvalidStateError{fmt.Errorf("duplicate state: %s", state.Name)}
	}

	s.states[created.Name] = created

	return created, nil
}

// Removes the [geospatial.State] with the provided name from the data store collection if it exists.
func (s *StateLocationMemoryStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, s.formatter.String(name))

	return nil
}
