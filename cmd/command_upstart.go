package cmd

import (
	"fmt"
	"github.com/dynport/dgtk/goup"
)

// Create an upstart command. That is a script executed on system start. See the github.com/dynport/dgtk/goup package
// for further details.
type UpstartCommand struct {
	Upstart *goup.Upstart // Upstart configuration.
}

func (uA *UpstartCommand) Docker() string {
	return ""
}

func (uA *UpstartCommand) Shell() string {
	if uA.Upstart == nil {
		return ""
	}
	fA := &FileCommand{
		Path:        fmt.Sprintf("/etc/init/%s.conf", uA.Upstart.Name),
		Content:     uA.Upstart.CreateScript(),
		Permissions: 0644,
	}
	return fA.Shell()
}

func (uA *UpstartCommand) Logging() string {
	return fmt.Sprintf("[UPSTART] Adding upstart script for '%s'.", uA.Upstart.Name)
}
