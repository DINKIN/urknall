package urknall

import (
	"crypto/sha256"
	"fmt"

	"github.com/dynport/urknall/cmd"
)

func renderTemplate(builder Template) (*packageImpl, error) {
	p := &packageImpl{reference: builder}
	e := validateTemplate(builder)
	if e != nil {
		return nil, e
	}
	builder.Render(p)
	return p, nil
}

func commandChecksum(c cmd.Command) (string, error) {
	if c, ok := c.(interface {
		Checksum() string
	}); ok {
		return c.Checksum(), nil
	}
	s := sha256.New()
	if _, e := s.Write([]byte(c.Shell())); e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", s.Sum(nil)), nil
}

func midTrunc(in string, l int) string {
	if len(in) <= l {
		return in
	}
	chars := l - 3
	beg := chars / 2
	end := beg
	if chars%2 == 1 {
		beg++
	}
	return in[0:beg] + "..." + in[len(in)-end:]
}
