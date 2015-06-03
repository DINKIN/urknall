package urknall

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/target"
)

// TODO: rename to remoteCommandRunner
type remoteTaskRunner struct {
	build   *Build
	dir     string
	command cmd.Command

	taskName    string
	taskStarted time.Time

	commandStarted time.Time
}

func (runner *remoteTaskRunner) run() error {
	runner.commandStarted = time.Now()

	checksum, e := commandChecksum(runner.command)
	if e != nil {
		return e
	}
	prefix := runner.dir + "/" + checksum

	if e = runner.writeScriptFile(prefix); e != nil {
		return e
	}

	errors := make(chan error)
	logs := runner.newLogWriter(prefix+".log", errors)

	c, e := runner.build.prepareCommand("sh " + prefix + ".sh")
	if e != nil {
		return e
	}

	var wg sync.WaitGroup

	// Get pipes for stdout and stderr and forward messages to logs channel.
	stdout, e := c.StdoutPipe()
	if e != nil {
		return e
	}
	wg.Add(1)
	go runner.forwardStream(logs, "stdout", &wg, stdout)

	stderr, e := c.StderrPipe()
	if e != nil {
		return e
	}
	wg.Add(1)
	go runner.forwardStream(logs, "stderr", &wg, stderr)

	if sc, ok := runner.command.(cmd.StdinConsumer); ok {
		c.SetStdin(sc.Input())
		defer sc.Input().Close()
	}

	e = c.Run()
	wg.Wait()
	close(logs)

	if e = runner.createChecksumFile(prefix, e); e != nil {
		return e
	}

	// Get errors that might have occured while handling the back-channel for the logs.
	for e = range errors {
		log.Printf("ERROR: %s", e.Error())
	}
	return e
}

func (runner *remoteTaskRunner) writeScriptFile(prefix string) (e error) {
	targetFile := prefix + ".sh"
	env := ""
	for _, e := range runner.build.Env {
		env += "export " + e + "\n"
	}
	rawCmd := fmt.Sprintf("cat <<\"EOSCRIPT\" > %s\n#!/bin/sh\nset -e\nset -x\n\n%s\n%s\nEOSCRIPT\n", targetFile, env, runner.command.Shell())
	c, e := runner.build.prepareInternalCommand(rawCmd)
	if e != nil {
		return e
	}

	return c.Run()
}

func (runner *remoteTaskRunner) createChecksumFile(prefix string, err error) (e error) {
	sourceFile := prefix + ".sh"
	targetFile := prefix + ".done"
	if err != nil {
		logError(err)
		targetFile = prefix + ".failed"
	}
	rawCmd := fmt.Sprintf("{ [ -f %[1]s ] || mv %[2]s %[1]s; } && echo %[1]s >> %[3]s/%[4]s.run",
		targetFile, sourceFile, runner.dir, runner.taskStarted.Format("20060102_150405"))
	c, e := runner.build.prepareInternalCommand(rawCmd)
	if e != nil {
		return e
	}

	if e = c.Run(); e != nil {
		return e
	}

	return err // return original error for convenience
}

func logError(e error) {
	log.Printf("ERROR: %s", e.Error())
}

func (runner *remoteTaskRunner) forwardStream(logs chan<- string, stream string, wg *sync.WaitGroup, r io.Reader) {
	defer wg.Done()

	m := message("task.io", runner.build.hostname(), runner.taskName)
	m.Message = runner.command.Shell()
	if logger, ok := runner.command.(cmd.Logger); ok {
		m.Message = logger.Logging()
	}
	m.Stream = stream

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		m.Line = scanner.Text()
		m.TotalRuntime = time.Since(runner.commandStarted)
		m.Publish(stream)
		logs <- time.Now().UTC().Format(time.RFC3339Nano) + "\t" + stream + "\t" + scanner.Text()
	}
}

func (runner *remoteTaskRunner) newLogWriter(path string, errors chan<- error) chan<- string {
	logs := make(chan string)
	go runner.writeLogs(path, errors, logs)
	return logs
}

func (runner *remoteTaskRunner) writeLogs(path string, errors chan<- error, logs <-chan string) {
	defer close(errors)

	var cmd target.ExecCommand

	err := func() error {
		// so ugly, but: sudo not required and "sh -c" adds some escaping issues with the variables. This is why Command is called directly.
		var err error
		if cmd, err = runner.build.Command("cat - > " + path); err != nil {
			return err
		}

		// Get pipe to stdin of the execute command.
		in, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		defer in.Close()

		// Run command, writing everything coming from stdin to a file.
		if err := cmd.Start(); err != nil {
			in.Close()
			return err
		}

		// Send all messages from logs to the stdin of the new session.
		for log := range logs {
			if _, err = io.WriteString(in, log+"\n"); err != nil {
				errors <- err
			}
		}
		return err
	}()
	if err != nil {
		errors <- err
	} else if err := cmd.Wait(); err != nil {
		errors <- err
	}
}
