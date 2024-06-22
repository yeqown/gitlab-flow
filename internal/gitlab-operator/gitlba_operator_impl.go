package gitlabop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	gogitlab "github.com/xanzy/go-gitlab"
	"github.com/yeqown/log"
)

// gitlabOperator implement IGitlabOperator to operate remote gitlab repository.
type gitlabOperator struct {
	gitlab *gogitlab.Client
}

type Config struct {
	AppID        string
	AppSecret    string
	AccessToken  string
	RefreshToken string
	Host         string
	ApiURL       string
}

// NewGitlabOperator generate IGitlabOperator.
func NewGitlabOperator(accessToken, apiURL string) IGitlabOperator {
	log.
		WithFields(log.Fields{
			"accessToken": accessToken,
			"apiURL":      apiURL,
		}).
		Debug("NewGitlabOperator get new access token")

	gitlab, err := gogitlab.NewOAuthClient(accessToken, gogitlab.WithBaseURL(apiURL))
	if err != nil {
		log.
			WithFields(log.Fields{
				"AccessToken": accessToken,
				"apiURL":      apiURL,
			}).
			Errorf("NewGitlabOperator could not initialize OAuth2 client: %v", err)
		// could not go ahead if we could not initialize gitlab client.
		panic(err)
	}

	return &gitlabOperator{
		gitlab: gitlab,
	}
}

func (g gitlabOperator) CreateBranch(ctx context.Context, req *CreateBranchRequest) (*CreateBranchResult, error) {
	ref := req.SrcBranch
	opt := &gogitlab.CreateBranchOptions{
		Branch: &req.TargetBranch,
		Ref:    &ref,
	}
	branch, _, err := g.gitlab.Branches.CreateBranch(req.ProjectID, opt)
	if err != nil {
		// if create failed then query from remote, if got then return
		var err2 error
		branch, _, err2 = g.gitlab.Branches.GetBranch(req.ProjectID, req.TargetBranch)
		if err2 != nil {
			return nil, fmt.Errorf("create branch failed: %v, query failed: %v", err, err2)
		}
	}

	return &CreateBranchResult{
		Name:   branch.Name,
		WebURL: branch.WebURL,
	}, nil
}

func (g gitlabOperator) CreateMilestone(ctx context.Context, req *CreateMilestoneRequest) (*CreateMilestoneResult, error) {
	opt := &gogitlab.CreateMilestoneOptions{
		Title:       &req.Title,
		Description: &req.Desc,
		StartDate:   nil,
		DueDate:     nil,
	}
	milestone, _, err := g.gitlab.Milestones.CreateMilestone(req.ProjectID, opt)
	if err != nil || milestone == nil {
		// if create failed then query from remote, if got then return
		opt := gogitlab.ListMilestonesOptions{
			ListOptions: gogitlab.ListOptions{
				Page:    1,
				PerPage: 20,
			},
			Title:  &req.Title,
			Search: &req.Title,
		}
		milestones, _, err2 := g.gitlab.Milestones.ListMilestones(req.ProjectID, &opt)
		if err2 != nil {
			return nil, fmt.Errorf("create milestone failed: %v, query failed: %v", err, err2)
		}

		matched := false
		for idx, v := range milestones {
			if strings.Compare(req.Title, v.Title) == 0 &&
				strings.Compare(req.Desc, v.Description) == 0 {
				//	matched
				milestone = milestones[idx]
				matched = true
			}
		}

		if !matched || milestone == nil {
			return nil, fmt.Errorf("[matched: %v] create milestone failed: %v", matched, err)
		}
	}

	return &CreateMilestoneResult{
		ID:     milestone.ID,
		WebURL: milestone.WebURL,
	}, nil
}

func (g gitlabOperator) GetMilestone(ctx context.Context, req *GetMilestoneRequest) (*GetMilestoneResult, error) {
	milestone, _, err := g.gitlab.Milestones.GetMilestone(req.ProjectID, req.MilestoneID)
	if err != nil {
		return nil, errors.Wrap(err, "get milestone failed")
	}

	return &GetMilestoneResult{
		ID:          milestone.ID,
		Title:       milestone.Title,
		Description: milestone.Description,
		WebURL:      milestone.WebURL,
	}, nil
}

func (g gitlabOperator) GetMilestoneMergeRequests(
	ctx context.Context, req *GetMilestoneMergeRequestsRequest) (*GetMilestoneMergeRequestsResult, error) {
	opt := gogitlab.GetMilestoneMergeRequestsOptions{}
	mrs, _, err := g.gitlab.Milestones.GetMilestoneMergeRequests(req.ProjectID, req.MilestoneID, &opt)
	if err != nil {
		return nil, errors.Wrap(err, "get milestone merge requests failed")
	}

	result := new(GetMilestoneMergeRequestsResult)
	result.Data = make([]MergeRequestShort, 0, len(mrs))

	for _, v := range mrs {
		result.Data = append(result.Data, MergeRequestShort{
			ID:           v.ID,
			Title:        v.Title,
			Description:  v.Description,
			WebURL:       v.WebURL,
			SourceBranch: v.SourceBranch,
			TargetBranch: v.TargetBranch,
		})
	}

	return result, nil
}

