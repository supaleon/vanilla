package file

import (
	"github.com/supaleon/vanilla/internal/fs/oss"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type objectIterator struct {
	client     *Client
	workdir    string
	key        string
	recursive  bool
	ignoreFunc func(key string) bool

	done  bool
	err   error
	objCh chan oss.Object

	once sync.Once
}

func (i *objectIterator) run() {
	var err error
	if i.recursive {
		err = i.walk(filepath.Join(i.workdir, i.key))
	} else {
		err = i.read(filepath.Join(i.workdir, i.key))
	}
	i.err = err
	close(i.objCh)
}

func (i *objectIterator) Next() (obj oss.Object, err error) {
	if i.done {
		err = i.err
		return
	}
	i.once.Do(func() {
		i.objCh = make(chan oss.Object, 1)
		go i.run()
	})
	select {
	case obj, _ = <-i.objCh:
		err = i.err
		return
	}
}

func (i *objectIterator) read(prefix string) (err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(prefix); err != nil {
		return
	}
	if !fi.IsDir() {
		mt := fi.ModTime()
		i.objCh <- &Object{
			key: i.key,
			metadata: &metadata{
				size:    fi.Size(),
				modTime: &mt,
				isDir:   fi.IsDir(),
				client:  i.client,
				key:     i.key,
			},
			client: i.client,
		}
		return
	}

	var list []os.FileInfo
	if list, err = ioutil.ReadDir(prefix); err != nil {
		return
	}
	for _, item := range list {
		// ignore storage system files.
		if i.ignoreFunc(item.Name()) {
			continue
		}
		mt := item.ModTime()
		i.objCh <- &Object{
			key: filepath.Join(i.key, item.Name()),
			metadata: &metadata{
				size:    item.Size(),
				modTime: &mt,
				isDir:   item.IsDir(),
				client:  i.client,
				key:     filepath.Join(i.key, item.Name()),
			},
			client: i.client,
		}
	}
	return
}

func (i *objectIterator) walk(prefix string) (err error) {
	err = filepath.Walk(prefix,
		func(path string, info os.FileInfo, innerErr error) error {
			if info != nil {
				// ignore storage system files.
				if i.ignoreFunc(info.Name()) {
					return nil
				}
				if innerErr == nil && !info.IsDir() {
					mt := info.ModTime()
					key := filepath.Join(i.key, strings.TrimPrefix(path, prefix))
					i.objCh <- &Object{
						key: key,
						metadata: &metadata{
							size:    info.Size(),
							modTime: &mt,
							isDir:   false,
							client:  i.client,
							key:     key,
						},
						client: i.client,
					}
				}
			}
			return innerErr
		},
	)
	return
}
