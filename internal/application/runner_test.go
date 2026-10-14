package application_test

import (
	"context"
	"gitea-exporter/internal/application"
	"gitea-exporter/internal/domain"
	"testing"
	"time"
)

func TestRunnerCollectPopulatesStore(t *testing.T) {
	source := &fakeRepository{
		organizations: []domain.Organization{{Name: "acme"}},
		members:       map[string][]domain.User{"acme": {{Name: "alice"}}},
		repositories:  map[string][]domain.Repository{"acme": {{Name: "repo1"}}},
		pullRequests:  map[string][]domain.PullRequest{},
		reposCalls:    map[string]int{},
	}
	service := application.NewService(nil)
	target := application.NewTarget("git-ext", source, service)
	runner := application.NewRunner(time.Hour, target)

	runner.Collect(context.Background())

	snapshot := target.Snapshot()
	if snapshot == nil {
		t.Fatal("snapshot not stored")
	}
	if len(snapshot.Orgs) != 1 {
		t.Fatalf("unexpected orgs: %+v", snapshot.Orgs)
	}
}
