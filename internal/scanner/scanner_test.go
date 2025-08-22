package scanner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/supaleon/vanilla/internal/token"
)

var testCases = []struct {
	Expect bool
	File   string
}{
	{
		Expect: true,
		File:   "testdata/bad-attr.html",
	},
}

func TestAny(t *testing.T) {
	for k := range 3 {
		println(k)
	}
}

func TestAll(t *testing.T) {
	var errCnt int
	for _, c := range testCases {
		println("🏄‍♂️========================> ", c.File)
		err := run(c.File)
		if err != nil && c.Expect {
			println("❌========================> ", "Test", c.File, "failed")
			errCnt++
			continue
		}
		println("✅========================> ", "Test", c.File, "Succeed")
	}

	if errCnt > 0 {
		t.Fail()
	}
}

func run(source string) (err error) {
	f, _ := os.Open(source)
	buf, _ := io.ReadAll(f)
	defer f.Close()

	fileSet := token.NewFileSet()
	file := fileSet.AddFile(source, fileSet.Base(), len(buf))
	scanner := New(file, buf, func(pos token.Position, msg string) {
		err = errors.New(fmt.Sprintln(pos.String(), msg))
		println(err.Error())
	})

	for {
		_, tok, lit := scanner.Scan()
		s := tok.String()
		if !tok.IsOperator() && tok != token.EOFToken {
			s = lit
		}
		//goland:noinspection GoDfaConstantCondition
		if tok == token.ErrorToken {
			println(fmt.Sprintf("----------token:%q", s))
		} else {
			println("----------token:", s)
		}
		if tok == token.EOFToken {
			break
		}
	}
	return
}
