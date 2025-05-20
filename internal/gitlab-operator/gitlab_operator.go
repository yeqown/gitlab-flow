package gitlabop

import "context"

// IGitlabOperator contains all operations those manage repository,
// milestones, branch, issue and merge requests.
type IGitlabOperator interface {
	// CreateBranch create a branch on remote gitlab repository, but this would check remote
	// resource if create failed.
	CreateBranch(ctx context.Context, req *CreateBranchRequest) (*CreateBranchResult, error)

	// CreateMilestone create a milestone on remote gitlab repository, but this would check remote
	// resource if create failed.
	CreateMilestone(ctx context.Context, req *CreateMilestoneRequest) (*CreateMilestoneResult, error)
	GetMilestone(ctx context.Context, req *GetMilestoneRequest) (*GetMilestoneResult, error)
	GetMilestoneMergeRequests(
		ctx context.Context, req *GetMilestoneMergeRequestsRequest) (*GetMilestoneMergeRequestsResult, error)
	GetMilestoneIssues(ctx context.Context, req *GetMilestoneIssuesRequest) (*GetMilestoneIssuesResult, error)

	// CreateIssue create an issue on remote repository, but this would check remote
	// resource if create failed.
	CreateIssue(ctx context.Context, req *CreateIssueRequest) (*CreateIssueResult, error)
	// CreateMergeRequest create an merge request on remote repository, but this would check remote
	// resource if create failed.
	CreateMergeRequest(ctx context.Context, req *CreateMergeRequest) (*CreateMergeResult, error)
	MergeMergeRequest(ctx context.Context, req *MergeMergeRequest) error

	ListMilestones(ctx context.Context, req *ListMilestoneRequest) (*ListMilestoneResult, error)
	ListProjects(ctx context.Context, req *ListProjectRequest) (*ListProjectResult, error)
}

// CreateBranchRequest
type CreateBranchRequest struct {
	TargetBranch string
	SrcBranch    string
	ProjectID    int
	// MilestoneID  int
	// IssueID      int
}

type CreateBranchResult struct {
	Name   string
	WebURL string
}

// CreateMilestoneRequest
type CreateMilestoneRequest struct {
	Title     string
	Desc      string
	ProjectID int
}

type CreateMilestoneResult struct {
	ID     int
	WebURL string
}

// GetMilestoneRequest .
type GetMilestoneRequest struct {
	MilestoneID int
	ProjectID   int
}

type GetMilestoneResult struct {
	ID          int
	Title       string
	Description string
	WebURL      string
}

// GetMilestoneMergeRequestsRequest
type GetMilestoneMergeRequestsRequest struct {
	MilestoneID int
	ProjectID   int
}

type GetMilestoneMergeRequestsResult struct {
	Data []MergeRequestShort
}

type MergeRequestShort struct {
	ID           int
	IID          int
	Title        string
	Description  string
	WebURL       string
	SourceBranch string
	TargetBranch string
}

// GetMilestoneIssuesRequest
type GetMilestoneIssuesRequest struct {
	MilestoneID int
	ProjectID   int
}

type GetMilestoneIssuesResult struct {
	Data []IssueShort
}

type IssueShort struct {
	ID          int
	IID         int
	Title       string
	Description string
	WebURL      string
	ProjectID   int
	MilestoneID int
}

// CreateIssueRequest
type CreateIssueRequest struct {
	Title, Desc, RelatedBranch string
	MilestoneID                int
	ProjectID                  int
}

type CreateIssueResult struct {
	ID     int
	IID    int
	WebURL string
}

// CreateMergeRequest
type CreateMergeRequest struct {
	Title, Desc, SrcBranch, TargetBranch string
	MilestoneID, IssueIID                int
	ProjectID                            int
	AutoMerge                            bool
}

type CreateMergeResult struct {
	ID     int
	IID    int
	WebURL string
}

type MergeMergeRequest struct {
	MergeRequestID int
	ProjectID      int
}

// ListMilestoneRequest
type ListMilestoneRequest struct {
	Page      int
	PerPage   int
	ProjectID int
}

type MilestoneShort struct {
	ID          int
	IID         int
	Name        string
	WebURL      string
	Description string
}

type ListMilestoneResult struct {
	Data []MilestoneShort
}

type ListProjectRequest struct {
	Page        int
	PerPage     int
	ProjectName string
}

type ProjectShort struct {
	ID     int
	Name   string
	WebURL string
}

type ListProjectResult struct {
	Data []ProjectShort
}

type IGitlabOauth2Support interface {
	// Enter is an asynchronous process that would not return accessToken and refreshToken synchronized.
	// IGitlabOauth2Support.Load will return the refreshToken and accessToken after signaling.
	Enter(refreshToken string) (err error)

	// Load only uses this after any signal from Enter channel. Blocked method.
	Load() (accessToken, refreshToken string)
}
