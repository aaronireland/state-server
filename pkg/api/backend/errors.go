package backend

import "fmt"

// Error struct which implements the Error interface and allows
// HTTP request handlers to generate the appropriate response
type StateNotFoundError struct {
	Name string
}

func (e *StateNotFoundError) Error() string {
	return fmt.Sprintf("no state found with name: %s", e.Name)
}

// Error struct which implements the Error interface and allows
// HTTP request handlers to generate the appropriate response
type InvalidStateError struct {
	Err error
}

func (e *InvalidStateError) Error() string {
	return fmt.Sprintf("%s", e.Err)
}
