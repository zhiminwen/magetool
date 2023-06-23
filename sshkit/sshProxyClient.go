package sshkit

import (
	"golang.org/x/crypto/ssh"
)

func (c *SSHClient) ProxySSHClient(host, port, user, password, keyfile string) (*SSHClient, error) {
	err := c.Connect()
	if err != nil {
		return nil, err
	}

	proxyClient, err := NewSSHClient(host, port, user, password, keyfile)
	if err != nil {
		return nil, err
	}

	conn, err := c.sshClient.Dial("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, host+":"+port, proxyClient.ClientConfig)
	if err != nil {
		return nil, err
	}

	client := ssh.NewClient(ncc, chans, reqs)

	proxyClient.sshClient = client
	proxyClient.isConnected = true

	return proxyClient, nil
}
