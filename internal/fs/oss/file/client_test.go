package file

import (
	"bytes"
	"fmt"
	"github.com/supaleon/vanilla/internal/fs/file"
	"github.com/supaleon/vanilla/internal/fs/oss"
	"io"
	"strings"
	"testing"
)

const (
	testMinPartSize  = 1024 * 1024 * 6
	testTemporaryDir = "testdata/001/.temporary"
)

var (
	testClient         *Client
	testWorkDir        = "testdata/001"
	testTempDir        = ""
	testPartSize       int64
	testEnableChecksum bool
	testForceChecksum  bool
	testFsync          bool
)

func testModeMultipart() {
	testPartSize = testMinPartSize
	testTempDir = testTemporaryDir
}

func testModeChecksum() {
	testEnableChecksum = true
	testTempDir = testTemporaryDir
}

func testModeForceChecksum() {
	testEnableChecksum = true
	testForceChecksum = true
	testTempDir = testTemporaryDir
}

func TestNew(t *testing.T) {
	var err error
	var options []Option
	if testTempDir != "" {
		options = append(options, WithTempDir(testTempDir))
	}
	if testPartSize > 0 {
		options = append(options, WithPartSize(testPartSize))
	}
	if testEnableChecksum {
		options = append(options, WithChecksum(false))
	}
	if testForceChecksum {
		options = append(options, WithChecksum(true))
	}
	if testFsync {
		options = append(options, WithFsync())
	}
	if testClient, err = New(testWorkDir, options...); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Open(t *testing.T) {
	if t.Run("TestNew", TestNew) {
		var err error
		for _, key := range testKeys {
			var obj oss.Object
			obj, err = testClient.Open(key)
			if err != nil {
				t.Fatalf("test key[%s] failed, err: %v", key, err)
			}
			fmt.Println(obj.Name())
		}
	}
}

func TestClient_Stat(t *testing.T) {
	testForceChecksum = true
	if t.Run("TestNew", TestNew) {
		for _, v := range testObjects {
			var err error
			var md oss.Metadata
			md, err = testClient.Stat(v.key)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("Size:", file.FormatSize(md.Size()))
			fmt.Println("time:", md.ModTime().String())
			fmt.Println("Etag:", md.Etag())
			fmt.Println("MD5:", md.MD5())
			fmt.Println("SHA256:", md.SHA256())
		}
	}
}

func TestClient_Iterate(t *testing.T) {
	testTempDir = testTemporaryDir
	testEnableChecksum = true
	if t.Run("TestNew", TestNew) {
		var err error
		var it oss.Iterator
		it, err = testClient.Iterate("")
		if err != nil {
			t.Fatal(err)
		}
		for {
			var f oss.Object
			f, err = it.Next()
			if err != nil {
				t.Fatal(err)
			}
			if f == nil {
				break
			}
			//var md oss.Metadata
			//md, err = f.Stat()
			//if err != nil {
			//	t.Fatal(err)
			//}
			fmt.Println(f.Name())
			//fmt.Println("Size:", files.FormatSize(md.Size()))
			//fmt.Println("ModTime:", md.ModTime().String())
			//fmt.Println("IsDir:", md.IsDir())
			//if !md.IsDir() {
			//	fmt.Println("Etag:", md.Etag())
			//	fmt.Println("MD5:", md.MD5())
			//	fmt.Println("SHA256:", md.SHA256())
			//}
			//fmt.Println("==================")
		}
	}
}

func TestClient_ReadDir(t *testing.T) {
	testTempDir = testTemporaryDir
	testEnableChecksum = true
	if t.Run("TestNew", TestNew) {
		var err error
		var it oss.Iterator
		it, err = testClient.ReadDir("")
		if err != nil {
			t.Fatal(err)
		}
		for {
			var f oss.Object
			f, err = it.Next()
			if err != nil {
				t.Fatal(err)
			}
			if f == nil {
				break
			}
			fmt.Println(f.Name())
		}
	}
}

func TestClient_Read(t *testing.T) {
	testForceChecksum = true
	if t.Run("TestNew", TestNew) {
		var err error
		var reader io.ReadCloser
		var md oss.Metadata
		reader, md, err = testClient.Read("mux.pdf")
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		fmt.Println("size:", md.Size())
		fmt.Println("modTime:", md.ModTime().String())
		fmt.Println("isDir:", md.IsDir())
		fmt.Println("calculateHexMD5:", md.Etag())
	}
}

func TestClient_Write(t *testing.T) {
	testModeChecksum()
	if t.Run("TestNew", TestNew) {
		for _, v := range testObjects {
			if !v.isDir {
				var err error
				var etag string

				cnt := 100
				n := int(v.partSize / 10)
				if n > 0 {
					cnt = n
				}
				etag, err = testClient.Write(v.key, bytes.NewBufferString(strings.Repeat("hello world\n", cnt)))
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println("Name: ", v.key)
				fmt.Println("Etag: ", etag)
			}
		}
	}
}

func TestClient_Remove(t *testing.T) {
	testEnableChecksum = true
	testTempDir = testTemporaryDir
	if t.Run("TestNew", TestNew) {
		for _, v := range testObjects {
			if v.isDir {
				continue
			}
			var err error
			if err = testClient.Remove(v.key); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestClient_RemoveAll(t *testing.T) {
	if t.Run("TestNew", TestNew) {
		var err error
		if err = testClient.RemoveAll(""); err != nil {
			t.Fatal(err)
		}
	}
}
