package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/tibeahx/claimer/app/internal/config"
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   2 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout: 2 * time.Second,
}

var httpClient = &http.Client{
	Transport: transport,
	Timeout:   2 * time.Second,
}

type GitlabClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
	projectID  string
	headers    http.Header
}

func NewGitlabClient(cfg *config.Config) GitlabClient {
	c := GitlabClient{
		httpClient: httpClient,
		baseURL:    cfg.Gitlab.BaseURL,
		token:      cfg.Gitlab.Token,
		projectID:  cfg.Gitlab.ProjectID,
	}

	headers := http.Header{
		"Private-Token": []string{fmt.Sprintf("Bearer %s", c.token)},
		"Content-Type":  []string{"application/json"},
	}

	c.headers = headers

	return c
}

func (c GitlabClient) SetHeaders(r *http.Request) {
	for header, val := range c.headers {
		if header == "" {
			continue
		}

		for _, v := range val {
			if v == "" {
				continue
			}

			r.Header.Add(header, v)
		}
	}
}

type Commit struct {
	ID               string         `json:"id"`
	ShortID          string         `json:"short_id"`
	CreatedAt        string         `json:"created_at"`
	ParentIDs        []string       `json:"parent_ids"`
	Title            string         `json:"title"`
	Message          string         `json:"message"`
	AuthorName       string         `json:"author_name"`
	AuthorEmail      string         `json:"author_email"`
	AuthoredDate     string         `json:"authored_date"`
	CommitterName    string         `json:"committer_name"`
	CommitterEmail   string         `json:"committer_email"`
	CommittedDate    string         `json:"committed_date"`
	Trailers         map[string]any `json:"trailers"`
	ExtendedTrailers map[string]any `json:"extended_trailers"`
	WebURL           string         `json:"web_url"`
}

type Branch struct {
	Name               string `json:"name"`
	Merged             bool   `json:"merged"`
	Protected          bool   `json:"protected"`
	Default            bool   `json:"default"`
	DevelopersCanPush  bool   `json:"developers_can_push"`
	DevelopersCanMerge bool   `json:"developers_can_merge"`
	CanPush            bool   `json:"can_push"`
	WebURL             string `json:"web_url"`
	Commit             Commit `json:"commit"`
}

func (c GitlabClient) ListBranches() ([]Branch, error) {
	url := fmt.Sprintf("%s/projects/%s/repository/branches", c.baseURL, c.projectID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	c.SetHeaders(req)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var result []Branch

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
