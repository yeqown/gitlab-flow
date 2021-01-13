package internal

import (
	"context"
	"strings"

	gitop "github.com/yeqown/gitlab-flow/internal/git-operator"
	gitlabop "github.com/yeqown/gitlab-flow/internal/gitlab-operator"
	"github.com/yeqown/gitlab-flow/internal/repository"
	"github.com/yeqown/gitlab-flow/internal/repository/impl"
	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/pkg/errors"
	"github.com/yeqown/log"
)

// flowImpl implement IFlow.
type flowImpl struct {
	// ctx contains all properties of those flow want to take care.
	ctx *types.FlowContext

	// gitlabOperator operate remote gitlab repository.
	gitlabOperator gitlabop.IGitlabOperator
	// gitOperator operate local git client.
	gitOperator gitop.IGitOperator
	// repo operate local data repository which persists flow data.
	repo repository.IFlowRepository
}

func NewFlow(ctx *types.FlowContext) IFlow {
	if ctx == nil {
		panic("empty FlowContext initialized")
	}
	flow := &flowImpl{
		ctx:            ctx,
		gitlabOperator: gitlabop.NewGitlabOperator(ctx.Conf.AccessToken, ctx.Conf.GitlabAPIURL),
		gitOperator:    gitop.New(ctx.CWD),
		repo:           impl.NewBasedSqlite3(nil), // TODO(@yeqown): fill this
	}

	// flowContext with null project information, so we need to fill it.
	if err := flow.fillContextWithProject(); err != nil {
		panic(err)
	}

	return flow
}

// fillContextWithProject .
// FlowContext with null project information, so we need to fill it.
func (f flowImpl) fillContextWithProject() error {
	// DONE(@yeqown): fill project information from local repository or remote gitlab repository.
	projectName := extractProjectNameFromCWD(f.ctx.CWD)
	project := new(types.ProjectBasics)
	project.Name = projectName

	// get from local
	out, err := f.repo.QueryProject(&repository.ProjectDO{ProjectName: projectName})
	if err == nil {
		project.ID = out.ProjectID
		project.WebURL = out.WebURL
		return nil
	}
	log.WithFields(log.Fields{"project": projectName}).Warn("could not found from local.")

	// query from remote repository.
	result, err := f.gitlabOperator.ListProjects(context.Background(), &gitlabop.ListProjectRequest{
		Page:        1,
		PerPage:     20,
		ProjectName: projectName,
	})

	// could not hit stage 1
	errCouldNotFound := errors.Wrapf(err, "could not match project(%s) neither local nor remote", projectName)
	if err != nil {
		return errCouldNotFound
	}

	// found and match
	for _, v := range result.Data {
		if strings.Compare(projectName, v.Name) == 0 {
			// matched
			project.ID = v.ID
			project.WebURL = v.WebURL
			// DONE(@yeqown): save into local database
			projectDO := repository.ProjectDO{
				ProjectName: project.Name,
				ProjectID:   project.ID,
				LocalDir:    f.ctx.CWD,
				WebURL:      project.WebURL,
			}
			if err = f.repo.SaveProject(&projectDO); err != nil {
				log.WithField("project", projectDO).Warn("could not save project")
			}

			return nil
		}
	}

	return errCouldNotFound
}

func (f flowImpl) FeatureStart(title, desc string) error {
	// query from local to prevent duplicate create.
	// create remote
	panic("implement me")
}

func (f flowImpl) FeatureDebugging(featureBranchName string) error {
	panic("implement me")
}

func (f flowImpl) FeatureTest(featureBranchName string) error {
	panic("implement me")
}

func (f flowImpl) FeatureRelease(featureBranchName string) error {
	panic("implement me")
}

func (f flowImpl) FeatureStartIssue(featureBranchName string, params ...string) error {
	panic("implement me")
}

func (f flowImpl) FeatureFinishIssue(featureBranchName, issueBranchName string) error {
	panic("implement me")
}

func (f flowImpl) HotfixStart(title, desc string) error {
	panic("implement me")
}

func (f flowImpl) HotfixRelease(hotfixBranchName string) error {
	panic("implement me")
}

