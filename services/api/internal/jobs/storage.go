package jobs

import (
	"errors"
	"os"
	"path/filepath"
)

// Reserve a whole job's scratch budget before starting it. Existing scratch is
// counted too, so parallel reservations are deliberately conservative.
func (m *Manager) reserveStorage() (func(), error) {
	if m.StorageBudget == 0 {
		return func() {}, nil
	}
	m.storageMu.Lock()
	defer m.storageMu.Unlock()
	var total int64
	err := filepath.WalkDir(m.Root, func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("unsafe storage entry")
		}
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			total += info.Size()
		}
		return nil
	})
	if err != nil || total+m.reservedBytes+m.MaxBytes > m.StorageBudget {
		return nil, errors.New("storage capacity reached")
	}
	m.reservedBytes += m.MaxBytes
	return func() {
		m.storageMu.Lock()
		m.reservedBytes -= m.MaxBytes
		m.storageMu.Unlock()
	}, nil
}
