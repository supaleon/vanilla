package scanner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/supaleon/vanilla/internal/token"
)

func TestAny(t *testing.T) {
	var z = 1
	if z == -1 {
		println("yes")
	}
}

var testCases = []struct {
	Expect bool
	File   string
}{
	{
		Expect: true,
		File:   "./testdata/Main.html",
	},
	{
		Expect: false,
		File:   "./testdata/bad-attr.html",
	},
	{
		Expect: false,
		File:   "./testdata/bad-if.html",
	},
}

func TestAll(t *testing.T) {
	var errCnt int
	for _, c := range testCases {
		println("👇========================> ", c.File)
		err := run(c.File)
		if err != nil && c.Expect {
			println("👆❌========================> ", "Test", c.File, "failed\n")
			errCnt++
			continue
		}
		println("👆✅========================> ", "Test", c.File, "succeed\n")
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
		if !tok.IsOperator() && tok != token.EOF {
			s = lit
		}
		//goland:noinspection GoDfaConstantCondition
		if tok == token.ILLEGAL {
			println(fmt.Sprintf("----------token:%q", s))
		} else {
			println("----------token:", s)
		}
		if tok == token.EOF {
			break
		}
	}
	return
}
