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
		Name: "orgs_total",
		Help: "Gives the total number of orgs in the gitea instance",
	}, []string{"target"})
	h.collectors["repos"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "repos_total",
		Help: "Gives the total number of repos in the gitea instance per org",
	}, []string{"target", "org"})
}

/*
func (h *Handler) TestDate(configFile string) {
	h.Targets = make(map[string]Server)
	h.Targets["git"] = Server{
		URL:   "https://git.kraetge.net",
		Token: "asdf",
	}
	h.Targets["git-ext"] = Server{
		URL:   "https://git-ext.brillux.de",
		Token: "qwert",
	}
	file, err := os.Create(configFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	d := yaml.NewEncoder(file)
	if err = d.Encode(h.Targets); err != nil {
		panic(err)
	}
}*/

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
		http.Error(w, "Target parameter fehlt", http.StatusBadRequest)
		return
	}
	target, ok := h.Targets[targetParam]
	if !ok {
		errorMsg := fmt.Sprintf("Invalid target: %s\n", targetParam)
		http.Error(w, errorMsg, http.StatusBadRequest)
		h.logger.Error(errorMsg)
		return
	}
	h.registry = prometheus.NewRegistry()

	if h.client == nil {
		h.client = gitea.NewGiteaClient(target.URL, target.Token, h.logger)
	}

	orgs := h.getOrgs(targetParam)
	h.getRepos(targetParam, orgs)

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

func (h *Handler) getOrgRepos(target string, org string) []structs.Repository {
	repos := h.client.GetRepos(org)
	reposTotal := len(repos)
	gauge := h.collectors["repos"].(*prometheus.GaugeVec)
	h.registry.Register(gauge)
	gauge.WithLabelValues(target, org).Set(float64(reposTotal))
	return repos
}

func (h *Handler) getRepos(target string, org []structs.Organization) {
	for _, o := range org {
		h.getOrgRepos(target, o.UserName)
	}
}
