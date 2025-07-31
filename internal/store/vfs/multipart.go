package vfs

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/supaleon/vanilla/internal/store/oss"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Multipart struct {
	id     string
	key    string
	file   *os.File
	client *Client
	mu     sync.Mutex
}

func (m *Multipart) ID() string {
	return m.id
}

func (m *Multipart) List() (parts []oss.Part, err error) {
	return m.client.listParts(m.key, m.id)
}

func (m *Multipart) Write(number int64, reader io.Reader) (etag string, err error) {
	if len(m.id) < 1 {
		err = ErrNotMultipartMode
		return
	}
	if number < int64(oss.MinPartNumber) || number > int64(oss.MaxPartNumber) {
		err = ErrInvalidPartNumber
		return
	}
	return m.client.writePart(m.key, m.id, number, reader)
}

func (m *Multipart) Abort() (err error) {
	return m.client.abortParts(m.key, m.id)
}

func (m *Multipart) Merge(parts ...oss.Part) (checksum string, err error) {
	if len(parts) > 0 {
		prefix := m.client.buildPartKeyPath(m.key, m.id)
		var file *os.File
		key := filepath.Join(m.client.workdir, m.key)
		if file, err = m.client.safeOpen(key); err != nil {
			return
		}
		defer func() {
			_ = file.Close()
			var fi os.FileInfo
			var innerErr error
			if fi, innerErr = os.Stat(key); innerErr == nil {
				// cache checksum.
				_ = m.client.summarizer.set(m.client.buildKeyPath(key), oss.Summary{
					Etag:    checksum,
					Size:    fi.Size(),
					ModTime: fi.ModTime().UnixMilli(),
				})
			}
		}()
		h5 := md5.New()
		for _, p := range parts {
			err = m.appendPart(file, filepath.Join(prefix, fmt.Sprintf("%d", p.Number)), h5)
			if err != nil {
				return
			}
		}
		checksum = fmt.Sprintf("%s-%d", hex.EncodeToString(h5.Sum(nil)), len(parts))
		if m.client.fsync {
			err = file.Sync()
		}
		if err == nil {
			_ = os.RemoveAll(prefix)
		}
	}
	return
}

func (c *Client) abortParts(key string, multipartId string) (err error) {
	return os.RemoveAll(c.buildPartKeyPath(key, multipartId))
}

func (c *Client) writePart(key string, id string, number int64, reader io.Reader) (etag string, err error) {
	var f *os.File
	prefix := c.buildPartKeyPath(key, id)
	if f, err = c.safeOpen(filepath.Join(prefix, fmt.Sprintf("%d", number))); err == nil {
		defer f.Close()
		h5 := md5.New()
		if _, err = io.Copy(f, io.TeeReader(reader, h5)); err == nil {
			if c.fsync {
				err = f.Sync()
			}
		}
		etag = hex.EncodeToString(h5.Sum(nil))
	}
	return
}

func (c *Client) listParts(key string, uid string) (parts []oss.Part, err error) {
	prefix := c.buildPartKeyPath(key, uid)
	var infos []os.FileInfo
	if infos, err = ioutil.ReadDir(prefix); err != nil {
		return
	}
	for _, info := range infos {
		var num int
		if num, err = strconv.Atoi(info.Name()); err != nil {
			err = nil
			continue
		}
		parts = append(parts, oss.Part{
			Size:   info.Size(),
			Number: int64(num),
			Etag:   calculateHexMD5(filepath.Join(prefix, info.Name())),
		})
	}
	return
}

func calculateHexMD5(file string) (hexMD5Str string) {
	var f *os.File
	var err error
	if f, err = os.Open(file); err != nil {
		return ""
	}
	defer f.Close()
	hexMD5Str, _ = oss.CalculateHexMD5(f)
	return
}

func (m *Multipart) appendPart(writer io.Writer, partFile string, sumWriter io.Writer) (err error) {
	var pc io.ReadCloser
	var checksum string
	if pc, checksum, err = m.readPart(partFile); err != nil {
		return
	}
	defer pc.Close()
	var md5byte []byte
	if md5byte, err = hex.DecodeString(checksum); err != nil {
		return
	}
	_, _ = sumWriter.Write(md5byte)
	_, err = io.Copy(writer, pc)
	return
}

func (m *Multipart) readPart(absFile string) (readCloser io.ReadCloser, etag string, err error) {
	var f *os.File
	if f, err = os.Open(absFile); err != nil {
		return
	}
	// small file, it's ok.
	if etag, err = oss.CalculateHexMD5(f); err != nil {
		_ = f.Close()
		return
	}
	readCloser = f
	return
}
