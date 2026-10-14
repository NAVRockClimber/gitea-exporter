package application_test

import (
	"gitea-exporter/internal/application"
	"gitea-exporter/internal/domain"
	"testing"
)

func TestSnapshotStoreSetGet(t *testing.T) {
	store := application.NewSnapshotStore()

	snapshot := domain.Snapshot{
		Orgs: []domain.Organization{{Name: "acme"}},
	}
	store.Set(snapshot)

	got := store.Get()
	if got == nil {
		t.Fatal("Get returned nil after Set")
	}
	if len(got.Orgs) != 1 || got.Orgs[0].Name != "acme" {
		t.Fatalf("unexpected snapshot: %+v", got)
	}
}

// TestSnapshotStoreConcurrentAccess proves the store is safe when a
// writer and a reader run at the same time. Run with -race to check.
func TestSnapshotStoreConcurrentAccess(t *testing.T) {
	store := application.NewSnapshotStore()
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 1000; i++ {
			store.Set(domain.Snapshot{Orgs: []domain.Organization{{Name: "acme"}}})
		}
	}()

	for i := 0; i < 1000; i++ {
		_ = store.Get()
	}
	<-done
}
