package file

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func HomeDir() string {
	if runtime.GOOS == "windows" {
		// First prefer the HOME environmental variable
		if home := os.Getenv("HOME"); len(home) > 0 {
			if _, err := os.Stat(home); err == nil {
				return home
			}
		}
		if homeDrive, homePath := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); len(homeDrive) > 0 && len(homePath) > 0 {
			homeDir := homeDrive + homePath
			if _, err := os.Stat(homeDir); err == nil {
				return homeDir
			}
		}
		if userProfile := os.Getenv("USERPROFILE"); len(userProfile) > 0 {
			if _, err := os.Stat(userProfile); err == nil {
				return userProfile
			}
		}
	}
	return os.Getenv("HOME")
}

func Segments(filename string) (dir, name, extension string) {
	if strings.HasSuffix(filename, string(os.PathSeparator)) {
		dir = strings.TrimSuffix(filename, string(os.PathSeparator))
		return
	}
	if strings.Contains(filename, string(os.PathSeparator)) {
		dir = filename[:strings.LastIndex(filename, string(os.PathSeparator))]
	}
	name = path.Base(filename)
	if strings.Contains(name, ".") {
		segments := strings.Split(name, ".")
		extension = name[len(segments[0])+1:]
		name = segments[0]
	}
	return
}

func Abs(filename string) (file string, err error) {
	if strings.Contains(filename, "~") {
		file = filepath.Join(HomeDir(), strings.Replace(filename, "~", "", -1))
		return
	}
	return filepath.Abs(filename)
}

func TempDir(rootDir, subDirPrefix string) (dirname string, err error) {
	if rootDir == "" {
		rootDir = os.TempDir()
	}
	if err = os.MkdirAll(rootDir, 0700); err == nil {
		return os.MkdirTemp(rootDir, subDirPrefix)
	}
	return
}

func FormatSize(size int64) (s string) {
	if size < 1024 {
		return fmt.Sprintf("%.2fB", float64(size)/float64(1))
	} else if size < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(size)/float64(1024))
	} else if size < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(size)/float64(1024*1024))
	} else if size < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(size)/float64(1024*1024*1024))
	} else if size < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(size)/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB", float64(size)/float64(1024*1024*1024*1024*1024))
	}
}
