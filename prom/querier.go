package prom

import (
	"fmt"
	"gitea-exporter/gitea"
	"log/slog"
	"slices"
	"time"

	"code.gitea.io/gitea/modules/structs"
	"github.com/prometheus/client_golang/prometheus"
)

type querier struct {
	Server
	targetName string
	registry   *prometheus.Registry
	collectors map[string]prometheus.Collector
	logger     *slog.Logger
	client     *gitea.Client
}

func newQuerier(server Server, targetName string, logger *slog.Logger) *querier {
	q := querier{
		targetName: targetName,
		Server:     server,
		logger:     logger,
	}
	q.init()
	return &q
}

func (q *querier) init() {
	q.collectors = make(map[string]prometheus.Collector)
	q.collectors["duration"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_probe_duration_seconds",
		Help: "Duration of the probe in seconds",
	}, []string{"target"})
	q.collectors["orgs"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_organizations_total",
		Help: "Gives the total number of orgs in the gitea instance",
	}, []string{"target"})
	q.collectors["org_members"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_organization_members_total",
		Help: "Gives the total number of members in each organization",
	}, []string{"target", "organization"})
	q.collectors["repos"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_repositories_total",
		Help: "Gives the total number of repos in the gitea instance per org",
	}, []string{"target", "organization"})
	q.collectors["pull_requests"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_pull_requests_total",
		Help: "Gives the total number of pull requests in the gitea instance per org",
	}, []string{"target", "organization", "repository"})
	q.collectors["gitea_pull_request_created_at_seconds"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_pull_request_created_at_seconds",
		Help: "Gives the creation time of pull requests in seconds since epoch",
	}, []string{"target", "organization", "repository", "pull_request_id", "poster_username"})
}

func (q *querier) fetchAll() {
	q.registry = prometheus.NewRegistry()
	q.client = gitea.NewGiteaClient(q.Server.URL, q.Server.Token, q.logger)

	startTime := time.Now()

	orgs := q.getOrgs()
	for _, org := range orgs {
		q.getOrgMembers(org.UserName)
		q.getOrgRepos(org.UserName)

		repos := q.getOrgRepos(org.UserName)
		q.getPullRequests(org.UserName, repos)
	}
	duration := time.Since(startTime).Seconds()
	durationGauge := q.collectors["duration"].(*prometheus.GaugeVec)
	q.registry.Register(durationGauge)
	durationGauge.WithLabelValues(q.targetName).Set(duration)
}

func (q *querier) getOrgs() []structs.Organization {
	orgs := q.client.GetOrgs()
	orgs = q.excludeOrgs(orgs)
	orgsTotal := len(orgs)
	gauge := q.collectors["orgs"].(*prometheus.GaugeVec)
	q.registry.Register(gauge)
	gauge.WithLabelValues(q.targetName).Set(float64(orgsTotal))
	return orgs
}

func (q *querier) excludeOrgs(orgs []structs.Organization) []structs.Organization {
	excludeOrgs := q.Server.ExcludeOrgs
	if len(excludeOrgs) == 0 {
		return orgs
	}
	for _, excludeOrg := range excludeOrgs {
		orgs = slices.DeleteFunc(orgs, func(org structs.Organization) bool {
			return org.UserName == excludeOrg
		})
	}
	return orgs
}

func (q *querier) getOrgMembers(org string) []structs.User {
	members := q.client.GetOrgMembers(org)
	membersTotal := len(members)
	gauge := q.collectors["org_members"].(*prometheus.GaugeVec)
	q.registry.Register(gauge)
	gauge.WithLabelValues(q.targetName, org).Set(float64(membersTotal))
	return members
}

func (q *querier) getOrgRepos(org string) []structs.Repository {
	repos := q.client.GetRepos(org)
	reposTotal := len(repos)
	gauge := q.collectors["repos"].(*prometheus.GaugeVec)
	q.registry.Register(gauge)
	gauge.WithLabelValues(q.targetName, org).Set(float64(reposTotal))
	return repos
}

func (q *querier) getPullRequests(org string, repos []structs.Repository) {
	for _, r := range repos {
		prs := q.getRepositoryPullRequests(q.targetName, org, r.Name)
		q.logger.Info(fmt.Sprintf("Org: %s, Repo: %s, PRs: %d", org, r.Name, len(prs)))
	}
}

func (q *querier) getRepositoryPullRequests(target string, org string, repo string) []structs.PullRequest {
	pullRequests := q.client.GetPullRequests(org, repo)
	pullRequestsTotal := len(pullRequests)
	gauge := q.collectors["pull_requests"].(*prometheus.GaugeVec)
	q.registry.Register(gauge)
	gauge.WithLabelValues(target, org, repo).Set(float64(pullRequestsTotal))

	if pullRequestsTotal == 0 {
		return pullRequests
	}
	createdAtGauge := q.collectors["gitea_pull_request_created_at_seconds"].(*prometheus.GaugeVec)
	q.registry.Register(createdAtGauge)
	for _, pr := range pullRequests {
		pullRequestID := fmt.Sprintf("%d", pr.ID)
		posterUsername := pr.Poster.UserName
		unixTime := float64(pr.Created.Unix())
		createdAtGauge.WithLabelValues(target, org, repo, pullRequestID, posterUsername).Set(unixTime)
	}
	return pullRequests
}
