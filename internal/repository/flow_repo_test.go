package repository_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/yeqown/gitlab-flow/internal/repository"
	"github.com/yeqown/gitlab-flow/internal/repository/impl"
)

type flowRepoTestSuite struct {
	suite.Suite

	repo repository.IFlowRepository
}

func (s *flowRepoTestSuite) SetupTest() {
	s.T().Log("called")
	s.repo = impl.NewBasedSqlite3(impl.ConnectDB("./secret", true))
	s.T().Logf("%+v", s.repo)
}

func (s *flowRepoTestSuite) TearDownSuite() {
	// do nothing
}

func (s *flowRepoTestSuite) Test_batchCreate_WithTx() {
	branches := []*repository.BranchDO{
		{
			ProjectID:   11,
			MilestoneID: 22,
			IssueIID:    33,
			BranchName:  "2312312",
		},
		{
			ProjectID:   1111,
			MilestoneID: 2222,
			IssueIID:    3333,
			BranchName:  "2312312-asdasda",
		},
	}

	issues := []*repository.IssueDO{
		{
			IssueIID:      1,
			Title:         "23123",
			Desc:          "123123",
			ProjectID:     1,
			MilestoneID:   1,
			RelatedBranch: "123123",
			WebURL:        "asdasdasd",
		},
		{
			IssueIID:      2,
			Title:         "23123222",
			Desc:          "123123222",
			ProjectID:     2,
			MilestoneID:   2,
			RelatedBranch: "123123",
			WebURL:        "asdasdasd-123123",
		},
	}

	tx := s.repo.StartTransaction()
	err := s.repo.BatchCreateBranch(branches, tx)
	s.Nil(err)

	err = s.repo.BatchCreateIssue(issues, tx)
	s.Nil(err)

	err = s.repo.CommitTransaction(tx)
	s.Nil(err)
}

func (s *flowRepoTestSuite) Test_batchCreate_WithoutTx() {
	s.NotNil(s.repo)

	branches := []*repository.BranchDO{
		{
			ProjectID:   1,
			MilestoneID: 2,
			IssueIID:    3,
			BranchName:  "2312312",
		},
		{
			ProjectID:   1333,
			MilestoneID: 2222,
			IssueIID:    3,
			BranchName:  "2312312-asdasda",
		},
	}
	err := s.repo.BatchCreateBranch(branches)
	s.Nil(err)
}

func (s *flowRepoTestSuite) Test_CreateBranch_withTx() {
	s.NotNil(s.repo)

	b := &repository.BranchDO{
		ProjectID:   112312,
		MilestoneID: 212312,
		IssueIID:    1231233,
		BranchName:  "with tx",
	}

	tx := s.repo.StartTransaction()
	err := s.repo.SaveBranch(b, tx)
	s.Nil(err)
	err = s.repo.CommitTransaction(tx)
	s.Nil(err)
}

func (s *flowRepoTestSuite) Test_CreateBranch_withoutTx() {
	s.NotNil(s.repo)

	b := &repository.BranchDO{
		ProjectID:   112312,
		MilestoneID: 212312,
		IssueIID:    343534,
		BranchName:  "without tx",
	}

	err := s.repo.SaveBranch(b)
	s.Nil(err)
}

func Test_flowRepo(t *testing.T) {
	suite.Run(t, new(flowRepoTestSuite))
}
