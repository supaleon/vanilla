package file

import (
	"errors"
	"fmt"
	"github.com/supaleon/vanilla/internal/fs/oss"
	"io"
	"os"
	"sync"
)

const (
	minPartNumber = int64(1)
	maxPartNumber = int64(10000)
)

var (
	ErrNotMultipartMode  = errors.New("not multipart mode")
	ErrInvalidPartNumber = fmt.Errorf("part number must be in range %d-%d", minPartNumber, maxPartNumber)
)

type Object struct {
	key         string
	multipartId string
	metadata    oss.Metadata
	file        *os.File

	client *Client

	mu   sync.Mutex
	once sync.Once
}

func (o *Object) Name() string {
	return o.key
}

func (o *Object) OpenMultipart(id string) (multipart oss.Multipart, err error) {
	if id == "" {
		id = generateMultipartID()
	}
	multipart = &Multipart{
		id:     id,
		key:    o.key,
		file:   o.file,
		client: o.client,
	}
	return
}

func (o *Object) Equal(other oss.Object) bool {
	var err error
	var srcMD, otherMD oss.Metadata
	if srcMD, err = o.Stat(); err != nil {
		return false
	}
	if otherMD, err = other.Stat(); err != nil {
		return false
	}
	// < 5M
	if srcMD.Size() < oss.MinPartSize {
		return srcMD.Etag() != "" && srcMD.Etag() == otherMD.Etag()
	}
	if srcMD.MD5() != "" {
		return srcMD.MD5() == otherMD.MD5()
	}
	if srcMD.SHA256() != "" {
		return srcMD.SHA256() == srcMD.SHA256()
	}
	return srcMD.Etag() == otherMD.Etag()
}

func (o *Object) Stat() (md oss.Metadata, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	// fresh current object's metadata.
	o.metadata, err = o.client.Stat(o.key)
	return o.metadata, err
}

func (o *Object) Read() (reader io.ReadCloser, metadata oss.Metadata, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.client.Read(o.key)
}

func (o *Object) ReadRange(offset int64, size int64) (readCloser io.ReadCloser, md oss.Metadata, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if size < 1 {
		err = errors.New("size must bigger than zero")
		return
	}
	if o.file == nil {
		if o.file, err = os.Open(o.client.buildKeyPath(o.key)); err != nil {
			return
		}
	}
	var fi os.FileInfo
	if fi, err = o.file.Stat(); err != nil {
		return
	}
	readCloser = &rangeCloser{
		SectionReader: io.NewSectionReader(o.file, offset, size),
		file:          o.file,
	}
	mt := fi.ModTime()
	meta := &metadata{
		size:    fi.Size(),
		modTime: &mt,
		isDir:   false,
		client:  o.client,
		key:     o.key,
	}
	md = meta
	return
}

func (o *Object) Write(reader io.Reader) (etag string, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.client.Write(o.key, reader)
}

func (o *Object) Remove() (err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.client.Remove(o.key)
}

func (o *Object) Summarize(sum oss.Summary) (err error) {
	return o.client.Summarize(o.key, sum)
}

func (o *Object) Close() (err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.once.Do(func() {
		if o.file != nil {
			err = o.file.Close()
		}
	})
	return
}
