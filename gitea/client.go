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
	c.endpoints["repos"] = "/orgs/%s/repos"
}

func (c *Client) getURL(endpoint string) string {
	return fmt.Sprintf("%s%s%s", c.serverURL, c.endpoints["basePath"], endpoint)
}

func (c *Client) GetOrgs() []structs.Organization {
	endpoint := c.endpoints["orgs"]
	url := c.getURL(endpoint)
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
	var orgs []structs.Organization
	if err = json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		errorMsg := "Error decoding response"
		c.logger.Error(errorMsg, "Error", err)
		return nil
	}
	return orgs
}

func (c *Client) GetRepos(org string) []structs.Repository {
	endpoint := fmt.Sprintf(c.endpoints["repos"], org)
	url := c.getURL(endpoint)
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
	var repos []structs.Repository
	if err = json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		errorMsg := "Error decoding response"
		c.logger.Error(errorMsg, "Error", err)
		return nil
	}
	return repos
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
