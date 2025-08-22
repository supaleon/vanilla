package file

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/supaleon/vanilla/internal/fs/file"
	"github.com/supaleon/vanilla/internal/fs/oss"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	tempPath = "fstemp"
)

var (
	ErrInvalidTempDir = errors.New("invalid temp dir")
	ErrInvalidKey     = errors.New("invalid file name")
)

type Option func(*Client)

type Client struct {
	// root dir for storage files.
	workdir string
	// min part size for multipart.
	partSize int64
	// temp dir provide a dir to store incomplete file parts and checksum cache files.
	tempDir string
	// ignoreFunc provide a func to ignore special files when ReadDir or Iterate dir.
	ignoreFunc func(key string) bool
	// force flush file content & metadata to stable storage to ensure data safe.
	fsync bool
	// calculate file checksum.
	enableChecksum bool
	// force calculate file content checksum even if it's a big file.
	forceChecksum bool
	// checksum cache manager.
	summarizer *cache
}

func WithPartSize(partSize int64) Option {
	return func(client *Client) {
		if partSize < oss.MinPartSize {
			partSize = oss.MinPartSize
		}
		client.partSize = partSize
	}
}

func WithTempDir(tempDir string) Option {
	return func(client *Client) {
		client.tempDir = tempDir
	}
}

func WithFsync() Option {
	return func(client *Client) {
		client.fsync = true
	}
}

func WithChecksum(force bool) Option {
	return func(client *Client) {
		if client.summarizer != nil {
			return
		}
		client.summarizer = &cache{
			summaries: make(map[string]oss.Summary),
			workdir:   filepath.Join(client.tempDir, summaryPrefix),
		}
		client.enableChecksum = true
		client.forceChecksum = force
	}
}

func WithIgnoreFunc(fn func(key string) bool) Option {
	return func(client *Client) {
		client.ignoreFunc = fn
	}
}

func New(workdir string, options ...Option) (c *Client, err error) {
	if workdir, err = file.Abs(workdir); err != nil {
		return
	}
	c = &Client{
		workdir: workdir,
	}
	if len(options) > 0 {
		for _, op := range options {
			op(c)
		}
	}
	if c.tempDir == "" {
		c.tempDir = filepath.Join(os.TempDir(), tempPath)
	}
	if c.tempDir, err = file.Abs(c.tempDir); err != nil {
		err = ErrInvalidTempDir
		return
	}
	if c.enableChecksum && c.summarizer != nil {
		c.summarizer.workdir = filepath.Join(c.tempDir, summaryPrefix)
	}
	return
}

func (c *Client) Name() string {
	return "vfs"
}

func (c *Client) Open(key string) (object oss.Object, err error) {
	if key == "" || strings.HasSuffix(key, PathSeparator) {
		err = ErrInvalidKey
		return
	}
	object = &Object{key: key, client: c}
	return
}

func (c *Client) Stat(key string) (meta oss.Metadata, err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(c.buildKeyPath(key)); err == nil {
		mt := fi.ModTime()
		m := &metadata{
			size:    fi.Size(),
			modTime: &mt,
			isDir:   fi.IsDir(),
			client:  c,
			key:     key,
		}
		if fi.IsDir() {
			m.size, err = c.sumDirSize(key)
		}
		meta = m
	}
	return
}

