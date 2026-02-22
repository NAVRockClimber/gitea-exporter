// Package gitea provides a client for the gitea api
package gitea

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"code.gitea.io/gitea/modules/structs"
)

// Client provides methods for querying the gitea api
type Client struct {
	serverURL string
	apiToken  string
	endpoints map[string]string
	logger    *slog.Logger
}

// NewGiteaClient creates a new gitea client
func NewGiteaClient(serverURL string, apiToken string, logger *slog.Logger) *Client {
	c := Client{
		serverURL: serverURL,
		apiToken:  apiToken,
		logger:    logger,
	}
	c.init()
	return &c
}

func (c *Client) init() {
	c.endpoints = make(map[string]string)
	c.endpoints["basePath"] = "/api/v1"
	c.endpoints["orgs"] = "/orgs"
	c.endpoints["org_members"] = "/orgs/%s/members"
	c.endpoints["repos"] = "/orgs/%s/repos"
	c.endpoints["pull_requests"] = "/repos/%s/%s/pulls?state=open"
}

func (c *Client) getURL(endpoint string) string {
	return fmt.Sprintf("%s%s%s", c.serverURL, c.endpoints["basePath"], endpoint)
}

// GetOrgs returns a list of orgs in the gitea instance
func (c *Client) GetOrgs() []structs.Organization {
	endpoint := c.endpoints["orgs"]
	url := c.getURL(endpoint)
	return fetchAPI[structs.Organization](c, url)
}

// GetOrgMembers returns a list of members for the given org
func (c *Client) GetOrgMembers(org string) []structs.User {
	endpoint := fmt.Sprintf(c.endpoints["org_members"], org)
	url := c.getURL(endpoint)
	return fetchAPI[structs.User](c, url)
}

// GetRepos returns a list of repos for the given org
func (c *Client) GetRepos(org string) []structs.Repository {
	endpoint := fmt.Sprintf(c.endpoints["repos"], org)
	url := c.getURL(endpoint)
	return fetchAPI[structs.Repository](c, url)
}

// GetPullRequests returns a list of pull requests for the given repo
func (c *Client) GetPullRequests(owner string, repo string) []structs.PullRequest {
	endpoint := fmt.Sprintf(c.endpoints["pull_requests"], owner, repo)
	url := c.getURL(endpoint)
	return fetchAPI[structs.PullRequest](c, url)
}

func fetchAPI[T any](c *Client, url string) []T {
	logMsg := fmt.Sprintf("Calling: %s", url)
	c.logger.Info(logMsg)
	req := c.createRequest(url)
	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		errorMsg := "Error during request"
		c.logger.Error(errorMsg, "Error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("Received: %s", resp.Status)
		c.logger.Error(errorMsg, "URL", url, "StatusCode", resp.StatusCode, "Status", resp.Status)
		return nil
	}
	var result []T
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		errorMsg := "Error decoding response"
		c.logger.Error(errorMsg, "Error", err)
		return nil
	}
	return result
}

func (c *Client) createRequest(url string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		errorMsg := "Error creating request"
		c.logger.Error(errorMsg, "Error", err)
		return nil
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.apiToken))
	req.Header.Set("User-Agent", "BX-DevOps-Metrics/1.0")
	req.Header.Set("accept", "application/json")

	return req
}
