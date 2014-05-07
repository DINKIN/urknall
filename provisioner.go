package urknall

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pubsub"
)

type checksumTree map[string]map[string]struct{}

type Provisioner interface {
	ProvisionRunlist(*Task, checksumTree) error
	BuildChecksumTree() (checksumTree, error)
}

func (build *Build) buildChecksumTree() (ct checksumTree, e error) {
	ct = checksumTree{}

	cmd, e := build.prepareCommand(fmt.Sprintf(`[ -d %[1]s ] && find %[1]s -type f -name \*.done`, ukCACHEDIR))
	if e != nil {
		return nil, e
	}
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}

	cmd.SetStdout(out)
	cmd.SetStderr(err)

	if e := cmd.Run(); e != nil {
		return nil, e
	}
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		pkgname := filepath.Dir(strings.TrimPrefix(line, ukCACHEDIR+"/"))
		checksum := strings.TrimSuffix(filepath.Base(line), ".done")
		if len(checksum) != 64 {
			return nil, fmt.Errorf("invalid checksum %q found for package %q", checksum, pkgname)
		}
		if _, found := ct[pkgname]; !found {
			ct[pkgname] = map[string]struct{}{}
		}
		ct[pkgname][checksum] = struct{}{}
	}

	return ct, nil
}

func (build *Build) run() (e error) {
	ct, e := build.buildChecksumTree()
	if e != nil {
		return e
	}

	for i := range build.Pkg.tasks {
		task := build.Pkg.tasks[i]
		m := &pubsub.Message{Key: pubsub.MessageRunlistsProvision, Hostname: build.hostname()}
		m.Publish("started")
		if e = build.provisionRunlist(task, ct); e != nil {
			m.PublishError(e)
			return e
		}
		m.Publish("finished")
	}
	return nil
}

func (build *Build) provisionRunlist(task *Task, ct checksumTree) (e error) {
	commands := task.rawCommands()

	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", task.name)

	var found bool
	var checksumHash map[string]struct{}
	if checksumHash, found = ct[task.name]; !found {
		ct[task.name] = map[string]struct{}{}
		checksumHash = ct[task.name]

		// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
		// different users (being part of that group) to create, modify and delete the contained checksum and log files.
		createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)

		cmd, e := build.prepareCommand(createChecksumDirCmd)
		if e != nil {
			return e
		}
		err := &bytes.Buffer{}

		cmd.SetStderr(err)

		if e := cmd.Run(); e != nil {
			return fmt.Errorf(err.String() + ": " + e.Error())
		}
	}

	for _, cmd := range commands {
		logMsg := cmd.Logging()
		m := &pubsub.Message{Key: pubsub.MessageRunlistsProvisionTask, TaskChecksum: cmd.checksum, Message: logMsg, Hostname: build.hostname(), RunlistName: task.name}
		if _, found := checksumHash[cmd.checksum]; found { // Task is cached.
			m.ExecStatus = pubsub.StatusCached
			m.Publish("finished")
			delete(checksumHash, cmd.checksum) // Delete checksums of cached tasks from hash.
			continue
		}

		if len(checksumHash) > 0 { // All remaining checksums are invalid, as something changed.
			if e = build.cleanUpRemainingCachedEntries(checksumDir, checksumHash); e != nil {
				return e
			}
			checksumHash = make(map[string]struct{})
		}
		m.ExecStatus = pubsub.StatusExecStart
		if build.DryRun {
			m.Publish("executed")
		} else {
			m.Publish("started")
			e = cmd.execute(build, checksumDir)
			m.Error = e
			m.ExecStatus = pubsub.StatusExecFinished
			m.Publish("finished")
		}
		if e != nil {
			return e
		}
	}

	return nil
}

type rawCommand struct {
	cmd.Command        // The command to be executed.
	checksum    string // The checksum of the command.
	task        *Task
}

func (cmd *rawCommand) execute(build *Build, checksumDir string) (e error) {
	sCmd := fmt.Sprintf("sh -x -e -c %q", cmd.Shell())
	for _, env := range build.Env {
		sCmd = env + " " + sCmd
	}
	r := &remoteTaskRunner{build: build, cmd: sCmd, rawCommand: cmd, dir: checksumDir}
	return r.run()
}

func (build *Build) cleanUpRemainingCachedEntries(checksumDir string, checksumHash map[string]struct{}) (e error) {
	invalidCacheEntries := make([]string, 0, len(checksumHash))
	for k, _ := range checksumHash {
		invalidCacheEntries = append(invalidCacheEntries, fmt.Sprintf("%s.done", k))
	}
	if build.DryRun {
		(&pubsub.Message{Key: pubsub.MessageCleanupCacheEntries, InvalidatedCacheEntries: invalidCacheEntries, Hostname: build.hostname()}).Publish(".dryrun")
	} else {
		cmd := fmt.Sprintf("cd %s && rm -f *.failed %s", checksumDir, strings.Join(invalidCacheEntries, " "))
		m := &pubsub.Message{Key: pubsub.MessageUrknallInternal, Hostname: build.hostname()}
		m.Publish("started")

		c, e := build.prepareCommand(cmd)
		if e != nil {
			return e
		}
		if e := c.Run(); e != nil {
			return e
		}
		//m.sshResult = result
		m.Publish("finished")
	}
	return nil
}
