package vfs

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/supaleon/vanilla/internal/store/oss"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	summaryPrefix  = "summaries"
	summaryMaxSize = 1024 * 1024
)

var (
	ErrSummaryNotExist = errors.New("summary not exist")
	ErrInvalidFile     = errors.New("invalid file")
)

// The cache provide 2 levels summary cache.
// NB: unsafe, but ok
type cache struct {
	workdir   string
	summaries map[string]oss.Summary

	mu sync.Mutex
}

func (c *cache) get(key string, size int64, modTime int64) (sum oss.Summary, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if key == "" {
		err = ErrInvalidFile
	}
	file := c.sha1file(key)
	var ok bool
	// lookup memory cache.
	if sum, ok = c.summaries[file]; ok {
		if sum.Size == size && sum.ModTime == modTime {
			return
		}
		delete(c.summaries, file)
	}
	// lookup file cache
	var f *os.File
	if f, err = os.Open(file); err != nil {
		return
	}
	defer f.Close()
	var buf []byte
	// NB: limit read all size.
	if buf, err = io.ReadAll(io.LimitReader(f, summaryMaxSize)); err != nil {
		return
	}
	if err = json.Unmarshal(buf, &sum); err != nil {
		return
	}
	if sum.Size == size && sum.ModTime == modTime {
		c.summaries[file] = sum
		return
	}
	err = ErrSummaryNotExist
	return
}

func (c *cache) set(key string, sum oss.Summary) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if key == "" {
		err = ErrInvalidFile
		return
	}
	file := c.sha1file(key)
	c.summaries[file] = sum
	return c.save(file, sum)
}

func (c *cache) save(file string, sum oss.Summary) (err error) {
	var buf []byte
	if buf, err = json.Marshal(sum); err != nil {
		return
	}
	_, err = os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
		err = os.MkdirAll(filepath.Dir(file), 0700)
	}
	if err == nil {
		var f *os.File
		if f, err = os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0700); err != nil {
			return
		}
		defer f.Close()
		_, err = io.Copy(f, bytes.NewBuffer(buf))
		return
	}
	return
}

func (c *cache) del(key string) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if key == "" {
		err = ErrInvalidFile
		return
	}
	return os.Remove(c.sha1file(key))
}

func (c *cache) sha1file(key string) (path string) {
	hashSHA1 := sha1.New()
	hashSHA1.Write([]byte(key))
	s := hex.EncodeToString(hashSHA1.Sum(nil))
	return filepath.Join(c.workdir, s[0:2], s[2:])
}