func (c *Client) sumDirSize(key string) (size int64, err error) {
	err = filepath.Walk(c.buildKeyPath(key), func(path string, info fs.FileInfo, inErr error) error {
		if inErr != nil {
			return inErr
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return
}

// Iterate provide a file Iterator to iterate file later.
// similar to filepath.Walk
func (c *Client) Iterate(prefix string) (iterator oss.Iterator, err error) {
	iterator = &objectIterator{
		key:        prefix,
		workdir:    c.workdir,
		client:     c,
		ignoreFunc: c.ignore,
		recursive:  true,
	}
	return
}

// ReadDir open an object descriptor Iterator of prefix given.
// similar to os.ReadDir
func (c *Client) ReadDir(key string) (it oss.Iterator, err error) {
	it = &objectIterator{
		key:        key,
		workdir:    c.workdir,
		client:     c,
		ignoreFunc: c.ignore,
	}
	return
}

// Read open the key as io.ReadCloser
func (c *Client) Read(key string) (readCloser io.ReadCloser, metadata oss.Metadata, err error) {
	return c.read(key)
}

func (c *Client) read(key string) (readCloser io.ReadCloser, meta oss.Metadata, err error) {
	var f *os.File
	if f, err = os.Open(c.buildKeyPath(key)); err != nil {
		return
	}
	var fi os.FileInfo
	if fi, err = f.Stat(); err != nil {
		return
	}
	readCloser = f
	mt := fi.ModTime()
	meta = &metadata{
		size:    fi.Size(),
		modTime: &mt,
		isDir:   false,
		client:  c,
		key:     key,
	}
	return
}

// Write write the content to the key based Object.
func (c *Client) Write(key string, reader io.Reader) (etag string, err error) {
	var f *os.File
	absKey := c.buildKeyPath(key)
	if f, err = c.safeOpen(absKey); err == nil {
		if c.enableChecksum {
			var h256Str string
			defer func() {
				_ = f.Close()
				var fi os.FileInfo
				var innerErr error
				if fi, innerErr = os.Stat(absKey); innerErr == nil {
					// cache checksum.
					_ = c.summarizer.set(c.buildKeyPath(key), oss.Summary{
						Etag:    etag,
						MD5:     etag,
						SHA256:  h256Str,
						Size:    fi.Size(),
						ModTime: fi.ModTime().UnixMilli(),
					})
				}
			}()
			h5 := md5.New()
			h256 := sha256.New()
			mw := io.MultiWriter(h256, h5)
			teeReader := io.TeeReader(reader, mw)
			if _, err = io.Copy(f, teeReader); err == nil {
				if c.fsync {
					err = f.Sync()
				}
			}
			if err == nil {
				etag = hex.EncodeToString(h5.Sum(nil))
				h256Str = hex.EncodeToString(h256.Sum(nil))
			}
		} else {
			_, err = io.Copy(f, reader)
		}
	}
	return
}

// Remove remove the key based Object.
func (c *Client) Remove(key string) (err error) {
	// remove incomplete multipart session
	_ = os.RemoveAll(c.buildPartKeyPath(key, ""))
	// remove incomplete checksum file
	if c.enableChecksum {
		_ = c.summarizer.del(c.buildKeyPath(key))
	}
	// remove file.
	return os.Remove(c.buildKeyPath(key))
}

// RemoveAll remove all dir and files recursion.
// NB: TODO unsafe operation.
func (c *Client) RemoveAll(pathInWorkdir string) (err error) {
	return os.RemoveAll(c.buildKeyPath(pathInWorkdir))
}

func (c *Client) Summarize(key string, summary oss.Summary) (err error) {
	return c.summarizer.set(c.buildKeyPath(key), summary)
}

func (c *Client) safeOpen(file string) (f *os.File, err error) {
	if _, err = os.Stat(file); err != nil {
		if !os.IsNotExist(err) {
			return
		}
		err = os.MkdirAll(filepath.Dir(file), 0700)
	}
	if err == nil {
		return os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0700)
	}
	return
}

func (c *Client) ignore(key string) bool {
	if strings.HasPrefix(key, c.tempDir) {
		return true
	}
	if strings.HasPrefix(key, ".DS_Store") {
		return true
	}
	if strings.HasSuffix(key, "~") {
		return true
	}
	if strings.HasSuffix(key, ".swp") {
		return true
	}
	if c.ignoreFunc != nil {
		return c.ignoreFunc(key)
	}
	return false
}

func (c *Client) summarize(key string, md *metadata) {
	if !c.enableChecksum {
		return
	}
	absKey := c.buildKeyPath(key)
	// get from cache if the file size bigger than partSize.
	if !c.forceChecksum && md.size > c.partSize && c.partSize > 0 {
		var sum oss.Summary
		var sumErr error
		if sum, sumErr = c.summarizer.get(absKey, md.size, md.ModTime().UnixMilli()); sumErr == nil {
			md.md5 = sum.MD5
			md.etag = sum.Etag
			md.sha256 = sum.SHA256
		}
		return
	}
	// calculate summary real time.
	var f *os.File
	var err error
	if f, err = os.Open(absKey); err != nil {
		return
	}
	defer f.Close()
	var fi os.FileInfo
	// file content changed.
	if fi, err = f.Stat(); err != nil || fi.Size() != md.size || fi.ModTime().UnixMilli() != md.modTime.UnixMilli() {
		return
	}
	hashMD5 := md5.New()
	hashSHA256 := sha256.New()
	hashEtag := md5.New()
	tw := io.MultiWriter(hashMD5, hashSHA256)
	if fi.Size() > c.partSize && c.partSize >= oss.MinPartSize {
		var parts []oss.Part
		if parts, err = oss.SplitObjectToFixedSizeParts(fi.Size(), c.partSize); err != nil {
			return
		}
		for _, part := range parts {
			section := io.NewSectionReader(f, part.Offset, part.Size)
			subMD5 := md5.New()
			if _, err = io.Copy(tw, io.TeeReader(section, subMD5)); err != nil {
				return
			}
			hashEtag.Write(subMD5.Sum(nil))
		}
		md.md5 = hex.EncodeToString(hashMD5.Sum(nil))
		md.etag = fmt.Sprintf("%s-%d", hex.EncodeToString(hashEtag.Sum(nil)), len(parts))
	} else {
		if _, err = io.Copy(tw, f); err != nil {
			return
		}
		md.md5 = hex.EncodeToString(hashMD5.Sum(nil))
		md.etag = md.md5
	}
	md.sha256 = hex.EncodeToString(hashSHA256.Sum(nil))
	_ = c.summarizer.set(absKey, oss.Summary{
		Etag:    md.etag,
		SHA256:  md.sha256,
		MD5:     md.md5,
		Size:    md.size,
		ModTime: md.modTime.UnixMilli(),
	})
}
