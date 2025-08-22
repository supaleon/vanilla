package file

import (
	"fmt"
	"testing"
)

func TestAbs(t *testing.T) {
	name, err := Abs("~/go")
	if err != nil {
		t.Fatal(err)
	}
	if name != HomeDir()+"/go" {
		t.Fatal("Assert failed")
	}
	fmt.Println(name)
}

func TestHomeDir(t *testing.T) {
	fmt.Println(HomeDir())
}

func TestTempDir(t *testing.T) {
	name, err := TempDir("", "aaa")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(name)
}

func TestSegments(t *testing.T) {
	dir, name, ext := Segments("/var/folders/wk/ttn4ghlx4nq_js8j7dyp9s9w0000gn/T/aaa561302397/sda")
	fmt.Println(dir, name, ext)
}
