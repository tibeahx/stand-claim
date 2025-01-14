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
	if err != nil {
		return nil, fmt.Errorf("failed to init gitlab client due to :%w", err)
	}

	gcw := &GitlabClientWrapper{
		token:  cfg.Gitlab.Token,
		client: c,
	}

	for _, opt := range opts {
		opt(gcw)
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
	Name    string
	Commits []*gitlab.Commit
}

func (c *GitlabClientWrapper) ListGroupProjectsWithBranchInfo(
	perPage int,
	orderBy string,
	sort string,
) (map[string]BranchInfo, error) {
	projects, _, err := c.client.Groups.ListGroupProjects(c.groupID, &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: perPage,
			OrderBy: orderBy,
			Sort:    sort,
		},
	})
	if err != nil {
		return nil, err
	}

	if projects == nil {
		return nil, errNoProjects
	}

	result := make(map[string]BranchInfo)

	for _, project := range projects {
		if project.Archived || project.Visibility == gitlab.PrivateVisibility {
			continue
		}

		bi, err := c.ListRepoBranches(50, "name", "desc")
		if err != nil {
			return nil, err
		}

		for _, b := range bi {
			branchInfo := BranchInfo{
				Name: b.Name,
			}
			branchInfo.Commits = append(branchInfo.Commits, b.Commits...)
			result[project.Name] = branchInfo
		}
	}

	return result, nil
}

func (c *GitlabClientWrapper) ListRepoBranches(
	perPage int,
	orderBy string,
	sort string,
) ([]BranchInfo, error) {
	branches, _, err := c.client.Branches.ListBranches(c.projectID, &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: perPage,
			OrderBy: orderBy,
			Sort:    sort,
		},
	})
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
			Name: branch.Name,
		}
		result[i].Commits = append(result[i].Commits, branch.Commit)
	}

	return result, nil
}

func (c *GitlabClientWrapper) CommitsFromBranch(branchName string) ([]*gitlab.Commit, error) {
	commits, _, err := c.client.Commits.ListCommits(c.projectID, &gitlab.ListCommitsOptions{
		RefName: &branchName,
	})
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (c *GitlabClientWrapper) PipelinesFromProject(branchName string, username string) ([]*gitlab.PipelineInfo, error) {
	pipelines, _, err := c.client.Pipelines.ListProjectPipelines(c.projectID, &gitlab.ListProjectPipelinesOptions{
		Ref:      &branchName,
		Username: &username,
	})
	if err != nil {
		return nil, err
	}

	return pipelines, nil
}

// стенд - это ветка, например дев стенд это и есть дев ветка
// значит нужно проверить, какие есть ветки в проекте
// затем зайти в пайплайн дева и стеджа
// затем найти там джобы с названиями веток с фичами
// если джобы завершены, ветки влиты в дев или стейдж
