package gitlabop

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	gogitlab "github.com/xanzy/go-gitlab"
	"github.com/yeqown/log"
)

// gitlabOperator implement IGitlabOperator to operate remote gitlab repository.
type gitlabOperator struct {
	gitlab *gogitlab.Client
}

// NewGitlabOperator generate IGitlabOperator.
func NewGitlabOperator(accessToken, APIURL string) IGitlabOperator {
	gitlab, err := gogitlab.NewClient(accessToken, gogitlab.WithBaseURL(APIURL))
	if err != nil {
		log.WithFields(log.Fields{
			"accessToken": accessToken,
			"apiURL":      APIURL,
		}).Errorf("could not generate gitlab client: %v", err)
		// could not go ahead if could not initialize gitlab client.
		panic(err)
	}

	return &gitlabOperator{gitlab: gitlab}
}

func (g gitlabOperator) CreateBranch(ctx context.Context, req *CreateBranchRequest) (*CreateBranchResult, error) {
	ref := req.SrcBranch
	opt := &gogitlab.CreateBranchOptions{
		Branch: &req.TargetBranch,
		Ref:    &ref,
	}
	branch, _, err := g.gitlab.Branches.CreateBranch(req.ProjectID, opt)
	if err != nil {
		return nil, errors.Wrap(err, "create branch failed")
	}

	return &CreateBranchResult{
		BranchName: branch.Name,
		BranchURL:  branch.WebURL,
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
	if err != nil {
		return nil, errors.Wrap(err, "CreateMilestone failed")
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
	if err != nil {
		return nil, errors.Wrap(err, "create Issue failed")
	}

	return &CreateIssueResult{
		ID:     issue.ID,
		IID:    issue.IID,
		WebURL: issue.WebURL,
	}, nil
}

func (g gitlabOperator) CreateMergeRequest(ctx context.Context, req *CreateMergeRequest) (*CreateMergeResult, error) {
	// MergeRequest is still Work in progress
	req.Title = "WIP: " + req.Title

	// Closes related issue
	if req.IssueIID != 0 {
		req.Desc = fmt.Sprintf("Closes #%d\n", req.IssueIID) + req.Desc
	}

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
	if err != nil {
		return nil, errors.Wrap(err, "create merge request failed")
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
