package fsutil

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goyave.dev/goyave/v5/util/errors"
)

var contentTypeByExtension = map[string]string{
	".jsonld": "application/ld+json",
	".json":   "application/json",
	".js":     "text/javascript",
	".mjs":    "text/javascript",
	".css":    "text/css",
}

// GetFileExtension returns the last part of a file name.
// If the file doesn't have an extension, returns an empty string.
func GetFileExtension(filename string) string {
	index := strings.LastIndex(filename, ".")
	if index == -1 {
		return ""
	}
	return filename[index+1:]
}

// GetMIMEType get the mime type and size of the given file.
// This function calls `http.DetectContentType`. If the detected content type
// could not be determined or if it's a text file, `GetMIMEType` will attempt to
// detect the MIME type based on the file extension. The following extensions are
// supported:
//   - `.jsonld`: "application/ld+json"
//   - `.json`: "application/json"
//   - `.js` / `.mjs`: "text/javascript"
//   - `.css`: "text/css"
//
// If a specific MIME type cannot be determined, returns "application/octet-stream" as a fallback.
func GetMIMEType(filesystem fs.FS, file string) (contentType string, size int64, err error) {
	var f fs.File
	f, err = filesystem.Open(file)
	if err != nil {
		err = errors.NewSkip(err, 3)
		return
	}
	defer func() {
		errClose := f.Close()
		if err == nil && errClose != nil {
			err = errors.New(errClose)
		}
	}()

	var stat fs.FileInfo
	stat, err = f.Stat()
	if err != nil {
		err = errors.NewSkip(err, 3)
		return
	}

	size = stat.Size()

	buffer := make([]byte, 512)
	contentType = "application/octet-stream"

	if size != 0 {
		_, err = f.Read(buffer)
		if err != nil {
			err = errors.NewSkip(err, 3)
			return
		}

		contentType = http.DetectContentType(buffer)
	}

	if strings.HasPrefix(contentType, "application/octet-stream") || strings.HasPrefix(contentType, "text/plain") {
		for ext, t := range contentTypeByExtension {
			if strings.HasSuffix(file, ext) {
				tmp := t
				if i := strings.Index(contentType, ";"); i != -1 {
					tmp = t + contentType[i:]
				}
				contentType = tmp
				break
			}
		}
	}

	return
}

// FileExists returns true if the file at the given path exists and is readable.
// Returns false if the given file is a directory.
func FileExists(fs fs.StatFS, file string) bool {
	if stats, err := fs.Stat(file); err == nil {
		return !stats.IsDir()
	}
	return false
}

// IsDirectory returns true if the file at the given path exists, is a directory and is readable.
func IsDirectory(fs fs.StatFS, path string) bool {
	if stats, err := fs.Stat(path); err == nil {
		return stats.IsDir()
	}
	return false
}

func timestampFileName(name string) string {
	var prefix string
	var extension string
	index := strings.LastIndex(name, ".")
	if index == -1 {
		prefix = name
		extension = ""
	} else {
		prefix = name[:index]
		extension = name[index:]
	}
	return prefix + "-" + strconv.FormatInt(time.Now().UnixNano()/int64(time.Microsecond), 10) + extension
}

// An FS provides access to a hierarchical file system
// and implements `io/fs`'s `FS`, `ReadDirFS` and `StatFS` interfaces.
type FS interface {
	fs.ReadDirFS
	fs.StatFS
}

// A WorkingDirFS is a file system with a `Getwd()` method.
type WorkingDirFS interface {
	FS

	// Getwd returns a rooted path name corresponding to the
	// current directory. If the current directory can be
	// reached via multiple paths (due to symbolic links),
	// Getwd may return any one of them.
	Getwd() (dir string, err error)
}

// A MkdirFS is a file system with a `Mkdir()` and a `MkdirAll()` methods.
type MkdirFS interface {
	FS

	// MkdirAll creates a directory named path,
	// along with any necessary parents, and returns `nil`,
	// or else returns an error.
	// The permission bits perm (before umask) are used for all
	// directories that `MkdirAll` creates.
	// If path is already a directory, `MkdirAll` does nothing
	// and returns `nil`.
	MkdirAll(path string, perm fs.FileMode) error

	// Mkdir creates a new directory with the specified name and permission
	// bits (before umask).
	// If there is an error, it will be of type `*PathError`.
	Mkdir(path string, perm fs.FileMode) error
}

// A WritableFS is a file system with a `OpenFile()` method.
type WritableFS interface {

	// OpenFile is the generalized open call. It opens the named file with specified flag
	// (`O_RDONLY` etc.). If the file does not exist, and the `O_CREATE` flag
	// is passed, it is created with mode perm (before umask). If successful,
	// methods on the returned file can be used for I/O.
	// If there is an error, it will be of type `*PathError`.
	OpenFile(path string, flag int, perm fs.FileMode) (io.ReadWriteCloser, error)
}

// A RemoveFS is a file system with a `Remove()` and a `RemoveAll()` methods.
type RemoveFS interface {

	// Remove removes the named file or (empty) directory.
	// If there is an error, it will be of type `*PathError`.
	Remove(path string) error

	// RemoveAll removes path and any children it contains.
	// It removes everything it can but returns the first error
	// it encounters. If the path does not exist, `RemoveAll`
	// returns `nil` (no error).
	// If there is an error, it will be of type `*PathError`.
	RemoveAll(path string) error
}

// Embed is an extension of `embed.FS` implementing `fs.StatFS`.
type Embed struct {
	embed.FS
}

// Stat returns a FileInfo describing the file.
func (e Embed) Stat(name string) (fileinfo fs.FileInfo, err error) {
	f, err := e.FS.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		e := f.Close()
		if err == nil && e != nil {
			err = &fs.PathError{Op: "close", Path: name, Err: e}
		}
	}()

	fileinfo, err = f.Stat()
	return
}
