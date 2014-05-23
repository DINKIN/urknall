package urknall

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

type testPackageBuilder struct {
	Version string `urknall:"required=true"`
}

func (t *testPackageBuilder) Build(p Package) {
	ts := NewTask()
	ts.SetKey("base")
	ts.Add("echo base {{ .Version }}")
}

func TestBuildPackage(t *testing.T) {
	Convey("Build package", t, func() {
		task := NewTask()
		task.SetKey("base")
		task.Add("apt-get update")
		So(task, ShouldNotBeNil)
		pkg := &packageImpl{}
		pkg.AddTask(task)

		So(len(pkg.tasks), ShouldEqual, 1)
		t.Log(task.Commands())
	})
}
