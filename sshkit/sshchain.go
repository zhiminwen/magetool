package sshkit

import "fmt"

const (
	AuthTypePassword = "password"
	AuthTypeKey      = "key"
)

type SSHHostConfig struct {
	Host           string
	Port           int
	User           string
	Authentication string
	Password       string
	Key            string
}

type newSSHClientFunc func(host, port, user, password, keyfile string) (*SSHClient, error)
type proxySSHClientFunc func(parent *SSHClient, host, port, user, password, keyfile string) (*SSHClient, error)

func NewSSHClientChain(hops []SSHHostConfig) (*SSHClient, error) {
	return buildSSHClientChain(hops, NewSSHClient, proxySSHClient)
}

func buildSSHClientChain(hops []SSHHostConfig, newClient newSSHClientFunc, proxyClient proxySSHClientFunc) (*SSHClient, error) {
	if len(hops) == 0 {
		return nil, fmt.Errorf("failed to construct ssh client chain: no hosts provided")
	}

	var client *SSHClient
	for i, hop := range hops {
		password, keyfile, err := resolveSSHAuth(hop)
		if err != nil {
			return nil, err
		}

		port := fmt.Sprintf("%d", hop.Port)
		if i == 0 {
			client, err = newClient(hop.Host, port, hop.User, password, keyfile)
		} else {
			client, err = proxyClient(client, hop.Host, port, hop.User, password, keyfile)
		}
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func proxySSHClient(parent *SSHClient, host, port, user, password, keyfile string) (*SSHClient, error) {
	return parent.ProxySSHClient(host, port, user, password, keyfile)
}

func resolveSSHAuth(hop SSHHostConfig) (password, keyfile string, err error) {
	if hop.Host == "" {
		return "", "", fmt.Errorf("failed to construct ssh client chain: host is required")
	}
	if hop.Port <= 0 {
		return "", "", fmt.Errorf("failed to construct ssh client chain for host %q: valid port is required", hop.Host)
	}
	if hop.User == "" {
		return "", "", fmt.Errorf("failed to construct ssh client chain for host %q: user is required", hop.Host)
	}

	switch hop.Authentication {
	case AuthTypePassword:
		if hop.Password == "" {
			return "", "", fmt.Errorf("failed to construct ssh client chain for host %q: password is required for password authentication", hop.Host)
		}
		return hop.Password, "", nil
	case AuthTypeKey:
		if hop.Key == "" {
			return "", "", fmt.Errorf("failed to construct ssh client chain for host %q: key is required for key authentication", hop.Host)
		}
		return "", hop.Key, nil
	default:
		if hop.Authentication == "" {
			return "", "", fmt.Errorf("failed to construct ssh client chain for host %q: authentication is required", hop.Host)
		}
		return "", "", fmt.Errorf("failed to construct ssh client chain for host %q: unsupported authentication %q", hop.Host, hop.Authentication)
	}
}