// SyncMilestone rebuilds local data related to `milestoneID`
//
// 1. pull milestone + MergeRequest + Issues by `milestoneID`.
// 2. parse `IssueID` from MR description.
// 3. handle and save data.
//
func (f flowImpl) SyncMilestone(milestoneID int) error {
	ctx := context.Background()
	projectId := f.ctx.Project.ID

	milestoneResult, err := f.gitlabOperator.GetMilestone(ctx, &gitlabop.GetMilestoneRequest{
		ProjectID:   projectId,
		MilestoneID: milestoneID,
	})
	if err != nil {
		return errors.Wrap(err, "get milestone failed")
	}

	milestoneMRsResult, err := f.gitlabOperator.
		GetMilestoneMergeRequests(ctx, &gitlabop.GetMilestoneMergeRequestsRequest{
			ProjectID:   projectId,
			MilestoneID: milestoneID,
		})
	if err != nil {
		return errors.Wrap(err, "get milestone merge requests failed")
	}

	milestoneIssuesResult, err := f.gitlabOperator.GetMilestoneIssues(ctx,
		&gitlabop.GetMilestoneIssuesRequest{
			ProjectID:   projectId,
			MilestoneID: milestoneID,
		})
	if err != nil {
		return errors.Wrap(err, "get milestone issues failed")
	}

	// format data
	i, mr, b, branchName := f.syncFormatResultIntoDO(milestoneResult, milestoneMRsResult.Data, milestoneIssuesResult.Data)
	_ = branchName
	log.WithFields(log.Fields{
		"milestoneResult":       milestoneResult,
		"milestoneMRsResult":    milestoneMRsResult,
		"milestoneIssuesResult": milestoneIssuesResult,
	}).Debugf("syncFormatResultIntoDO calling")

	// save data models
	tx := f.repo.StartTransaction()
	err = f.repo.SaveMilestone(&repository.MilestoneDO{
		ProjectID:   projectId,
		MilestoneID: milestoneResult.ID,
		Title:       milestoneResult.Title,
		Desc:        milestoneResult.Description,
		WebURL:      milestoneResult.WebURL,
	}, tx)
	if err != nil {
		return errors.Wrap(err, "save milestone failed")
	}
	err = f.repo.BatchCreateIssue(i, tx)
	if err != nil {
		return errors.Wrap(err, "save issues failed")
	}
	err = f.repo.BatchCreateMergeRequest(mr, tx)
	if err != nil {
		return errors.Wrap(err, "save merge requests failed")
	}
	err = f.repo.BatchCreateBranch(b, tx)
	if err != nil {
		return errors.Wrap(err, "save branches failed")
	}
	if err = f.repo.CommitTransaction(tx); err != nil {
		return errors.Wrap(err, "CommitTransaction failed")
	}

	_ = f.gitOperator.FetchOrigin()

	return nil
}

// syncFormatResultIntoDO rebuild local data from remote gitlab repository.
// @return issues, mrs, branches, featureBranchName
func (f flowImpl) syncFormatResultIntoDO(
	milestone *gitlabop.GetMilestoneResult,
	mrs []gitlabop.MergeRequestShort,
	issues []gitlabop.IssueShort,
) ([]*repository.IssueDO, []*repository.MergeRequestDO, []*repository.BranchDO, string) {
	var (
		issueDO  = make([]*repository.IssueDO, 0, 10)
		mrDO     = make([]*repository.MergeRequestDO, 0, 10)
		branchDO = make([]*repository.BranchDO, 0, 10)

		c                 = make(map[int]*repository.IssueDO)
		featureBranchName string
		projectID         = f.ctx.Project.ID
		milestoneID       = milestone.ID
	)

	// pre handle issue into cache
	for _, v := range issues {
		c[v.IID] = &repository.IssueDO{
			IssueIID:    v.IID,
			Title:       v.Title,
			Desc:        v.Description,
			ProjectID:   projectID,
			MilestoneID: milestoneID,
			WebURL:      v.WebURL,
			//RelatedBranch: ,
		}
	}

	for _, mr := range mrs {
		issueIID := parseIssueIIDFromMergeRequestIssue(mr.Description)

		log.WithFields(log.Fields{
			"id":       mr.ID,
			"desc":     mr.Description,
			"issueIID": issueIID,
		}).
			Debug("sync handle merge request")

		if issueIID != 0 {
			// 如果 MR 关联了Issue, 才会处理该issue到本地数据中
			issue, ok := c[issueIID]
			if !ok {
				// 记录日志
				log.WithFields(log.Fields{
					"issueIID":    issueIID,
					"issues":      issues,
					"issuesCache": c,
				}).Warn("从里程碑issue清单中定位issue失败")

			}

			// 生成数据
			issueDO = append(issueDO, &repository.IssueDO{
				IssueIID:      issueIID,
				Title:         issue.Title,
				Desc:          issue.Desc,
				ProjectID:     projectID,
				MilestoneID:   milestoneID,
				RelatedBranch: mr.SourceBranch,
				WebURL:        issue.WebURL,
			})
		}

		mrDO = append(mrDO, &repository.MergeRequestDO{
			ProjectID:      projectID,
			MilestoneID:    milestoneID,
			IssueIID:       issueIID,
			MergeRequestID: mr.ID,
			SourceBranch:   mr.SourceBranch,
			TargetBranch:   mr.TargetBranch,
			WebURL:         mr.WebURL,
		})

		// 打印迭代分支
		if featureBranchName == "" && strings.HasPrefix(mr.SourceBranch, types.FeatureBranchPrefix) {
			featureBranchName = mr.SourceBranch
		}
		if featureBranchName == "" && strings.HasPrefix(mr.TargetBranch, types.FeatureBranchPrefix) {
			featureBranchName = mr.TargetBranch
		}

		branchDO = append(branchDO, &repository.BranchDO{
			ProjectID:   projectID,
			MilestoneID: milestoneID,
			IssueIID:    issueIID,
			BranchName:  mr.SourceBranch,
		})

		// 需要把 targetBranch 也同步
		if notBuiltinBranch(mr.TargetBranch) {
			branchDO = append(branchDO, &repository.BranchDO{
				ProjectID:   projectID,
				MilestoneID: milestoneID,
				IssueIID:    issueIID,
				BranchName:  mr.TargetBranch,
			})
		}
	}

	return issueDO, mrDO, branchDO, featureBranchName
}
