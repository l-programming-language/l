//go:build darwin || linux

package phase1

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const RuntimeRegistryEnv = "P1_OWNERSHIP"

var ErrRuntimeAdmissionClosed = errors.New("runtime helper admission closed")

// RuntimeLifetime is written while admission and lifetime locks are held. Its
// unique file remains after exit; only a held lock authorizes fallback signals.
type RuntimeLifetime struct {
	Record, Label string
	PID, PGID     int
}

func ownershipLock(file *os.File, mode int, deadline time.Time) error {
	for {
		err := syscall.Flock(int(file.Fd()), mode|syscall.LOCK_NB)
		if err == nil || !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("ownership lock deadline: %w", err)
		}
		time.Sleep(time.Millisecond)
	}
}

// RegisterRuntimeHelper is the helper's first action. The same source is copied
// into the standalone test companion, so both sides use one ownership protocol.
// The returned file must stay open for the entire helper lifetime.
func RegisterRuntimeHelper(label string) (*os.File, RuntimeLifetime, error) {
	dir := os.Getenv(RuntimeRegistryEnv)
	var record RuntimeLifetime
	admission, err := os.OpenFile(filepath.Join(dir, "admission"), os.O_RDONLY, 0)
	if dir == "" || errors.Is(err, os.ErrNotExist) {
		if admission != nil {
			admission.Close() //nolint:errcheck // Preserve admission-closed status; this is a read-only descriptor.
		}
		return nil, record, ErrRuntimeAdmissionClosed
	}
	if err != nil {
		return nil, record, err
	}
	defer admission.Close() //nolint:errcheck // Releasing a read-only admission lock cannot affect the helper's recorded lifetime.
	if err := ownershipLock(admission, syscall.LOCK_SH, time.Now().Add(time.Second)); err != nil {
		return nil, record, err
	}
	state := make([]byte, 4)
	if n, err := admission.ReadAt(state, 0); err != nil || n != 4 || string(state) != "open" {
		return nil, record, ErrRuntimeAdmissionClosed
	}
	file, err := os.CreateTemp(dir, "alive-")
	if err != nil {
		return nil, record, err
	}
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close() //nolint:errcheck // Preserve the lock-acquisition failure while releasing the temporary record.
		return nil, record, err
	}
	pgid, err := syscall.Getpgid(0)
	if err == nil {
		record = RuntimeLifetime{filepath.Base(file.Name()), label, os.Getpid(), pgid}
		err = json.NewEncoder(file).Encode(record)
	}
	if err != nil {
		file.Close() //nolint:errcheck // Preserve the registration failure while releasing the incomplete record.
		return nil, record, err
	}
	return file, record, nil
}
