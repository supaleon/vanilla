package file

import (
	"sync"
	"time"
)

type metadata struct {
	key     string
	size    int64
	modTime *time.Time
	isDir   bool
	md5     string
	sha256  string
	etag    string
	client  *Client
	mu      sync.Mutex
}

func (m *metadata) Size() int64 {
	return m.size
}

func (m *metadata) ModTime() *time.Time {
	return m.modTime
}

func (m *metadata) IsDir() bool {
	return m.isDir
}

// MD5 parses object hex md5 summary from s3cmd user metadata X-Amz-Meta-S3cmd-Attrs
func (m *metadata) MD5() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.md5 == "" {
		m.client.summarize(m.key, m)
	}
	return m.md5
}

// SHA256 parses object hex sha256 summary from minio mc user metadata X-Amz-Meta-Sha256
func (m *metadata) SHA256() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sha256 == "" {
		m.client.summarize(m.key, m)
	}
	return m.sha256
}

// Etag calculates object hex md5 if object smaller than the oss.PartSize,
// or calculates object calculateHexMD5=hex(md5(md5(part1)+md5(partN)...))-{parts count}.
func (m *metadata) Etag() string {
	if m.etag == "" {
		m.client.summarize(m.key, m)
	}
	return m.etag
}
