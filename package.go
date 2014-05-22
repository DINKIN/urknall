package urknall

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall/utils"
)

type Template interface {
	Render(pkg Package)
}

type Package interface {
	AddTemplate(string, Template)
	AddCommands(string, ...command)
	AddTask(Task)
	Tasks() []Task
}

type packageImpl struct {
	tasks          []Task
	taskNames      map[string]struct{}
	reference      interface{} // used for rendering
	cacheKeyPrefix string
}

func (p *packageImpl) Tasks() []Task {
	return p.tasks
}

func (pkg *packageImpl) AddCommands(name string, cmds ...command) {
	if pkg.cacheKeyPrefix != "" {
		name = pkg.cacheKeyPrefix + "." + name
	}
	name = utils.MustRenderTemplate(name, pkg.reference)
	task := &taskImpl{name: name}
	for _, c := range cmds {
		if r, ok := c.(renderer); ok {
			r.Render(pkg.reference)
		}
		task.Add(c)
	}
	pkg.addTask(task, false)
}

func (pkg *packageImpl) AddTemplate(name string, tpl Template) {
	if pkg.cacheKeyPrefix != "" {
		name = pkg.cacheKeyPrefix + "." + name
	}
	name = utils.MustRenderTemplate(name, pkg.reference)
	e := validatePackage(tpl)
	if e != nil {
		panic(e)
	}
	if pkg.reference != nil {
		name = utils.MustRenderTemplate(name, pkg.reference)
	}
	pkg.validateTaskName(name)
	child := &packageImpl{cacheKeyPrefix: name, reference: tpl}
	tpl.Render(child)
	for _, task := range child.Tasks() {
		pkg.addTask(task, false)
	}
}

func (pkg *packageImpl) addTask(task Task, addPrefix bool) {
	if addPrefix {
		name := task.CacheKey()
		if pkg.cacheKeyPrefix != "" {
			name = pkg.cacheKeyPrefix + "." + task.CacheKey()
		}
		name = utils.MustRenderTemplate(name, pkg.reference)
		task.SetCacheKey(name)
	}
	pkg.validateTaskName(task.CacheKey())
	pkg.taskNames[task.CacheKey()] = struct{}{}
	pkg.tasks = append(pkg.tasks, task)
}

func (pkg *packageImpl) AddTask(task Task) {
	pkg.addTask(task, true)
}

func (pkg *packageImpl) precompile() (e error) {
	for _, task := range pkg.tasks {
		c, e := task.Commands()
		if e != nil {
			return e
		}
		if len(c) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", task.CacheKey())
		}

		if tc, ok := task.(interface {
			Compile() error
		}); ok {
			e := tc.Compile()
			if e != nil {
				return e
			}
		}
	}

	return nil
}

func (pkg *packageImpl) validateTaskName(name string) {
	if name == "" {
		panic("package names must not be empty!")
	}

	if strings.Contains(name, " ") {
		panic(fmt.Sprintf(`package names must not contain spaces (%q does)`, name))
	}

	if pkg.taskNames == nil {
		pkg.taskNames = map[string]struct{}{}
	}

	if _, ok := pkg.taskNames[name]; ok {
		panic(fmt.Sprintf("package with name %q exists already", name))
	}
}
