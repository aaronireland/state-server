package cmd

import (
	"os/exec"
	"testing"

	"github.com/aaronireland/state-server/pkg/server"
	"github.com/stretchr/testify/assert"
)

// This is more of a functional test. If the daemon were refactored to use an
// interface and dependency injection, the os/exec and filesystem could be mocked
func TestFindProcess(t *testing.T) {
	sleepCmd := exec.Command("sh", "-c", "sleep 5")
	err := sleepCmd.Start()
	assert.Nil(t, err, "sleep command should start")

	defer func() {
		sleepCmd.Process.Kill()
	}()

	gotProcs, err := findProcess("sleep 5")
	assert.Nil(t, err, "findProcess should not return an error")
	assert.Equal(t, len(gotProcs), 1, "findProcess should find the sleep command")
	assert.Equal(t, sleepCmd.Process.Pid, gotProcs[0].Pid, "findProcess should locate the correct pid")
}

func TestWaitFor(t *testing.T) {

	sleepCmd := exec.Command("sh", "-c", "sleep 6")
	err := sleepCmd.Start()
	assert.Nil(t, err, "sleep command should start")

	originalWaitFor := waitForSeconds
	waitForSeconds = 2
	defer func() {
		waitForSeconds = originalWaitFor
		sleepCmd.Process.Kill()
	}()

	shutdown := make(chan error, 1)
	go waitFor(server.Shutdown, sleepCmd.Process.Pid, shutdown)
	err = <-shutdown
	assert.NotNil(t, err, "The waitFor should return a timeout error")
	assert.Contains(t, err.Error(), "timeout", "the waitFor function should return a timeout error")
}
