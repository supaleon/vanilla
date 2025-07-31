package oss

import (
	"io"
	"reflect"
	"testing"
)

func TestCalculateSha256(t *testing.T) {
	type args struct {
		reader io.ReadSeeker
	}
	tests := []struct {
		name         string
		args         args
		wantChecksum []byte
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChecksum, err := CalculateSha256(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateSha256() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotChecksum, tt.wantChecksum) {
				t.Errorf("CalculateSha256() = %v, want %v", gotChecksum, tt.wantChecksum)
			}
		})
	}
}

func TestCalculateMD5(t *testing.T) {
	type args struct {
		reader io.ReadSeeker
	}
	tests := []struct {
		name         string
		args         args
		wantChecksum []byte
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChecksum, err := CalculateMD5(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateMD5() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotChecksum, tt.wantChecksum) {
				t.Errorf("CalculateMD5() = %v, want %v", gotChecksum, tt.wantChecksum)
			}
		})
	}
}

func TestCalculateBase64MD5(t *testing.T) {
	type args struct {
		reader io.ReadSeeker
	}
	tests := []struct {
		name         string
		args         args
		wantChecksum string
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChecksum, err := CalculateBase64MD5(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateBase64MD5() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChecksum != tt.wantChecksum {
				t.Errorf("CalculateBase64MD5() = %v, want %v", gotChecksum, tt.wantChecksum)
			}
		})
	}
}

func TestCalculateHexMD5(t *testing.T) {
	type args struct {
		reader io.ReadSeeker
	}
	tests := []struct {
		name         string
		args         args
		wantChecksum string
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChecksum, err := CalculateHexMD5(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateHexMD5() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChecksum != tt.wantChecksum {
				t.Errorf("CalculateHexMD5() = %v, want %v", gotChecksum, tt.wantChecksum)
			}
		})
	}
}

func TestCalculateHexSha256(t *testing.T) {
	type args struct {
		reader io.ReadSeeker
	}
	tests := []struct {
		name         string
		args         args
		wantChecksum string
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChecksum, err := CalculateHexSha256(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateHexSha256() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChecksum != tt.wantChecksum {
				t.Errorf("CalculateHexSha256() = %v, want %v", gotChecksum, tt.wantChecksum)
			}
		})
	}
}

func TestEncodePart(t *testing.T) {
	type args struct {
		part Part
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodePart(tt.args.part); got != tt.want {
				t.Errorf("EncodePart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodePart(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name     string
		args     args
		wantPart Part
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPart, err := DecodePart(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodePart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPart, tt.wantPart) {
				t.Errorf("DecodePart() = %v, want %v", gotPart, tt.wantPart)
			}
		})
	}
}

func TestSplitObjectToFixedCountParts(t *testing.T) {
	type args struct {
		size       int64
		partCounts int64
	}
	tests := []struct {
		name      string
		args      args
		wantParts []Part
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotParts, err := SplitObjectToFixedCountParts(tt.args.size, tt.args.partCounts)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitObjectToFixedCountParts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotParts, tt.wantParts) {
				t.Errorf("SplitObjectToFixedCountParts() = %v, want %v", gotParts, tt.wantParts)
			}
		})
	}
}

func TestSplitObjectToFixedSizeParts(t *testing.T) {
	type args struct {
		size     int64
		partSize int64
	}
	tests := []struct {
		name      string
		args      args
		wantParts []Part
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotParts, err := SplitObjectToFixedSizeParts(tt.args.size, tt.args.partSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitObjectToFixedSizeParts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotParts, tt.wantParts) {
				t.Errorf("SplitObjectToFixedSizeParts() = %v, want %v", gotParts, tt.wantParts)
			}
		})
	}
}
