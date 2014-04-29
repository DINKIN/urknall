package ssh

import (
	"fmt"
	"net"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"github.com/dynport/urknall/cmd"
)

func New(address string) *Host {
	return &Host{Address: address}
}

type Host struct {
	Address  string
	Password string

	address string
	port    int
	user    string

	client *ssh.Client
}

func (host *Host) User() string {
	host.parseAddress()
	return host.user
}

func (host *Host) String() string {
	host.parseAddress()
	return host.address
}

func (host *Host) parseAddress() {
	if host.port > 0 {
		return
	}
	hostAndPort := strings.Split(host.Address, ":")
	var addr string
	if len(hostAndPort) == 2 {
		addr = hostAndPort[0]
	} else {
		host.port = 22
		addr = host.Address
	}
	userAndAddress := strings.Split(addr, "@")
	if len(userAndAddress) == 2 {
		host.user = userAndAddress[0]
		host.address = userAndAddress[1]
	} else {
		host.user = "root"
		host.address = addr
	}

}

type SshClient interface {
	Client() (*ssh.Client, error)
}

func (c *Host) Client() (*ssh.Client, error) {
	c.parseAddress()
	var e error
	config := &ssh.ClientConfig{
		User: c.user,
	}
	if c.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(c.Password))
	} else if sshSocket := os.Getenv("SSH_AUTH_SOCK"); sshSocket != "" {
		if c, e := net.Dial("unix", sshSocket); e == nil {
			config.Auth = append(config.Auth, ssh.PublicKeysCallback(agent.NewClient(c).Signers))
		}
	}
	con, e := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.address, c.port), config)
	if e != nil {
		return nil, e
	}
	return &ssh.Client{Conn: con}, nil
}

func (c *Host) Command(cmd string) (cmd.ExecCommand, error) {
	if c.client == nil {
		var e error
		c.client, e = c.Client()
		if e != nil {
			return nil, e
		}
	}
	ses, e := c.client.NewSession()
	if e != nil {
		return nil, e
	}
	return &Command{command: cmd, session: ses}, nil
}
