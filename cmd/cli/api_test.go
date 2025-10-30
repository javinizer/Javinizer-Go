package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPICmd_CommandStructure(t *testing.T) {
	cmd := newAPICmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "api", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Verify flags exist
	assert.NotNil(t, cmd.Flags().Lookup("host"))
	assert.NotNil(t, cmd.Flags().Lookup("port"))

	// Verify Run function is set
	assert.NotNil(t, cmd.Run)
}

func TestNewAPICmd_FlagDefaults(t *testing.T) {
	cmd := newAPICmd()

	// Host flag should default to empty (uses config)
	host, err := cmd.Flags().GetString("host")
	assert.NoError(t, err)
	assert.Empty(t, host)

	// Port flag should default to 0 (uses config)
	port, err := cmd.Flags().GetInt("port")
	assert.NoError(t, err)
	assert.Equal(t, 0, port)
}

func TestNewAPICmd_FlagTypes(t *testing.T) {
	cmd := newAPICmd()

	// Verify flag types
	hostFlag := cmd.Flags().Lookup("host")
	assert.NotNil(t, hostFlag)
	assert.Equal(t, "string", hostFlag.Value.Type())

	portFlag := cmd.Flags().Lookup("port")
	assert.NotNil(t, portFlag)
	assert.Equal(t, "int", portFlag.Value.Type())
}
