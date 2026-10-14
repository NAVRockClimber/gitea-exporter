package application

import (
	"context"
	"gitea-exporter/internal/domain"
)

// Service turns Gitea data into a domain.Snapshot. It applies the
// org-exclusion policy and fetches each org's members, repos, and open
// PRs exactly once per Collect call.
type Service struct {
	excludedOrganizations map[string]struct{}
}

// everything collected for one organization
type organizationData struct {
	members      []domain.User
	repositories []domain.Repository
	pullRequests []domain.PullRequest
}

// NewService builds a Service that skips the given organization names.
func NewService(excludedOrganizations []string) *Service {
	excluded := make(map[string]struct{}, len(excludedOrganizations))
	for _, organization := range excludedOrganizations {
		excluded[organization] = struct{}{}
	}
	return &Service{excludedOrganizations: excluded}
}

// Collect fetches a full Snapshot from source. Any fetch error aborts the
// collection — no partial snapshots, for now.
func (service *Service) Collect(ctx context.Context, source domain.GiteaRepository) (domain.Snapshot, error) {
	organizations, err := source.ListOrgs(ctx)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot := domain.Snapshot{
		Orgs:    make([]domain.Organization, 0, len(organizations)),
		Members: map[string][]domain.User{},
		Repos:   map[string][]domain.Repository{},
		OpenPRs: []domain.PullRequest{},
	}

	for _, organization := range organizations {
		if _, excluded := service.excludedOrganizations[organization.Name]; excluded {
			continue
		}
		data, err := service.collectOrganization(ctx, source, organization)
		if err != nil {
			return domain.Snapshot{}, err
		}
		snapshot.Orgs = append(snapshot.Orgs, organization)
		snapshot.Members[organization.Name] = data.members
		snapshot.Repos[organization.Name] = data.repositories
		snapshot.OpenPRs = append(snapshot.OpenPRs, data.pullRequests...)
	}

	return snapshot, nil
}

func (service *Service) collectOrganization(ctx context.Context, source domain.GiteaRepository, organization domain.Organization) (organizationData, error) {
	var orgData organizationData

	members, err := source.ListMembers(ctx, organization.Name)
	if err != nil {
		return organizationData{}, err
	}
	orgData.members = members

	repositories, err := source.ListRepos(ctx, organization.Name)
	if err != nil {
		return organizationData{}, err
	}
	orgData.repositories = repositories

	pullRequests, err := service.collectPullRequests(ctx, source, organization.Name, repositories)
	if err != nil {
		return organizationData{}, err
	}
	orgData.pullRequests = pullRequests
	return orgData, nil
}

func (service *Service) collectPullRequests(ctx context.Context, source domain.GiteaRepository, organization string, repositories []domain.Repository) ([]domain.PullRequest, error) {
	var pullRequests []domain.PullRequest
	for _, repository := range repositories {
		if !repository.HasPullRequests {
			continue
		}
		if repository.Empty {
			continue
		}
		repositoryPullRequests, err := source.ListOpenPRs(ctx, organization, repository.Name)
		if err != nil {
			return nil, err
		}
		pullRequests = append(pullRequests, repositoryPullRequests...)
	}
	return pullRequests, nil
}
