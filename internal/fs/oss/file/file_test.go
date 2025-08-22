package file

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestObject_Md5(t *testing.T) {
	hashSHA1 := sha1.New()
	hashSHA1.Write([]byte(""))
	s := hex.EncodeToString(hashSHA1.Sum(nil))
	fmt.Println(s)
}
