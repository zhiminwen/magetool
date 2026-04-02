package sshkit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSHClientWithPassword(t *testing.T) {
	client, err := NewSSHClient("example.com", "22", "ubuntu", "password", "")

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "example.com", client.Host)
	assert.Equal(t, "22", client.Port)
	assert.Equal(t, "ubuntu", client.User)
	assert.Equal(t, "password", client.Password)
	assert.Equal(t, "", client.PrivatekeyFile)
	assert.False(t, client.isConnected)
	require.NotNil(t, client.ClientConfig)
	assert.Equal(t, "ubuntu", client.ClientConfig.User)
}

func TestNewSSHClientRejectsEmptyAuth(t *testing.T) {
	client, err := NewSSHClient("example.com", "22", "ubuntu", "", "")

	require.Error(t, err)
	assert.Nil(t, client)
}

func TestCloseResetsStateWithoutClient(t *testing.T) {
	client := &SSHClient{
		isConnected: true,
	}

	err := client.Close()

	assert.NoError(t, err)
	assert.False(t, client.isConnected)
	assert.Nil(t, client.sshClient)
}

type integrationConfig struct {
	host     string
	port     string
	user     string
	password string
	keyfile  string
}

func loadIntegrationConfig(t *testing.T) integrationConfig {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping SSH integration tests in short mode")
	}

	cfg := integrationConfig{
		host:     os.Getenv("SSHKIT_TEST_HOST"),
		port:     os.Getenv("SSHKIT_TEST_PORT"),
		user:     os.Getenv("SSHKIT_TEST_USER"),
		password: os.Getenv("SSHKIT_TEST_PASSWORD"),
		keyfile:  os.Getenv("SSHKIT_TEST_KEYFILE"),
	}

	if cfg.host == "" || cfg.port == "" || cfg.user == "" || (cfg.password == "" && cfg.keyfile == "") {
		t.Skip("skipping SSH integration test; set SSHKIT_TEST_HOST, SSHKIT_TEST_PORT, SSHKIT_TEST_USER, and either SSHKIT_TEST_PASSWORD or SSHKIT_TEST_KEYFILE")
	}

	return cfg
}

func TestExecuteWithPassword(t *testing.T) {
	cfg := loadIntegrationConfig(t)
	if cfg.password == "" {
		t.Skip("skipping password integration test; SSHKIT_TEST_PASSWORD not set")
	}

	client, err := NewSSHClient(cfg.host, cfg.port, cfg.user, cfg.password, "")
	require.NoError(t, err)

	hostname, err := client.Capture("hostname")
	require.NoError(t, err)
	assert.NotEmpty(t, hostname)
}

func TestExecuteWithKey(t *testing.T) {
	cfg := loadIntegrationConfig(t)
	if cfg.keyfile == "" {
		t.Skip("skipping key integration test; SSHKIT_TEST_KEYFILE not set")
	}

	client, err := NewSSHClient(cfg.host, cfg.port, cfg.user, "", cfg.keyfile)
	require.NoError(t, err)
	require.NoError(t, client.Execute("echo starting"))
}
