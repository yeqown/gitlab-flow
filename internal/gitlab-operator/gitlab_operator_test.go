package gitlabop

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/stretchr/testify/suite"
)

type gitlabOperatorTestSuite struct {
	suite.Suite

	op          IGitlabOperator
	projectID   int
	milestoneID int
	issueIID    int
}

func (g *gitlabOperatorTestSuite) SetupSuite() {
	r, err := os.Open("./testdata/secret.json")
	g.Require().Nil(err)

	data, err := ioutil.ReadAll(r)
	g.Require().Nil(err)
	var c = new(Config)
	err = json.Unmarshal(data, c)
	g.Require().Nil(err)

	g.op = NewGitlabOperator(c.AccessToken, c.ApiURL)

	// this only could be test locally
	g.projectID = 851
	g.milestoneID = 1140
	g.issueIID = 18
}

func (g *gitlabOperatorTestSuite) TearDownSuite() {
	// do nothing
}

func (g gitlabOperatorTestSuite) Test_CreateBranch() {
	ctx := context.Background()
	req := CreateBranchRequest{
		TargetBranch: "feature/branch",
		SrcBranch:    types.MasterBranch.String(),
		ProjectID:    g.projectID,
	}
	result, err := g.op.CreateBranch(ctx, &req)
	g.Nil(err)
	g.NotNil(result)
	g.T().Logf("result=%+v", result)
}

func (g gitlabOperatorTestSuite) Test_CreateMilestone() {
	ctx := context.Background()
	req := CreateMilestoneRequest{
		Title:     "milestoneTest",
		Desc:      "milestoneDesc",
		ProjectID: g.projectID,
	}
	result, err := g.op.CreateMilestone(ctx, &req)
	g.Nil(err)
	g.NotNil(result)
	g.T().Logf("result=%+v", result)
}

func (g gitlabOperatorTestSuite) Test_CreateIssue() {
	ctx := context.Background()
	req := CreateIssueRequest{
		Title:         "milestoneTest",
		Desc:          "milestoneDesc",
		RelatedBranch: "feature/branch",
		MilestoneID:   g.milestoneID,
		ProjectID:     g.projectID,
	}
	result, err := g.op.CreateIssue(ctx, &req)
	g.Nil(err)
	g.NotNil(result)
	g.T().Logf("result=%+v", result)
}

func (g gitlabOperatorTestSuite) Test_CreateMergeRequest() {
	ctx := context.Background()
	req := CreateMergeRequest{
		Title:        "MR Title",
		Desc:         "MR Desc",
		SrcBranch:    "feature/branch",
		TargetBranch: types.MasterBranch.String(),
		MilestoneID:  g.milestoneID,
		IssueIID:     g.issueIID,
		ProjectID:    g.projectID,
	}
	result, err := g.op.CreateMergeRequest(ctx, &req)
	g.Nil(err)
	g.NotNil(result)
	g.T().Logf("result=%+v", result)
}

func Test_gitlabOperator(t *testing.T) {
	suite.Run(t, new(gitlabOperatorTestSuite))
}
