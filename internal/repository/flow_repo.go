package repository

import (
	"github.com/pkg/errors"
	gorm2 "gorm.io/gorm"
)

// IFlowRepository is used to manage local flow data.
type IFlowRepository interface {
	StartTransaction() *gorm2.DB
	CommitTransaction(tx *gorm2.DB) error

	SaveProject(m *ProjectDO, txs ...*gorm2.DB) error
	QueryProject(filter *ProjectDO) (*ProjectDO, error)

	SaveMilestone(m *MilestoneDO, txs ...*gorm2.DB) error
	QueryMilestone(filter *MilestoneDO) (*MilestoneDO, error)
	QueryMilestones(filter *MilestoneDO) ([]*MilestoneDO, error)

	SaveBranch(m *BranchDO, txs ...*gorm2.DB) error
	BatchCreateBranch(records []*BranchDO, txs ...*gorm2.DB) error
	QueryBranch(filter *BranchDO) (*BranchDO, error)

	SaveIssue(m *IssueDO, txs ...*gorm2.DB) error
	BatchCreateIssue(records []*IssueDO, txs ...*gorm2.DB) error
	QueryIssue(filter *IssueDO) (*IssueDO, error)
	QueryIssues(filter *IssueDO) ([]*IssueDO, error)

	SaveMergeRequest(m *MergeRequestDO, txs ...*gorm2.DB) error
	BatchCreateMergeRequest(records []*MergeRequestDO, txs ...*gorm2.DB) error
	QueryMergeRequest(filter *MergeRequestDO) (*MergeRequestDO, error)
	QueryMergeRequests(filter *MergeRequestDO) ([]*MergeRequestDO, error)
}

func IsErrNotFound(err error) bool {
	return errors.Is(err, gorm2.ErrRecordNotFound)
}

// ProjectDO data model
type ProjectDO struct {
	gorm2.Model

	ProjectName string `gorm:"column:name"`
	ProjectID   int    `gorm:"column:project_id"`
	LocalDir    string `gorm:"column:local_dir"`
	WebURL      string `gorm:"column:web_url"`
}

func (m *ProjectDO) TableName() string {
	return "project"
}

// MilestoneDO data model
type MilestoneDO struct {
	gorm2.Model

	ProjectID   int    `gorm:"column:project_id"`
	MilestoneID int    `gorm:"column:milestone_id"`
	Title       string `gorm:"column:title"`
	Desc        string `gorm:"column:desc"`
	WebURL      string `gorm:"column:web_url"`
}

func (m *MilestoneDO) TableName() string {
	return "project_milestone"
}

// BranchDO data model
type BranchDO struct {
	gorm2.Model

	ProjectID   int    `gorm:"column:project_id"`
	MilestoneID int    `gorm:"column:milestone_id"`
	IssueIID    int    `gorm:"column:issue_iid"`
	BranchName  string `gorm:"column:branch_name"`
}

func (m *BranchDO) TableName() string {
	return "project_branch"
}

// IssueDO data model
type IssueDO struct {
	gorm2.Model

	IssueIID      int    `gorm:"column:issue_iid"`
	Title         string `gorm:"column:title"`
	Desc          string `gorm:"column:desc"`
	ProjectID     int    `gorm:"column:project_id"`
	MilestoneID   int    `gorm:"column:milestone_id"`
	RelatedBranch string `gorm:"column:related_branch"`
	WebURL        string `gorm:"column:web_url"`
}

func (m *IssueDO) TableName() string {
	return "project_issue"
}

// MergeRequestDO data model
type MergeRequestDO struct {
	gorm2.Model

	ProjectID      int    `gorm:"column:project_id"`
	MilestoneID    int    `gorm:"column:milestone_id"`
	IssueIID       int    `gorm:"column:issue_iid"`
	MergeRequestID int    `gorm:"column:merge_request_id"`
	SourceBranch   string `gorm:"column:source_branch"`
	TargetBranch   string `gorm:"column:target_branch"`
	WebURL         string `gorm:"column:web_url"`
}

func (m *MergeRequestDO) TableName() string {
	return "project_merge_request"
}
