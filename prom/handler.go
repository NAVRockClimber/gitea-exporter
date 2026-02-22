package prom

import (
	"fmt"
	"gitea-exporter/gitea"
	"log/slog"
	"net/http"
	"os"

	"code.gitea.io/gitea/modules/structs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.yaml.in/yaml/v2"
)

// Handler handle prometheus queries
type Handler struct {
	Targets    Targets `yaml:"targets"`
	logger     *slog.Logger
	registry   *prometheus.Registry
	collectors map[string]prometheus.Collector
	client     *gitea.Client
}

// Targets is map of probing targets
type Targets map[string]Server

// Server struct describe a gitea server
type Server struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type collectorDescription struct {
	name string
	help string
}

// NewHandler creeates a new prometheus handler
func NewHandler(configFile string, logger *slog.Logger) *Handler {
	t := readConfig(configFile)
	h := Handler{
		Targets: t,
		logger:  logger,
	}
	h.init()
	return &h
}

func (h *Handler) init() {
	h.collectors = make(map[string]prometheus.Collector)
	h.collectors["orgs"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_organizations_total",
		Help: "Gives the total number of orgs in the gitea instance",
	}, []string{"target"})
	h.collectors["org_members"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_organization_members_total",
		Help: "Gives the total number of members in each organization",
	}, []string{"target", "organization"})
	h.collectors["repos"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_repositories_total",
		Help: "Gives the total number of repos in the gitea instance per org",
	}, []string{"target", "organization"})
	h.collectors["pull_requests"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_pull_requests_total",
		Help: "Gives the total number of pull requests in the gitea instance per org",
	}, []string{"target", "organization", "repository"})
	h.collectors["gitea_pull_request_created_at_seconds"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitea_pull_request_created_at_seconds",
		Help: "Gives the creation time of pull requests in seconds since epoch",
	}, []string{"target", "organization", "repository", "pull_request_id", "poster_username"})
}

func readConfig(configFile string) Targets {
	file, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var targets Targets
	d := yaml.NewDecoder(file)
	if err = d.Decode(&targets); err != nil {
		panic(err)
	}
	return targets
}

// ProbeHandler handles the incoming prometheus queries
func (h *Handler) ProbeHandler(w http.ResponseWriter, r *http.Request) {
	targetParam := r.URL.Query().Get("target")
	if targetParam == "" {
		http.Error(w, "Target parameter missing", http.StatusBadRequest)
		return
	}
	target, ok := h.Targets[targetParam]
	if !ok {
		errorMsg := fmt.Sprintf("Invalid target: %s", targetParam)
		http.Error(w, errorMsg, http.StatusBadRequest)
		h.logger.Error(errorMsg)
		return
	}
	h.registry = prometheus.NewRegistry()

	if h.client == nil {
		h.client = gitea.NewGiteaClient(target.URL, target.Token, h.logger)
	}

	orgs := h.getOrgs(targetParam)
	for _, org := range orgs {
		h.getOrgMembers(targetParam, org.UserName)
		h.getOrgRepos(targetParam, org.UserName)

		repos := h.getOrgRepos(targetParam, org.UserName)
		h.getPullRequests(targetParam, org.UserName, repos)
	}

	httpHandler := promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{})
	httpHandler.ServeHTTP(w, r)
}

func (h *Handler) getOrgs(target string) []structs.Organization {
	orgs := h.client.GetOrgs()
	orgsTotal := len(orgs)
	gauge := h.collectors["orgs"].(*prometheus.GaugeVec)
	h.registry.Register(gauge)
	gauge.WithLabelValues(target).Set(float64(orgsTotal))
	return orgs
}

func (h *Handler) getOrgMembers(target string, org string) []structs.User {
	members := h.client.GetOrgMembers(org)
	membersTotal := len(members)
	gauge := h.collectors["org_members"].(*prometheus.GaugeVec)
	h.registry.Register(gauge)
	gauge.WithLabelValues(target, org).Set(float64(membersTotal))
	return members
}

func (h *Handler) getOrgRepos(target string, org string) []structs.Repository {
	repos := h.client.GetRepos(org)
	reposTotal := len(repos)
	gauge := h.collectors["repos"].(*prometheus.GaugeVec)
	h.registry.Register(gauge)
	gauge.WithLabelValues(target, org).Set(float64(reposTotal))
	return repos
}

func (h *Handler) getPullRequests(target string, org string, repos []structs.Repository) {
	for _, r := range repos {
		prs := h.getRepositoryPullRequests(target, org, r.Name)
		h.logger.Info(fmt.Sprintf("Org: %s, Repo: %s, PRs: %d", org, r.Name, len(prs)))
	}
}

func (h *Handler) getRepositoryPullRequests(target string, org string, repo string) []structs.PullRequest {
	pullRequests := h.client.GetPullRequests(org, repo)
	pullRequestsTotal := len(pullRequests)
	gauge := h.collectors["pull_requests"].(*prometheus.GaugeVec)
	h.registry.Register(gauge)
	gauge.WithLabelValues(target, org, repo).Set(float64(pullRequestsTotal))

	if pullRequestsTotal == 0 {
		return pullRequests
	}
	createdAtGauge := h.collectors["gitea_pull_request_created_at_seconds"].(*prometheus.GaugeVec)
	h.registry.Register(createdAtGauge)
	for _, pr := range pullRequests {
		pullRequestID := fmt.Sprintf("%d", pr.ID)
		posterUsername := pr.Poster.UserName
		unixTime := float64(pr.Created.Unix())
		createdAtGauge.WithLabelValues(target, org, repo, pullRequestID, posterUsername).Set(unixTime)
	}
	return pullRequests
}
