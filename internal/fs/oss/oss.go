package oss

import (
	"io"
	"time"
)

const (
	MaxSize       = 1024 * 1024 * 1024 * 1024 * 5
	MaxPartSize   = 1024 * 1024 * 1024 * 5
	MinPartSize   = 1024 * 1024 * 5
	MinPartNumber = 1
	MaxPartNumber = 10000
	PartSize      = 1024 * 1024 * 16
)

type Part struct {
	Size   int64  `json:"size" yaml:"size"`
	Number int64  `json:"number" yaml:"number"`
	Offset int64  `json:"offset" yaml:"offset"`
	Etag   string `json:"etag" yaml:"etag"`
}

//X-Amz-Meta-S3cmd-Attrs: atime:1645514736/ctime:1645514679/gid:20/gname:staff/md5:2120d00e7f7b2b9f8f291f0786c35932/mode:33188/mtime:1645514675/uid:501/uname:tomato

type Summary struct {
	Etag    string `json:"etag"`
	SHA256  string `json:"sha_256"`
	MD5     string `json:"md_5"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
}

type Summarizer interface {
	Summarize(Summary) error
}

type Metadata interface {
	Size() int64
	ModTime() *time.Time
	IsDir() bool
	// MD5 parses object hex md5 summary from s3cmd user metadata X-Amz-Meta-S3cmd-Attrs
	MD5() string
	// SHA256 parses hex sha256 summary from minio mc user metadata X-Amz-Meta-Sha256
	SHA256() string
	// Etag calculates object hex md5 by given part size.
	Etag() string
}

type Object interface {

	// Name returns the object relative key.
	Name() string

	// Stat reads object metadata.
	// similar to os.Stat
	Stat() (Metadata, error)

	// OpenMultipart open a multipart descriptor of a given key for later operation.
	// If the id is empty, create a new multipart session and generate an id for later operation.
	// If not exist, create a new one.
	// Similar to os.OpenFile
	OpenMultipart(id string) (multipart Multipart, err error)

	// Read reads object content to the io.ReadCloser.
	// similar to os.ReadFile
	Read() (readCloser io.ReadCloser, metadata Metadata, err error)

	// ReadRange reads range of object to the io.ReadCloser,
	// similar to os.ReadFile
	// The return objectMetadata is the object metadata, not range's metadata.
	ReadRange(offset int64, size int64) (readCloser io.ReadCloser, objectMetadata Metadata, err error)

	// Write Writes object content from the reader.
	// similar to os.WriteFile
	Write(reader io.Reader) (etag string, err error)

	// Removes remove the object of key given.
	// similar to os.Remove
	Remove() (err error)

	// Equal compare object checksum with other object given.
	Equal(Object) bool
}

type Multipart interface {
	ID() string
	List() (parts []Part, err error)
	Write(number int64, reader io.Reader) (etag string, err error)
	Merge(parts ...Part) (etag string, err error)
	Abort() (err error)
}

type Iterator interface {
	Next() (Object, error)
}

type Class interface {
	Name() string
}

type Liveness interface {
	// Alive return current storage liveness status.
	Alive() (yes bool, err error)
}

type OSS interface {
	// Open open an object descriptor of given key for later operation.
	// if not exist, create a new one.
	// similar to os.OpenFile
	Open(key string) (Object, error)

	Stat(key string) (Metadata, error)

	// Iterate provide an object descriptor Iterator to iterate object descriptor later.
	// similar to filepath.Walk
	Iterate(prefix string) (iterator Iterator, err error)

	// ReadDir open an object descriptor Iterator of prefix given.
	// similar to os.ReadDir
	ReadDir(prefix string) (iterator Iterator, err error)

	// Read reads object content of key given to the io.ReadCloser.
	// similar to os.ReadFile
	Read(key string) (readCloser io.ReadCloser, metadata Metadata, err error)

	// Write write object content of key given from the reader.
	// similar to os.WriteFile
	Write(key string, reader io.Reader) (etag string, err error)

	// Remove remove the object of key given.
	// similar to os.Remove
	Remove(key string) error

	// RemoveAll remove objects of prefix given.
	// similar to os.RemoveAll
	RemoveAll(prefix string) error
}
