package oss

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ErrInvalidEncodingPart       = errors.New("invalid encoding part string")
	ErrInvalidEncodingPartNumber = errors.New("invalid encoding part number")
	ErrInvalidEncodingPartSize   = errors.New("invalid encoding part size")
	ErrInvalidEncodingPartOffset = errors.New("invalid encoding part offset")
	ErrInvalidPartSize           = errors.New("file size must greater than part size")
	ErrInvalidPartCounts         = errors.New("file size must greater than part counts")
)

type MD5Reader interface {
	io.Reader
	MD5() []byte
}

// CalculateSha256 compute the content's sha256 checksum based on io.ReadSeeker
// and then move the whence to start.
func CalculateSha256(reader io.ReadSeeker) (checksum []byte, err error) {
	h := sha256.New()
	if _, err = io.Copy(h, reader); err == nil {
		checksum = h.Sum(nil)
	}
	// reset whence.
	_, err = reader.Seek(0, io.SeekStart)
	return
}

// CalculateMD5 compute the content's md5 checksum based on io.ReadSeeker
// and then move the whence to start.
func CalculateMD5(reader io.ReadSeeker) (checksum []byte, err error) {
	h := md5.New()
	if _, err = io.Copy(h, reader); err == nil {
		checksum = h.Sum(nil)
	}
	// reset whence.
	_, err = reader.Seek(0, io.SeekStart)
	return
}

// CalculateBase64MD5 compute the content's md5 checksum in base64 based on io.ReadSeeker
// and then move the whence to start.
func CalculateBase64MD5(reader io.ReadSeeker) (checksum string, err error) {
	var buf []byte
	if buf, err = CalculateMD5(reader); err != nil {
		checksum = base64.StdEncoding.EncodeToString(buf)
	}
	return
}

func CalculateHexMD5(reader io.ReadSeeker) (checksum string, err error) {
	var buf []byte
	if buf, err = CalculateMD5(reader); err == nil {
		checksum = hex.EncodeToString(buf)
	}
	return
}

func CalculateHexSha256(reader io.ReadSeeker) (checksum string, err error) {
	var buf []byte
	if buf, err = CalculateSha256(reader); err == nil {
		checksum = hex.EncodeToString(buf)
	}
	return
}

func EncodePart(part Part) string {
	return fmt.Sprintf("%d-%d-%d", part.Number, part.Size, part.Offset)
}

func DecodePart(name string) (part Part, err error) {
	basename := filepath.Base(name)
	segments := strings.Split(basename, "-")
	if len(segments) != 3 {
		err = ErrInvalidEncodingPart
		return
	}
	var size, offset int64
	var num int64
	num, err = strconv.ParseInt(segments[0], 10, 64)
	if err != nil {
		err = ErrInvalidEncodingPartNumber
		return
	}
	partNumber := num

	size, err = strconv.ParseInt(segments[1], 10, 64)
	if err != nil {
		err = ErrInvalidEncodingPartSize
		return
	}
	partSize := size

	offset, err = strconv.ParseInt(segments[2], 10, 64)
	if err != nil {
		err = ErrInvalidEncodingPartOffset
		return
	}
	partOffset := offset
	part = Part{
		Size:   partSize,
		Number: partNumber,
		Offset: partOffset,
	}
	return
}

// SplitObjectToFixedCountParts split file to fixed counts parts.
func SplitObjectToFixedCountParts(size int64, partCounts int64) (parts []Part, err error) {
	if size <= partCounts {
		err = ErrInvalidPartCounts
		return
	}
	partSize := size / partCounts
	lastPartSize := partSize
	if size%partCounts > 0 {
		lastPartSize = size % partCounts
	}

	for i := int64(0); i < partCounts; i++ {
		part := Part{}
		part.Number = i + 1
		part.Offset = i * partSize
		if i == partCounts-1 {
			part.Size = lastPartSize
		} else {
			part.Size = partSize
		}
		parts = append(parts, part)
	}
	return
}

// SplitObjectToFixedSizeParts split file to fixed size parts, except the last part.
func SplitObjectToFixedSizeParts(size int64, partSize int64) (parts []Part, err error) {
	if size <= partSize {
		err = ErrInvalidPartSize
		return
	}
	partCount := size / partSize
	if size%partSize > 0 {
		partCount = partCount + 1
	}
	for i := int64(0); i < partCount; i++ {
		part := Part{}
		part.Number = i + 1
		part.Offset = i * partSize
		if i == partCount-1 {
			part.Size = size % partSize
		} else {
			part.Size = partSize
		}
		parts = append(parts, part)
	}
	return
}
