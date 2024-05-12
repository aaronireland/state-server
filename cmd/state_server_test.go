package cmd

import (
	"testing"

	"github.com/aaronireland/state-server/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	gotCommand, gotAction, gotBackgroundProcess := parseArgs("test")
	assert.Equal(t, "test", gotCommand, "should return the first arg")
	assert.Equal(t, server.Start, gotAction, "default action is start")
	assert.False(t, gotBackgroundProcess, "default is no background process")

	gotCommand, gotAction, gotBackgroundProcess = parseArgs("test", "foo")
	assert.Equal(t, "test", gotCommand, "should return the first arg")
	assert.Equal(t, server.Start, gotAction, "default action is start")
	assert.False(t, gotBackgroundProcess, "default is no background process")

	gotCommand, gotAction, gotBackgroundProcess = parseArgs("test", "start")
	assert.Equal(t, "test", gotCommand, "should return the first arg")
	assert.Equal(t, server.Start, gotAction, "action is start")
	assert.True(t, gotBackgroundProcess, "background process")

	gotCommand, gotAction, gotBackgroundProcess = parseArgs("test", "stop")
	assert.Equal(t, "test", gotCommand, "should return the first arg")
	assert.Equal(t, server.Shutdown, gotAction, "action is start")
	assert.True(t, gotBackgroundProcess, "background process")
}
