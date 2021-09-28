package filelock

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

var ErrAlreadyLocked = errors.New("locked by another process")

type FileLock struct {
	path string
	file *os.File
}

func Lock(filePath string) (*FileLock, error) {
	if exists, err := Exists(filePath); err != nil || exists {
		return nil, ErrAlreadyLocked
	}
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to mkdir: %w", err)
	}
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return nil, ErrAlreadyLocked
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	if err := lock(f); err != nil {
		var pid int
		if _, err = fmt.Fscanf(f, "%d\n", &pid); err == nil {
			return nil, fmt.Errorf("%w: locked by %v", err, pid)
		}
		return nil, ErrAlreadyLocked
	}

	f.Seek(0, 0)
	if n, err := fmt.Fprintf(f, "%d\n", os.Getpid()); err == nil {
		f.Truncate(int64(n))
	}

	return &FileLock{path: filePath, file: f}, nil
}

func (fl *FileLock) Unlock() error {
	if err := unlock(fl.file); err != nil {
		return fmt.Errorf("failed to unlock the lock file: %w", err)
	}
	if err := fl.file.Close(); err != nil {
		return fmt.Errorf("failed to close the lock file: %w", err)
	}
	if err := os.Remove(fl.file.Name()); err != nil {
		return fmt.Errorf("failed to remove the lock file: %w", err)
	}
	return nil
}

func (fl *FileLock) Close() error {
	return fl.Unlock()
}

func Exists(filePath string) (bool, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	var pid int
	if _, err = fmt.Fscanf(f, "%d\n", &pid); err == nil {
		log.Printf("lock %v in use by %v\n", filepath.Base(f.Name()), pid)
	}

	return true, nil
}

// use fcntl POSIX locks for the most consistent behavior across platforms, and
// hopefully some campatibility over NFS and CIFS.
// hashicorp/terraform/internal/states/statemgr/filesystem_lock_unix.go
func lock(f *os.File) error {
	log.Printf("locking %s using fcntl flock", f.Name())
	flock := &syscall.Flock_t{
		Type:   syscall.F_RDLCK | syscall.F_WRLCK,
		Pid:    int32(os.Getpid()),
		Whence: int16(io.SeekStart),
		Start:  0,
		Len:    0,
	}

	return syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, flock)
}

func unlock(f *os.File) error {
	log.Printf("unlocking %s using fcntl flock", f.Name())
	flock := &syscall.Flock_t{
		Type:   syscall.F_UNLCK,
		Pid:    int32(os.Getpid()),
		Whence: int16(io.SeekStart),
		Start:  0,
		Len:    0,
	}

	return syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, flock)
}
