package sshkit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSHClientChainSingleHop(t *testing.T) {
	client, err := NewSSHClientChain([]SSHHostConfig{
		{
			Host:           "example.com",
			Port:           22,
			User:           "ubuntu",
			Authentication: AuthTypePassword,
			Password:       "secret",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "example.com", client.Host)
	assert.Equal(t, "22", client.Port)
	assert.Equal(t, "ubuntu", client.User)
	assert.Equal(t, "secret", client.Password)
	assert.Equal(t, "", client.PrivatekeyFile)
}

func TestNewSSHClientChainRejectsEmptyChain(t *testing.T) {
	client, err := NewSSHClientChain(nil)

	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "no hosts provided")
}

func TestNewSSHClientChainRejectsUnsupportedAuth(t *testing.T) {
	client, err := NewSSHClientChain([]SSHHostConfig{
		{
			Host:           "example.com",
			Port:           22,
			User:           "ubuntu",
			Authentication: "agent",
		},
	})

	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), `unsupported authentication "agent"`)
}

func TestNewSSHClientChainRejectsMissingAuthValue(t *testing.T) {
	testCases := []struct {
		name string
		hop  SSHHostConfig
		want string
	}{
		{
			name: "missing authentication type",
			hop: SSHHostConfig{
				Host: "example.com",
				Port: 22,
				User: "ubuntu",
			},
			want: "authentication is required",
		},
		{
			name: "missing password",
			hop: SSHHostConfig{
				Host:           "example.com",
				Port:           22,
				User:           "ubuntu",
				Authentication: AuthTypePassword,
			},
			want: "password is required",
		},
		{
			name: "missing key",
			hop: SSHHostConfig{
				Host:           "example.com",
				Port:           22,
				User:           "ubuntu",
				Authentication: AuthTypeKey,
			},
			want: "key is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewSSHClientChain([]SSHHostConfig{tc.hop})

			require.Error(t, err)
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestNewSSHClientChainRejectsMissingRequiredFields(t *testing.T) {
	testCases := []struct {
		name string
		hop  SSHHostConfig
		want string
	}{
		{
			name: "missing host",
			hop: SSHHostConfig{
				Port:           22,
				User:           "ubuntu",
				Authentication: AuthTypePassword,
				Password:       "secret",
			},
			want: "host is required",
		},
		{
			name: "missing port",
			hop: SSHHostConfig{
				Host:           "example.com",
				User:           "ubuntu",
				Authentication: AuthTypePassword,
				Password:       "secret",
			},
			want: "valid port is required",
		},
		{
			name: "missing user",
			hop: SSHHostConfig{
				Host:           "example.com",
				Port:           22,
				Authentication: AuthTypePassword,
				Password:       "secret",
			},
			want: "user is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewSSHClientChain([]SSHHostConfig{tc.hop})

			require.Error(t, err)
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestBuildSSHClientChainSequencesProxyHops(t *testing.T) {
	var calls []string
	newClient := func(host, port, user, password, keyfile string) (*SSHClient, error) {
		calls = append(calls, fmt.Sprintf("new:%s:%s:%s:%s:%s", host, port, user, password, keyfile))
		return &SSHClient{Host: host, Port: port, User: user, Password: password, PrivatekeyFile: keyfile}, nil
	}
	proxyClient := func(parent *SSHClient, host, port, user, password, keyfile string) (*SSHClient, error) {
		calls = append(calls, fmt.Sprintf("proxy:%s->%s:%s:%s:%s:%s", parent.Host, host, port, user, password, keyfile))
		return &SSHClient{Host: host, Port: port, User: user, Password: password, PrivatekeyFile: keyfile}, nil
	}

	client, err := buildSSHClientChain([]SSHHostConfig{
		{
			Host:           "jump-1",
			Port:           22,
			User:           "alice",
			Authentication: AuthTypePassword,
			Password:       "pw1",
		},
		{
			Host:           "jump-2",
			Port:           2222,
			User:           "bob",
			Authentication: AuthTypeKey,
			Key:            "/tmp/id_jump_2",
		},
		{
			Host:           "target",
			Port:           22,
			User:           "carol",
			Authentication: AuthTypePassword,
			Password:       "pw3",
		},
	}, newClient, proxyClient)

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "target", client.Host)
	assert.Equal(t, []string{
		"new:jump-1:22:alice:pw1:",
		"proxy:jump-1->jump-2:2222:bob::/tmp/id_jump_2",
		"proxy:jump-2->target:22:carol:pw3:",
	}, calls)
}
