package file

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/supaleon/vanilla/internal/kits"
	"io"
	"os"
	"path/filepath"
)

const (
	PathSeparator = string(os.PathSeparator)
	MultipartPath = "fragments"
)

type rangeCloser struct {
	*io.SectionReader
	file *os.File
}

func (c *rangeCloser) Close() error {
	return nil
}

func generateMultipartID() string {
	return kits.RandString(32)
}

func (c *Client) buildKeyPath(key string) string {
	return filepath.Join(c.workdir, key)
}

// The buildPartKeyPath build a temp relative path for part writing.
func (c *Client) buildPartKeyPath(key string, multipartId string) (path string) {
	h := md5.New()
	h.Write([]byte(key))
	hexMD5 := hex.EncodeToString(h.Sum(nil))
	if multipartId == "" {
		return filepath.Join(c.tempDir, MultipartPath, hexMD5)
	}
	return filepath.Join(c.tempDir, MultipartPath, hexMD5, multipartId)
}
