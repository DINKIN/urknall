package urknall

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pubsub"
)

type Task interface {
	Add(cmds ...interface{}) Task
	Commands() ([]cmd.Command, error)
	Key() string
	SetKey(string)
}

// A runlist is a container for commands. Use the following methods to add new commands.
type task struct {
	commands []cmd.Command

	name        string      // Name of the compilable.
	taskBuilder interface{} // only used for rendering templates TODO(gf): rename

	compiled  bool
	validated bool
}

func (t *task) SetKey(key string) {
	t.name = key
}

func (t *task) Key() string {
	return t.name
}

func (t *task) Commands() ([]cmd.Command, error) {
	e := t.Compile()
	if e != nil {
		return nil, e
	}
	return t.commands, nil
}

// Create a task from a set of commands without configuration.
func NewTask() Task {
	return &task{}
}

func (task *task) Add(cmds ...interface{}) Task {
	for _, c := range cmds {
		switch t := c.(type) {
		case string:
			// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
			// explicitly renders the template.
			task.addCommand(&stringCommand{cmd: t})
		case cmd.Command:
			task.addCommand(t)
		default:
			panic(fmt.Sprintf("type %T not supported!", t))
		}
	}
	return task
}

func (task *task) validate() error {
	if !task.validated {
		if task.taskBuilder == nil {
			return nil
		}
		e := validatePackage(task.taskBuilder)
		if e != nil {
			return e
		}
		task.validated = true
	}
	return nil
}

// Add the given command to the runlist.
func (task *task) addCommand(c cmd.Command) {
	if task.taskBuilder != nil {
		e := task.validate()
		if e != nil {
			panic(e.Error())
		}
		if renderer, ok := c.(cmd.Renderer); ok {
			renderer.Render(task.taskBuilder)
		}
		if validator, ok := c.(cmd.Validator); ok {
			if e := validator.Validate(); e != nil {
				panic(e.Error())
			}
		}
	}
	task.commands = append(task.commands, c)
}

func (task *task) Compile() (e error) {
	if task.compiled {
		return nil
	}
	m := &pubsub.Message{RunlistName: task.name, Key: pubsub.MessageRunlistsPrecompile}
	m.Publish("started")
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("failed to precompile package: %v %q", task.name, r)
			}
			m.Error = e
			m.Stack = string(debug.Stack())
			m.Publish("panic")
			log.Printf("ERROR: %s", r)
			log.Print(string(debug.Stack()))
		}
	}()

	e = task.validate()
	if e != nil {
		return e
	}
	m.Publish("finished")
	task.compiled = true
	return nil
}

type anonymousTask struct {
	cmds []interface{}
}

func (anon *anonymousTask) BuildTask(pkg Task) {
	for i := range anon.cmds {
		pkg.Add(anon.cmds[i])
	}
}
