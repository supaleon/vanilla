package lang

import (
	"github.com/supaleon/vanilla/internal/lang/token"
	"github.com/supaleon/vanilla/internal/lang/tokenizer"
	"io"
	"math"
	"os"
	"testing"
)

func TestHtml(t *testing.T) {
	var err error
	var file *os.File
	file, err = os.Open("./Hello.html")
	if err != nil {
		t.Fatal(err)
	}
	var fileInfo os.FileInfo
	fileInfo, err = file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	size := math.MaxInt
	if int64(size) > fileInfo.Size() {
		size = int(fileInfo.Size())
	}
	tokenFile := fileSet.AddFile(file.Name(), fileSet.Base(), size)
	src, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	scanner := tokenizer.NewScanner(tokenFile, src, func(pos token.Position, msg string) {
		t.Fatal(pos, msg)
	})
	scanner.Scan()
}
