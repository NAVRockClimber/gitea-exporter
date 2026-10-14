package domain

import "context"

// GiteaRepository is everything the app needs to READ from Gitea.
// It returns lists (not counts) so the caller can decide what to count,
// and so results can be cached and reused without re-fetching.
type GiteaRepository interface {
	ListOrgs(ctx context.Context) ([]Organization, error)
	ListMembers(ctx context.Context, org string) ([]User, error)
	ListRepos(ctx context.Context, org string) ([]Repository, error)
	ListOpenPRs(ctx context.Context, org, repo string) ([]PullRequest, error)
}
