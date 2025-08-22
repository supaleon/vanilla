package file

import (
	"bytes"
	"fmt"
	"github.com/supaleon/vanilla/internal/fs/oss"
	"testing"
)

var (
	testMultipartKey    = "mp.txt"
	testMultipartId     string
	testMultipart       oss.Multipart
	testMultipartObject oss.Object
	testParts           []oss.Part
)

func TestObject_OpenMultipart(t *testing.T) {
	if t.Run("TestNew", TestNew) {
		var err error
		if testMultipartObject, err = testClient.Open(testMultipartKey); err != nil {
			t.Fatal(err)
		}
		if testMultipart, err = testMultipartObject.OpenMultipart(testMultipartId); err != nil {
			t.Fatal(err)
		}
		fmt.Println("Object Name:", testMultipartObject.Name())
		fmt.Println("MultipartId:", testMultipart.ID())
	}
}

func TestMultipart_Write(t *testing.T) {
	if t.Run("TestNew", TestNew) {
		testMultipartId = "xRLoYQfdSAlzvbSXgXPAnchqCqKUMjii"
		if t.Run("TestObject_OpenMultipart", TestObject_OpenMultipart) {
			var err error
			for i := 1; i < 10; i++ {
				var partEtag string
				data := fmt.Sprintf("test words %d \n", i)
				buf := []byte(data)
				//fmt.Println(int64(len(buf)))
				testParts = append(testParts, oss.Part{
					Size:   int64(len(buf)),
					Number: int64(i),
				})
				partEtag, err = testMultipart.Write(int64(i), bytes.NewBufferString(data))
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println("Part Etag:", partEtag)
			}
			//var etag string
			//if etag, err = testMultipartObject.MergeParts(parts...); err != nil {
			//	t.Fatal(err)
			//}
			//fmt.Println("Total Etag:", etag)
		}
	}
}

func TestMultipart_List(t *testing.T) {
	testMultipartId = "xRLoYQfdSAlzvbSXgXPAnchqCqKUMjii"
	if t.Run("TestObject_OpenMultipart", TestObject_OpenMultipart) {
		var err error
		testParts, err = testMultipart.List()
		if err != nil {
			t.Fatal(err)
		}
		for _, part := range testParts {
			fmt.Printf("%-15s %-50d \n", "Part Number:", part.Number)
			fmt.Printf("%-15s %-50d \n", "Part Size:", part.Size)
			fmt.Printf("%-15s %-50d \n", "Part Offset:", part.Offset)
			fmt.Printf("%-15s %-50s \n\n", "Part Etag:", part.Etag)
		}
	}
}

func TestMultipart_Merge(t *testing.T) {
	if t.Run("TestMultipart_Write", TestMultipart_Write) {
		var err error
		var etag string
		etag, err = testMultipart.Merge(testParts...)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Etag: ", etag)
	}
}

func TestMultipart_Abort(t *testing.T) {
	if t.Run("TestMultipart_List", TestMultipart_List) {
		var err error
		err = testMultipart.Abort()
		if err != nil {
			t.Fatal(err)
		}
	}
}
