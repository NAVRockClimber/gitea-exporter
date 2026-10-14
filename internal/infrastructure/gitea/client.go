package gitea

import (
	"context"
	"encoding/json"
	"fmt"
	"gitea-exporter/internal/domain"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gitea.dev/modules/structs"
)

const basePath = "/api/v1"

// Client reads from the Gitea REST API and maps the results into
// domain types. It is the HTTP-backed implementation of
// domain.GiteaRepository.
type Client struct {
	serverURL  string
	token      string
	httpClient *http.Client
	endpoints  map[string]string
}

// New builds a Client for one Gitea instance.
func New(serverURL, token string) *Client {
	c := &Client{
		serverURL:  strings.TrimRight(serverURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	c.init()
	return c
}

func (c *Client) init() {
	c.endpoints = map[string]string{
		"orgs":          "/orgs",
		"org_members":   "/orgs/%s/members",
		"repos":         "/orgs/%s/repos",
		"pull_requests": "/repos/%s/%s/pulls?state=open",
	}
}

func (c *Client) createRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	req.Header.Set("User-Agent", "BX-DevOps-Metrics/1.0")
	req.Header.Set("accept", "application/json")

	return req, nil
}

// get performs a GET against basePath+path, decodes the JSON body
// and returns the data. All errors are returned with context via %w.
func (c *Client) get[T any](ctx context.Context, url string) ([]T, error) {
	slog.Debug("calling Gitea API", "url", url)

	req, err := c.createRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message, err := c.getErrorResponse(resp)
		if err != nil {
			return nil, err
		}
		slog.Info("Error", "url", url, "response", message)
		return nil, fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	var out []T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode %s: %w", url, err)
	}
	return out, nil
}

func (c *Client) getErrorResponse(resp *http.Response) (string, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("reading error response", "error", err)
		return "", err
	}
	return string(bodyBytes), nil
}

func (c *Client) getURL(endpoint string, pathElements ...any) string {
	endpointString := c.endpoints[endpoint]
	path := endpointString
	if strings.Contains(endpointString, "%s") {
		path = fmt.Sprintf(endpointString, pathElements...)
	}
	return fmt.Sprintf("%s%s%s", c.serverURL, basePath, path)
}

func (c *Client) ListOrgs(ctx context.Context) ([]domain.Organization, error) {
	url := c.getURL("orgs")
	orgs, err := c.get[structs.Organization](ctx, url)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Organization, 0, len(orgs))
	for _, o := range orgs {
		out = append(out, domain.Organization{Name: o.UserName})
	}
	return out, nil
}
func (c *Client) ListMembers(ctx context.Context, org string) ([]domain.User, error) {
	url := c.getURL("org_members", org)
	members, err := c.get[structs.User](ctx, url)
	if err != nil {
		return nil, err
	}
	out := make([]domain.User, 0, len(members))
	for _, m := range members {
		out = append(out, domain.User{Name: m.UserName})
	}
	return out, nil
}
func (c *Client) ListRepos(ctx context.Context, org string) ([]domain.Repository, error) {
	url := c.getURL("repos", org)
	repos, err := c.get[structs.Repository](ctx, url)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Repository, 0, len(repos))
	for _, r := range repos {
		out = append(out, domain.Repository{
			Name:            r.Name,
			HasPullRequests: r.HasPullRequests,
			Empty:           r.Empty,
		})
	}
	return out, nil

}
func (c *Client) ListOpenPRs(ctx context.Context, org, repo string) ([]domain.PullRequest, error) {
	url := c.getURL("pull_requests", org, repo)
	prs, err := c.get[structs.PullRequest](ctx, url)
	if err != nil {
		return nil, err
	}
	out := make([]domain.PullRequest, 0, len(prs))
	for _, p := range prs {
		out = append(out, domain.PullRequest{
			ID:        p.ID,
			Org:       org,
			Repo:      repo,
			Author:    p.Poster.UserName,
			CreatedAt: p.Created,
		})
	}
	return out, nil

}
