package domain

import "time"

// Organization is one organization on the Gitea instance.
type Organization struct {
	Name string
}

// Repository is one repository inside an organization.
type Repository struct {
	Name            string
	HasPullRequests bool
	Empty           bool
}

// User is one member of an organization.
type User struct {
	Name string
}

// PullRequest is one OPEN pull request. We only ever collect open PRs,
// so a closed PR is represented by its absence from a Snapshot.
type PullRequest struct {
	ID        int64
	Org       string
	Repo      string
	Author    string
	CreatedAt *time.Time
}

// Snapshot is the complete result of one collection pass.
type Snapshot struct {
	Orgs    []Organization
	Members map[string][]User       // org name -> members
	Repos   map[string][]Repository // org name -> repos
	OpenPRs []PullRequest
}
