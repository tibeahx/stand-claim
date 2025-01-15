package gitlabwrapper

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/tibeahx/claimer/app/internal/config"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	errMrIsntMergedInTarget = errors.New("mr isn't merged in target")
	errMrIsntApproved       = errors.New("mr isn't approved")
)

type wrapperOptions func(*GitlabClientWrapper)

type GitlabClientWrapper struct {
	client    *gitlab.Client
	mu        sync.RWMutex
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

func (c *GitlabClientWrapper) GetFeaturesWithStateAsync(envBranches []string) (map[string]FeatureState, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	branches, err := c.getBranches(envBranches...)
	if err != nil {
		return nil, err
	}

	states := make(map[string]FeatureState, len(branches))

	var wg sync.WaitGroup
	for _, branch := range branches {
		wg.Add(1)

		go func(branch *gitlab.Branch) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				state := c.determineBranchStateAsync(branch, envBranches)
				c.mu.Lock()
				states[branch.Name] = state
				c.mu.Unlock()
			}
		}(branch)
	}

	wg.Wait()

	return states, nil
}

func (c *GitlabClientWrapper) determineBranchStateAsync(branch *gitlab.Branch, envBranches []string) FeatureState {
	mergedTargets := 0
	var lastMergedTarget string

	var wg sync.WaitGroup
	for _, target := range envBranches {
		wg.Add(1)

		go func(target string) {
			defer wg.Done()

			merged, err := c.isBranchMergedIntoTarget(branch.Name, target)
			if err != nil {
				if errors.Is(err, errMrIsntApproved) || errors.Is(err, errMrIsntMergedInTarget) {
					return
				}
				return
			}

			if merged {
				c.mu.Lock()
				mergedTargets++
				lastMergedTarget = target
				c.mu.Unlock()
			}
		}(target)
	}

	wg.Wait()

	switch {
	case mergedTargets == 0:
		return stateInDevelopment
	case mergedTargets == len(envBranches):
		return stateInProduction
	default:
		return FeatureState(fmt.Sprintf(string(stateMerged), lastMergedTarget))
	}
}

func (c *GitlabClientWrapper) getBranches(exceptions ...string) ([]*gitlab.Branch, error) {
	branches, _, err := c.client.Branches.ListBranches(c.projectID, &gitlab.ListBranchesOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get branches due to :%w", err)
	}

	filteredBranches := make([]*gitlab.Branch, 0, len(branches))

	for _, branch := range branches {
		if slices.Contains(exceptions, branch.Name) {
			continue
		}
		filteredBranches = append(filteredBranches, &gitlab.Branch{
			Name:               branch.Name,
			Commit:             branch.Commit,
			Merged:             branch.Merged,
			Protected:          branch.Protected,
			Default:            branch.Default,
			DevelopersCanPush:  branch.DevelopersCanPush,
			DevelopersCanMerge: branch.DevelopersCanMerge,
			CanPush:            branch.CanPush,
			WebURL:             branch.WebURL,
		})
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
