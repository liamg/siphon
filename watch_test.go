package main

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_Watch(t *testing.T) {

	require.Equal(t, 0, os.Getuid(), "test must be run as root")

	pingBuffer := bytes.NewBuffer([]byte{})
	pingCmd := exec.Command("ping", "8.8.8.8")
	pingCmd.Stdout = pingBuffer
	require.NoError(t, pingCmd.Start())

	go func() {
		time.Sleep(5 * time.Second)
		require.NoError(t, pingCmd.Process.Kill())
		require.NoError(t, pingCmd.Process.Release())
	}()

	time.Sleep(time.Second * 1)

	watchBuffer := bytes.NewBuffer([]byte{})

	watchCmd := exec.Command("go", "run", ".", strconv.Itoa(pingCmd.Process.Pid))
	watchCmd.Stdout = watchBuffer
	require.NoError(t, watchCmd.Start())

	_ = pingCmd.Wait()
	_ = watchCmd.Wait()

	assert.True(t, strings.HasSuffix(pingBuffer.String(), watchBuffer.String()))
	assert.Greater(t, len(watchBuffer.String()), 0)
}
