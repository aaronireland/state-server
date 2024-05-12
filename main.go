package main

import (
	"fmt"
	"os"

	"github.com/aaronireland/state-server/cmd"
)

func main() {
	args := os.Args
	if err := cmd.StateServer(args...); err != nil {
		fmt.Printf("State API Server failed: %s\n", err)
		os.Exit(1)
	}
}
