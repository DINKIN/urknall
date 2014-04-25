package local

import (
	"io"
	"os/exec"
)

type Command struct {
	command *exec.Cmd
}

func (c *Command) StdoutPipe() (io.Reader, error) {
	return c.command.StdoutPipe()
}

func (c *Command) StderrPipe() (io.Reader, error) {
	return c.command.StderrPipe()
}

func (c *Command) StdinPipe() (io.Writer, error) {
	return c.command.StdinPipe()
}

func (c *Command) SetStdout(w io.Writer) {
	c.command.Stdout = w
}

func (c *Command) SetStderr(w io.Writer) {
	c.command.Stderr = w
}

func (c *Command) SetStdin(r io.Reader) {
	c.command.Stdin = r
}

func (c *Command) Wait() error {
	return c.command.Wait()
}

func (c *Command) Start() error {
	return c.command.Start()
}

func (c *Command) Run() error {
	return c.command.Run()
}
