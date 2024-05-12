// Package cmd starts and stops the HTTP server for the State Server API.
//
// # Usage
//
// Run the http server:
//
//	/path/to/state-server
//
// Run the http server as a backgroun process:
//
//	/path/to/state-server start
//
// Stop the server:
//
//	/path/to/state-server stop
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/aaronireland/state-server/pkg/api/backend"
	"github.com/aaronireland/state-server/pkg/server"
)

// The main command function that runs the State Server HTTP server
func StateServer(args ...string) error {
	command, action, backgroundProcess := parseArgs(args...)

	if backgroundProcess {
		return daemon(command, action)
	}

	store := backend.NewMemoryStore()
	config, err := server.LoadConfig()
	if err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Println("State API Server starting...")
	if err := server.NewStateServer(config, store).Start(ctx); err != nil {
		return err
	}

	return nil
}

func parseArgs(args ...string) (string, server.Action, bool) {
	command := args[0]

	if len(args) < 2 {
		return command, server.Start, false
	}

	switch args[1] {
	case "start":
		return command, server.Start, true
	case "stop":
		return command, server.Shutdown, true
	default:
		return command, server.Start, false
	}
}
