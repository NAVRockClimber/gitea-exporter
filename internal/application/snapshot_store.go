package application

import (
	"gitea-exporter/internal/domain"
	"sync/atomic"
)

// SnapshotStore holds the most recently collected Snapshot. It uses
// copy-on-write: Set publishes a fresh snapshot, and readers load a
// stable pointer to the current one, lock-free.
//
// Readers must treat the returned snapshot as read-only: its maps and
// slices are shared, not copied.
type SnapshotStore struct {
	current atomic.Pointer[domain.Snapshot]
}

func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{}
}

// Set publishes snapshot as the current value. It must not be mutated
// afterward.
func (store *SnapshotStore) Set(snapshot domain.Snapshot) {
	store.current.Store(&snapshot)
}

// Get returns a pointer to the current snapshot, or nil if none has
// been published yet.
func (store *SnapshotStore) Get() *domain.Snapshot {
	return store.current.Load()
}