func (g gitlabOperator) GetMilestoneIssues(
	ctx context.Context, req *GetMilestoneIssuesRequest) (*GetMilestoneIssuesResult, error) {
	opt := gogitlab.GetMilestoneIssuesOptions{}
	issues, _, err := g.gitlab.Milestones.GetMilestoneIssues(req.ProjectID, req.MilestoneID, &opt)
	if err != nil {
		return nil, errors.Wrap(err, "get milestone issues failed")
	}

	result := new(GetMilestoneIssuesResult)
	result.Data = make([]IssueShort, 0, len(issues))

	for _, v := range issues {
		result.Data = append(result.Data, IssueShort{
			ID:          v.ID,
			IID:         v.IID,
			Title:       v.Title,
			Description: v.Description,
			WebURL:      v.WebURL,
			ProjectID:   v.ProjectID,
			MilestoneID: v.Milestone.ID,
		})
	}

	return result, nil
}

func (g gitlabOperator) CreateIssue(ctx context.Context, req *CreateIssueRequest) (*CreateIssueResult, error) {
	now := time.Now()
	opt3 := &gogitlab.CreateIssueOptions{
		Title:       &req.Title,
		Description: &req.Desc,
		MilestoneID: &req.MilestoneID,
		CreatedAt:   &now,
	}
	issue, _, err := g.gitlab.Issues.CreateIssue(req.ProjectID, opt3)
	if err != nil || issue == nil {
		// if create failed then query from remote, if got then return
		opt := gogitlab.ListProjectIssuesOptions{
			ListOptions: gogitlab.ListOptions{
				Page:    1,
				PerPage: 20,
			},
			Search: &req.Title,
		}
		issues, _, err2 := g.gitlab.Issues.ListProjectIssues(req.ProjectID, &opt)
		if err2 != nil {
			return nil, fmt.Errorf("create issue failed: %v, query failed: %v", err, err2)
		}

		matched := false
		for idx, v := range issues {
			if strings.Compare(req.Title, v.Title) == 0 &&
				strings.Compare(req.Desc, v.Description) == 0 {
				//	matched
				issue = issues[idx]
				matched = true
			}
		}

		if !matched || issue == nil {
			return nil, fmt.Errorf("[matched: %v] create issue failed: %v", matched, err)
		}

	}

	return &CreateIssueResult{
		ID:     issue.ID,
		IID:    issue.IID,
		WebURL: issue.WebURL,
	}, nil
}

func (g gitlabOperator) CreateMergeRequest(ctx context.Context, req *CreateMergeRequest) (*CreateMergeResult, error) {
	opt5 := &gogitlab.CreateMergeRequestOptions{
		Title:        &req.Title,
		Description:  &req.Desc,
		MilestoneID:  &req.MilestoneID,
		SourceBranch: &req.SrcBranch,
		TargetBranch: &req.TargetBranch,
		// AssigneeID:         nil,
		// AssigneeIDs:        nil,
		// TargetProjectID:    nil,
		// RemoveSourceBranch: true,
	}
	mr, _, err := g.gitlab.MergeRequests.CreateMergeRequest(req.ProjectID, opt5)
	if err != nil || mr == nil {
		// if create failed then query from remote, if got then return
		opt := gogitlab.ListProjectMergeRequestsOptions{
			ListOptions: gogitlab.ListOptions{
				Page:    1,
				PerPage: 20,
			},
			Search:       &req.Title,
			TargetBranch: &req.TargetBranch,
			SourceBranch: &req.SrcBranch,
			// IIDs:         []int{req.IssueIID},
		}
		mergeRequests, _, err2 := g.gitlab.MergeRequests.ListProjectMergeRequests(req.ProjectID, &opt)
		if err2 != nil {
			return nil, fmt.Errorf("create merge request failed: %v, query failed: %v", err, err2)
		}

		matched := false
		for idx, v := range mergeRequests {
			if strings.Compare(req.Title, v.Title) == 0 &&
				strings.Compare(req.Desc, v.Description) == 0 {
				//	matched
				mr = mergeRequests[idx]
				matched = true
			}
		}

		if !matched || mr == nil {
			return nil, fmt.Errorf("[matched: %v] create merge request failed: %v", matched, err)
		}
	}

	return &CreateMergeResult{
		ID:     mr.ID,
		WebURL: mr.WebURL,
	}, nil
}

func (g gitlabOperator) ListMilestones(ctx context.Context, req *ListMilestoneRequest) (*ListMilestoneResult, error) {
	var active = "active"
	ms, _, err := g.gitlab.Milestones.ListMilestones(req.ProjectID, &gogitlab.ListMilestonesOptions{
		ListOptions: gogitlab.ListOptions{
			Page:    req.Page,
			PerPage: req.PerPage,
		},
		// Search: nil,
		// IIDs:   nil,
		// Title:  nil,
		State: &active,
	})
	if err != nil {
		return nil, errors.Wrap(err, "query milestones failed")
	}

	result := new(ListMilestoneResult)
	result.Data = make([]MilestoneShort, 0, len(ms))
	for _, v := range ms {
		result.Data = append(result.Data, MilestoneShort{
			ID:   v.ID,
			IID:  v.IID,
			Name: v.Title,
		})
	}

	return result, nil
}

func (g gitlabOperator) ListProjects(ctx context.Context, req *ListProjectRequest) (*ListProjectResult, error) {
	opt := &gogitlab.ListProjectsOptions{
		ListOptions: gogitlab.ListOptions{
			Page:    req.Page,
			PerPage: req.PerPage,
		},
		Search: &req.ProjectName,
	}
	projects, _, err := g.gitlab.Projects.ListProjects(opt)
	if err != nil {
		return nil, err
	}

	result := new(ListProjectResult)
	result.Data = make([]ProjectShort, 0, len(projects))
	for _, v := range projects {
		result.Data = append(result.Data, ProjectShort{
			ID:     v.ID,
			Name:   v.Name,
			WebURL: v.WebURL,
		})
	}

	return result, nil
}
