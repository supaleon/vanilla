package vfs

import (
	"reflect"
	"sync"
	"testing"

	"github.com/supaleon/vanilla/internal/store/oss"
)

func Test_objectIterator_run(t *testing.T) {
	type fields struct {
		client     *Client
		workdir    string
		key        string
		recursive  bool
		ignoreFunc func(key string) bool
		done       bool
		err        error
		objCh      chan oss.Object
		once       sync.Once
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &objectIterator{
				client:     tt.fields.client,
				workdir:    tt.fields.workdir,
				key:        tt.fields.key,
				recursive:  tt.fields.recursive,
				ignoreFunc: tt.fields.ignoreFunc,
				done:       tt.fields.done,
				err:        tt.fields.err,
				objCh:      tt.fields.objCh,
				once:       tt.fields.once,
			}
			i.run()
		})
	}
}

func Test_objectIterator_Next(t *testing.T) {
	type fields struct {
		client     *Client
		workdir    string
		key        string
		recursive  bool
		ignoreFunc func(key string) bool
		done       bool
		err        error
		objCh      chan oss.Object
		once       sync.Once
	}
	tests := []struct {
		name    string
		fields  fields
		wantObj oss.Object
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &objectIterator{
				client:     tt.fields.client,
				workdir:    tt.fields.workdir,
				key:        tt.fields.key,
				recursive:  tt.fields.recursive,
				ignoreFunc: tt.fields.ignoreFunc,
				done:       tt.fields.done,
				err:        tt.fields.err,
				objCh:      tt.fields.objCh,
				once:       tt.fields.once,
			}
			gotObj, err := i.Next()
			if (err != nil) != tt.wantErr {
				t.Errorf("objectIterator.Next() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotObj, tt.wantObj) {
				t.Errorf("objectIterator.Next() = %v, want %v", gotObj, tt.wantObj)
			}
		})
	}
}

func Test_objectIterator_read(t *testing.T) {
	type fields struct {
		client     *Client
		workdir    string
		key        string
		recursive  bool
		ignoreFunc func(key string) bool
		done       bool
		err        error
		objCh      chan oss.Object
		once       sync.Once
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &objectIterator{
				client:     tt.fields.client,
				workdir:    tt.fields.workdir,
				key:        tt.fields.key,
				recursive:  tt.fields.recursive,
				ignoreFunc: tt.fields.ignoreFunc,
				done:       tt.fields.done,
				err:        tt.fields.err,
				objCh:      tt.fields.objCh,
				once:       tt.fields.once,
			}
			if err := i.read(tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("objectIterator.read() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_objectIterator_walk(t *testing.T) {
	type fields struct {
		client     *Client
		workdir    string
		key        string
		recursive  bool
		ignoreFunc func(key string) bool
		done       bool
		err        error
		objCh      chan oss.Object
		once       sync.Once
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &objectIterator{
				client:     tt.fields.client,
				workdir:    tt.fields.workdir,
				key:        tt.fields.key,
				recursive:  tt.fields.recursive,
				ignoreFunc: tt.fields.ignoreFunc,
				done:       tt.fields.done,
				err:        tt.fields.err,
				objCh:      tt.fields.objCh,
				once:       tt.fields.once,
			}
			if err := i.walk(tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("objectIterator.walk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
