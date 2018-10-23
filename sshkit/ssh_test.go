package sshkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	client := NewSSHClient("192.168.5.10", "22", "ubuntu", "password", "")
	assert.NotNil(t, client)
}

func TestKeyfile(t *testing.T) {
	client := NewSSHClient("192.168.5.10", "22", "ubuntu", "", `C:\Tools\Kitty\mykey.openssh`)
	assert.NotNil(t, client)
}

func TestExecuteWithPassword(t *testing.T) {
	client := NewSSHClient("192.168.5.10", "22", "ubuntu", "password", "")
	hostname, err := client.Capture("hostname")
	if err != nil {
		t.Fatalf("Failed to capture:%v", err)
	}

	assert.Equal(hostname, "ubuntu")

}

func TestExecuteWithKey(t *testing.T) {
	client := NewSSHClient("192.168.5.10", "22", "ubuntu", "", `C:\Tools\Kitty\mykey.openssh`)
	client.Execute("bash", "-c", "echo starting")
}
