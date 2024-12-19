package ssh

import (
	"fmt"
	"ops_cli/pkg/log"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	host     string
	user     string
	password string
	port     int
	client   *ssh.Client
}

func NewClient(host, user, password string, port int) *Client {
	return &Client{
		host:     host,
		user:     user,
		password: password,
		port:     port,
	}
}

func (c *Client) Connect() error {
	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         60 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	log.Debug("Attempting to connect to %s with user %s using password authentication", addr, c.user)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}

	log.Debug("Successfully established SSH connection to %s", addr)
	c.client = client
	return nil
}

func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *Client) RunCommand(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("client not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	log.Debug("Executing command: %s", cmd)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %v", err)
	}

	return string(output), nil
}
