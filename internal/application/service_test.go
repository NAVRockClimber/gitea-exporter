package application_test

import (
	"context"
	"testing"
	"time"

	"gitea-exporter/internal/application"
	"gitea-exporter/internal/domain"
)

type fakeRepository struct {
	organizations []domain.Organization
	members       map[string][]domain.User
	repositories  map[string][]domain.Repository
	pullRequests  map[string][]domain.PullRequest

	reposCalls map[string]int
}

func (fake *fakeRepository) ListOrgs(ctx context.Context) ([]domain.Organization, error) {
	return fake.organizations, nil
}

func (fake *fakeRepository) ListMembers(ctx context.Context, organization string) ([]domain.User, error) {
	return fake.members[organization], nil
}

func (fake *fakeRepository) ListRepos(ctx context.Context, organization string) ([]domain.Repository, error) {
	fake.reposCalls[organization]++
	return fake.repositories[organization], nil
}

func (fake *fakeRepository) ListOpenPRs(ctx context.Context, organization, repository string) ([]domain.PullRequest, error) {
	return fake.pullRequests[organization+"/"+repository], nil
}

func TestCollectBuildsSnapshotAndExcludes(t *testing.T) {
	createdAt := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

	fake := &fakeRepository{
		organizations: []domain.Organization{{Name: "acme"}, {Name: "skip-me"}},
		members: map[string][]domain.User{
			"acme":    {{Name: "alice"}},
			"skip-me": {{Name: "bob"}},
		},
		repositories: map[string][]domain.Repository{
			"acme":    {{Name: "repo1"}},
			"skip-me": {{Name: "repo3"}},
		},
		pullRequests: map[string][]domain.PullRequest{
			"acme/repo1": {{ID: 1, Org: "acme", Repo: "repo1", Author: "alice", CreatedAt: &createdAt}},
		},
		reposCalls: map[string]int{},
	}

	service := application.NewService([]string{"skip-me"})

	snapshot, err := service.Collect(context.Background(), fake)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	if len(snapshot.Orgs) != 1 || snapshot.Orgs[0].Name != "acme" {
		t.Fatalf("unexpected orgs: %+v", snapshot.Orgs)
	}
	if _, exists := snapshot.Members["skip-me"]; exists {
		t.Fatal("excluded org should not appear in Members")
	}
	if _, exists := snapshot.Repos["skip-me"]; exists {
		t.Fatal("excluded org should not appear in Repos")
	}
	if len(snapshot.OpenPRs) != 1 {
		t.Fatalf("unexpected PRs: %+v", snapshot.OpenPRs)
	}

	if fake.reposCalls["acme"] != 1 {
		t.Fatalf("ListRepos called %d times for acme, want 1 (local reuse)", fake.reposCalls["acme"])
	}
}
