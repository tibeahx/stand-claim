package gitlab

import (
	"errors"
	"fmt"

	"github.com/tibeahx/claimer/app/internal/config"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitlabClientWrapper struct {
	client    *gitlab.Client
	token     string
	projectID int
}

func NewGitlabClientWrapper(cfg *config.Config) (*GitlabClientWrapper, error) {
	c, err := gitlab.NewClient(cfg.Gitlab.Token, gitlab.WithBaseURL(cfg.Gitlab.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to init gitlab client due to :%w", err)
	}

	gw := &GitlabClientWrapper{
		client:    c,
		token:     cfg.Gitlab.Token,
		projectID: cfg.Gitlab.ProjectID,
	}

	return gw, nil
}

func (c *GitlabClientWrapper) ListProjectJobs() ([]*gitlab.Job, error) {
	opts := &gitlab.ListJobsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 5,
			Sort:    "desc",
			OrderBy: "id",
		},
	}

	jobs, _, err := c.client.Jobs.ListProjectJobs(
		c.projectID,
		opts,
		gitlab.WithToken(gitlab.AuthType(3), c.token),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list project jobs due to: %w", err)
	}

	if jobs == nil {
		return nil, errors.New("no jobs found")
	}

	return jobs, nil
}

func (c *GitlabClientWrapper) ListRepoBranches() ([]*gitlab.Branch, error) {
	opts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Sort:    "desc",
			OrderBy: "name",
		},
	}

	branches, _, err := c.client.Branches.ListBranches(
		c.projectID,
		opts,
		gitlab.WithToken(gitlab.AuthType(3), c.token),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list repo branches due to: %w", err)
	}

	if branches == nil {
		return nil, errors.New("no branches found")
	}

	return branches, nil
}
