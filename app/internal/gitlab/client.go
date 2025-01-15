package gitlabwrapper

import (
	"errors"
	"fmt"
	"slices"

	"github.com/tibeahx/claimer/app/internal/config"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	errInvalidBranchState   = errors.New("invalid branch state")
	errMrIsntMergedInTarget = errors.New("mr isn't merged in target")
	errMrIsntApproved       = errors.New("mr isn't approved")
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

type FeatureState string

const (
	stateInDevelopment FeatureState = "in development"
	stateInProduction  FeatureState = "in production"
	stateMerged        FeatureState = "in %s"
)

func (c *GitlabClientWrapper) GetFeatureState(targetBranches []string, exceptions ...string) (map[string]FeatureState, error) {
	branches, err := c.getBranches(exceptions...)
	if err != nil {
		return nil, err
	}

	states := make(map[string]FeatureState, len(branches))

	for _, branch := range branches {
		states[branch.Name] = stateInDevelopment

		for _, target := range targetBranches {
			merged, err := c.isBranchMergedIntoTarget(branch.Name, target)
			if err != nil {
				if errors.Is(err, errMrIsntApproved) || errors.Is(err, errMrIsntMergedInTarget) {
					continue
				}
				return nil, err
			}
			if merged {
				states[branch.Name] = FeatureState(fmt.Sprintf(string(stateMerged), target))
				break
			}
		}

		allMerged := true

		for _, target := range targetBranches {
			merged, err := c.isBranchMergedIntoTarget(branch.Name, target)
			if err != nil || !merged {
				allMerged = false
				break
			}
		}

		if allMerged {
			states[branch.Name] = stateInProduction
		}
	}

	return states, nil
}

func (c *GitlabClientWrapper) getBranches(exceptions ...string) ([]*gitlab.Branch, error) {
	branches, _, err := c.client.Branches.ListBranches(c.projectID, &gitlab.ListBranchesOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get branches due to :%w", err)
	}

	filteredBranches := make([]*gitlab.Branch, len(branches))

	for i, branch := range branches {
		if slices.Contains(exceptions, branch.Name) {
			continue
		}
		filteredBranches[i] = &gitlab.Branch{
			Name:               branch.Name,
			Commit:             branch.Commit,
			Merged:             branch.Merged,
			Protected:          branch.Protected,
			Default:            branch.Default,
			DevelopersCanPush:  branch.DevelopersCanPush,
			DevelopersCanMerge: branch.DevelopersCanMerge,
			CanPush:            branch.CanPush,
			WebURL:             branch.WebURL,
		}
	}

	return filteredBranches, nil
}

func (c *GitlabClientWrapper) getBranch(branchName string) (*gitlab.Branch, error) {
	branch, _, err := c.client.Branches.GetBranch(c.projectID, branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch due to :%w", err)
	}

	return branch, nil
}

func (c *GitlabClientWrapper) isBranchMergedIntoTarget(sourceBranch string, targetBranch string) (bool, error) {
	branch, err := c.getBranch(sourceBranch)
	if err != nil {
		return false, err
	}

	if !c.isBranchMerged(branch, branch.Name) {
		return false, errInvalidBranchState
	}

	mrs, err := c.getMRs(branch.Name)
	if err != nil {
		return false, err
	}

	if !c.isMrMergedInTarget(mrs, targetBranch) {
		return false, errMrIsntMergedInTarget
	}

	for _, mr := range mrs {
		if c.isMrApproved(mr) {
			return true, nil
		}
	}

	return false, errMrIsntApproved
}

func (c *GitlabClientWrapper) getMRs(sourceBranch string) ([]*gitlab.MergeRequest, error) {
	mrs, _, err := c.client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		SourceBranch: &sourceBranch,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get mrs due to :%w", err)
	}

	return mrs, nil
}

func (c *GitlabClientWrapper) isMrMergedInTarget(mrs []*gitlab.MergeRequest, targetBranch string) bool {
	for _, mr := range mrs {
		if mr.TargetBranch == targetBranch && mr.State == "merged" {
			return true
		}
	}
	return false
}

func (c *GitlabClientWrapper) isMrApproved(mr *gitlab.MergeRequest) bool {
	return mr.State == "approved"
}

func (c *GitlabClientWrapper) isBranchMerged(b *gitlab.Branch, sourceBranch string) bool {
	return b.Merged && b.Name == sourceBranch
}
