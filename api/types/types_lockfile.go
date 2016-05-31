package types

import (
	"os"
	"sync"
	"time"
)

// LockFile is a file-based lock that implements the sync.Locker interface to
// enable portable, interprocess synchronization. The calling process is
// responsible for creating the filetree structure at which the lock file is
// located and for ensuring the correct permissions for creating the lock file.
type LockFile struct {
	sync.Mutex
	path string
}

// NewLockFile returns a new LockFile object.
func NewLockFile(path string) *LockFile {
	return &LockFile{
		Mutex: sync.Mutex{},
		path:  path,
	}
}

// Lock locks the lock file. If the lock is already in use, the calling
// goroutine blocks until the lock file is available.
func (l *LockFile) Lock() {
	l.Mutex.Lock()
	for {
		f, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			continue
		}
		f.Close()
		return
	}
}

// Unlock unlocks the lock file. It is a run-time error if the lock file is
// not locked on entry to Unlock.
//
// A locked LockFile is not associated with a particular goroutine. It is
// allowed for one goroutine to lock a Mutex and then arrange for another
// goroutine to unlock it.
func (l *LockFile) Unlock() {
	l.Mutex.Unlock()
	os.RemoveAll(l.path)
}
