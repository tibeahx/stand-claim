package gitlabwrapper

import (
	"errors"
	"fmt"

	"github.com/tibeahx/claimer/app/internal/config"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	errNoJobs     = errors.New("no jobs found")
	errNoBranches = errors.New("no branches found")
	errNoProjects = errors.New("no projects found")
)

type wrapperOptions func(*GitlabClientWrapper)

type GitlabClientWrapper struct {
	client    *gitlab.Client
	token     string
	projectID int
	groupID   int
}

func WithGroupID(groupID int) wrapperOptions {
	return func(gcw *GitlabClientWrapper) {
		gcw.groupID = groupID
	}
}

func WithProjectID(projectID int) wrapperOptions {
	return func(gcw *GitlabClientWrapper) {
		gcw.projectID = projectID
	}
}

func NewGitlabClientWrapper(cfg *config.Config, opts ...wrapperOptions) (*GitlabClientWrapper, error) {
	c, err := gitlab.NewClient(cfg.Gitlab.Token, gitlab.WithBaseURL(cfg.Gitlab.BaseURL))

	gcw := &GitlabClientWrapper{
		token:  cfg.Gitlab.Token,
		client: c,
	}

	for _, opt := range opts {
		opt(gcw)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to init gitlab client due to :%w", err)
	}

	return gcw, nil
}

func (c *GitlabClientWrapper) ListProjectJobs(opts *gitlab.ListJobsOptions) ([]*gitlab.Job, error) {
	jobs, _, err := c.client.Jobs.ListProjectJobs(c.projectID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list project jobs due to: %w", err)
	}

	if jobs == nil {
		return nil, errNoJobs
	}

	return jobs, nil
}

type BranchInfo struct {
	Name   string
	Commit *gitlab.Commit
}

func (c *GitlabClientWrapper) ListGroupProjectsWithBranchInfo(opts *gitlab.ListGroupProjectsOptions) (map[string][]BranchInfo, error) {
	projects, _, err := c.client.Groups.ListGroupProjects(c.groupID, opts)
	if err != nil {
		return nil, err
	}

	if projects == nil {
		return nil, errNoProjects
	}

	biOpts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Sort:    "desc",
			OrderBy: "name",
		},
	}

	result := make(map[string][]BranchInfo)

	for _, project := range projects {
		if project.Archived || project.Visibility == gitlab.PrivateVisibility {
			continue
		}

		bi, err := c.ListRepoBranches(biOpts)
		if err != nil {
			return nil, err
		}

		result[project.Name] = bi
	}

	return result, nil
}

func (c *GitlabClientWrapper) ListRepoBranches(opts *gitlab.ListBranchesOptions) ([]BranchInfo, error) {
	branches, _, err := c.client.Branches.ListBranches(c.projectID, opts)
	if err != nil {
		return nil, err
	}

	if branches == nil {
		return nil, errNoBranches
	}

	result := make([]BranchInfo, len(branches))

	for i, branch := range branches {
		if branch.Name == "" || branch.Commit == nil {
			continue
		}
		result[i] = BranchInfo{
			Name:   branch.Name,
			Commit: branch.Commit,
		}
	}

	return result, nil
}

// нужно из группы достать все проекты, затем у этих проектов достать все бранчи, чтобы получилось в итоге  проект: имяБранча: коммит
// func (c *GitlabClientWrapper) ListGroupBranches(opts *gitlab.ListBranchesOptions)

// главное - дернуть ручку бота, которая покажет кто из пользователей на каком стенде какую ветку держит
// для этого надо пойти в гитлаб, дернуть получение всех веток
// далее для каждой ветки распарсить овнера данной ветки
// пока что так
