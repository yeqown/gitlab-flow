package internal

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/yeqown/gitlab-flow/pkg"

	"github.com/AlecAivazis/survey/v2"

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
	log.
		WithField("context", ctx).
		Debugf("constructing flow")

	if ctx == nil {
		log.Fatal("empty FlowContext initialized")
		panic("can not reach")
	}
	flow := &flowImpl{
		ctx:            ctx,
		gitlabOperator: gitlabop.NewGitlabOperator(ctx.Conf.AccessToken, ctx.Conf.GitlabAPIURL),
		gitOperator:    gitop.NewBasedCmd(ctx.CWD),
		repo:           impl.NewBasedSqlite3(impl.ConnectDB(ctx.ConfPath, ctx.Conf.DebugMode)),
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
	// FIXME(@yeqown): projectName would be different from project path, use git repository name as project name.
	projectName := extractProjectNameFromCWD(f.ctx.CWD)
	project := new(types.ProjectBasics)
	project.Name = projectName

	// get from local
	out, err := f.repo.QueryProject(&repository.ProjectDO{ProjectName: projectName})
	if err == nil {
		project.ID = out.ProjectID
		project.WebURL = out.WebURL
		f.ctx.Project = project
		return nil
	}
	log.
		WithFields(log.Fields{"project": projectName}).
		Warn("could not found from local")

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
			log.
				WithFields(log.Fields{
					"project":   projectName,
					"projectID": v.ID,
					"webURL":    v.WebURL,
				}).
				Debug("hit project in remote")
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
				log.
					WithField("project", projectDO).
					Warn("could not save project")
			}
			f.ctx.Project = project
			return nil
		}
	}

	return errCouldNotFound
}

func (f flowImpl) FeatureBegin(title, desc string) error {
	log.
		WithFields(log.Fields{
			"title": title,
			"desc":  desc,
		}).
		Debug("FeatureBegin called")

	// create milestone
	result, err := f.createMilestone(title, desc)
	if err != nil {
		return errors.Wrap(err, "CreateMilestone failed")
	}

	// create feature branch
	featureBranchName := genFeatureBranchName(title)
	_, err = f.createBranch(featureBranchName, types.MasterBranch.String(), result.ID, 0)
	if err != nil {
		return err
	}

	return nil
}

func (f flowImpl) FeatureDebugging(featureBranchName string) error {
	return f.featureProcessMR(featureBranchName, types.DevBranch)
}

func (f flowImpl) FeatureTest(featureBranchName string) error {
	return f.featureProcessMR(featureBranchName, types.TestBranch)
}

func (f flowImpl) FeatureRelease(featureBranchName string) error {
	return f.featureProcessMR(featureBranchName, types.MasterBranch)
}

func (f flowImpl) FeatureBeginIssue(featureBranchName string, title, desc string) error {
	// DONE(@yeqown): is featureBranchName is empty, use current branch name.
	if featureBranchName == "" {
		featureBranchName, _ = f.gitOperator.CurrentBranch()
	}

	featureBranchName = genFeatureBranchName(featureBranchName)
	featureBranch, err := f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  f.ctx.Project.ID,
		BranchName: featureBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate feature branch from local failed")
	}

	// query milestone
	milestone, err := f.repo.QueryMilestone(&repository.MilestoneDO{
		ProjectID:   f.ctx.Project.ID,
		MilestoneID: featureBranch.MilestoneID,
	})
	if err != nil {
		return errors.Wrap(err, "locate milestone failed")
	}

	if len(title) == 0 {
		title = milestone.Title
	}
	if len(desc) == 0 {
		desc = milestone.Desc
	}

	// create and save issue
	issue, err := f.createIssue(title, desc, featureBranch.BranchName, milestone.MilestoneID)
	if err != nil {
		return err
	}

	// create issue branch
	issueBranchName := genIssueBranchName(milestone.Title, issue.IID)
	issueBranch, err := f.createBranch(issueBranchName, featureBranch.BranchName, milestone.MilestoneID, issue.IID)
	if err != nil {
		return err
	}

	// create mr
	targetBranch := featureBranch.BranchName
	mrTitle := genMRTitle(issueBranchName, targetBranch)
	mr, err := f.createMergeRequest(mrTitle, desc, milestone.MilestoneID, issue.IID, issueBranchName, targetBranch)
	if err != nil {
		return nil
	}
	log.
		WithFields(log.Fields{
			"issue_branch":  issueBranch.WebURL,
			"merge_request": mr.WebURL,
		}).
		Debug("create issue finished")

	return nil
}

func (f flowImpl) FeatureFinishIssue(featureBranchName, issueBranchName string) (err error) {
	var milestoneID = 0

	// DONE(@yeqown): if issueBranchName is empty, make current branch name as default.
	if issueBranchName == "" {
		issueBranchName, _ = f.gitOperator.CurrentBranch()
	}
	if issueBranchName == "" {
		return errors.New("issue branch could not be empty")
	}
	if b, err := f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  f.ctx.Project.ID,
		BranchName: issueBranchName,
	}); err != nil {
		return errors.Wrapf(err, "locate issue branch(%s) failed", issueBranchName)
	} else {
		milestoneID = b.MilestoneID
	}

	// DONE(@yeqown) get feature branch name from issueBranchName
	if featureBranchName == "" {
		featureBranchName = parseFeaturenameFromIssueName(issueBranchName)
	}
	if featureBranchName == "" {
		return errors.New("feature branch could not be empty")
	}
	featureBranchName = genFeatureBranchName(featureBranchName)

	if _, err = f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:   f.ctx.Project.ID,
		BranchName:  featureBranchName,
		MilestoneID: milestoneID,
	}); err != nil {
		return errors.Wrapf(err, "locate feature branch(%s) failed", featureBranchName)
	}

	// locate MR
	mr, err := f.repo.QueryMergeRequest(&repository.MergeRequestDO{
		ProjectID:    f.ctx.Project.ID,
		MilestoneID:  milestoneID,
		SourceBranch: issueBranchName,
		TargetBranch: featureBranchName,
	})
	if err != nil {
		log.
			WithFields(log.Fields{
				"projectID":   f.ctx.Project.ID,
				"milestoneID": milestoneID,
				"error":       err,
			}).
			Error("locate MR failed")
		return errors.Wrap(err, "locate MR failed")
	}

	log.
		WithFields(log.Fields{
			"feature_branch":    featureBranchName,
			"issue_branch":      issueBranchName,
			"project_name":      f.ctx.Project.Name,
			"merge_request_url": mr.WebURL,
		}).
		Debug("issue info")

	f.printAndOpenBrowser("Issue Merge Request", mr.WebURL)

	return nil
}

func (f flowImpl) HotfixBegin(title, desc string) error {
	hotfixBranchName := genHotfixBranchName(title)

	// create ISSUE
	issue, err := f.createIssue(title, desc, hotfixBranchName, 0)
	if err != nil {
		return errors.Wrap(err, "create issue failed")
	}

	// create merge request
	branch, err := f.createBranch(hotfixBranchName, types.MasterBranch.String(), 0, issue.IID)
	if err != nil {
		return errors.Wrap(err, "create branch failed")
	}

	log.
		WithFields(log.Fields{
			"issue":  issue,
			"branch": branch,
		}).
		Debug("hotfix begin finished")

	return nil
}

func (f flowImpl) HotfixFinish(hotfixBranchName string) error {
	if hotfixBranchName == "" {
		hotfixBranchName, _ = f.gitOperator.CurrentBranch()
	}

	hotfixBranchName = genHotfixBranchName(hotfixBranchName)
	_, err := f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  f.ctx.Project.ID,
		BranchName: hotfixBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate hotfix branch failed")
	}

	// locate issue
	issue, err := f.repo.QueryIssue(&repository.IssueDO{
		ProjectID:     f.ctx.Project.ID,
		RelatedBranch: hotfixBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate issue failed")
	}

	// MR to master
	title := genMRTitle(hotfixBranchName, types.MasterBranch.String())
	mr, err := f.createMergeRequest(
		title, issue.Desc, 0, issue.IssueIID, hotfixBranchName, types.MasterBranch.String())
	if err != nil {
		return errors.Wrap(err, "create hotfix MR failed")
	}

	log.
		WithFields(log.Fields{
			"issue":        issue,
			"mergeRequest": mr,
		}).
		Debug("hotfix begin finished")

	return nil
}

// SyncMilestone rebuilds local data related to `milestoneID`
//
// 1. pull milestone + MergeRequest + Issues by `milestoneID`.
// 2. parse `IssueID` from MR description.
// 3. handle and save data.
//
func (f flowImpl) SyncMilestone(milestoneID int, interact bool) error {
	ctx := context.Background()
	projectId := f.ctx.Project.ID

	if interact && milestoneID == 0 {
		// if interact to choose milestone, and milestoneID is empty.
		result, err := f.gitlabOperator.ListMilestones(ctx, &gitlabop.ListMilestoneRequest{
			Page:      1,
			PerPage:   20,
			ProjectID: projectId,
		})
		if err != nil {
			return errors.Wrap(err, "list milestones failed")
		}

		milestoneOptions := make([]string, len(result.Data))
		for idx, v := range result.Data {
			milestoneOptions[idx] = fmt.Sprintf("%s >> %d >> %d", v.Name, v.ID, v.IID)
		}

		qs := []*survey.Question{
			{
				Name: "milestones",
				Prompt: &survey.Select{
					Message: "choosing one by moving your arrow up and down",
					Options: milestoneOptions,
				},
			},
		}
		r := struct {
			Milestone string `survey:"milestones"`
		}{}
		if err = survey.Ask(qs, &r); err != nil {
			return errors.Wrap(err, "survey.Ask failed")
		}

		milestoneID, _ = strconv.Atoi(strings.Split(r.Milestone, ">>")[1])
	}

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
				}).Warn("从里程碑issue清单中locate issue failed")

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

// createBranch .
func (f flowImpl) createBranch(
	targetBranchName, srcBranch string, milestoneID, issueIID int) (*gitlabop.CreateBranchResult, error) {
	wg := sync.WaitGroup{}
	ctx := context.Background()
	req := gitlabop.CreateBranchRequest{
		TargetBranch: targetBranchName,
		SrcBranch:    srcBranch,
		MilestoneID:  milestoneID,
		IssueID:      issueIID,
		ProjectID:    f.ctx.Project.ID,
	}

	result, err := f.gitlabOperator.CreateBranch(ctx, &req)
	if err != nil {
		return nil, errors.Wrap(err, "create branch failed [gitlab]")
	}

	// fetch origin branch and checkout
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := f.gitOperator.FetchOrigin(); err != nil {
			log.
				WithFields(log.Fields{"error": err}).
				Errorf("FetchOrigin failed")
			return
		}

		if err := f.gitOperator.Checkout(targetBranchName, false); err != nil {
			log.
				WithFields(log.Fields{
					"targetBranch": targetBranchName,
					"error":        err,
				}).
				Errorf("Checkout branch failed")
		}
	}()

	// 保存 featureBranch 记录
	wg.Add(1)
	go func() {
		wg.Done()
		m := repository.BranchDO{
			ProjectID:   f.ctx.Project.ID,
			MilestoneID: milestoneID,
			IssueIID:    issueIID,
			BranchName:  targetBranchName,
		}

		if err := f.repo.SaveBranch(&m); err != nil {
			log.
				WithFields(log.Fields{
					"branch": targetBranchName,
					"model":  m,
					"error":  err,
				}).
				Errorf("save branch data failed")
		}
	}()
	wg.Wait()

	return result, nil
}

// createMilestone create Milestone
func (f flowImpl) createMilestone(title, desc string) (*gitlabop.CreateMilestoneResult, error) {
	ctx := context.Background()
	result, err := f.gitlabOperator.CreateMilestone(ctx, &gitlabop.CreateMilestoneRequest{
		Title:     title,
		Desc:      desc,
		ProjectID: f.ctx.Project.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "CreateMilestone failed")
	}

	if err = f.repo.SaveMilestone(&repository.MilestoneDO{
		ProjectID:   f.ctx.Project.ID,
		MilestoneID: result.ID,
		Title:       title,
		Desc:        desc,
		WebURL:      result.WebURL,
	}); err != nil {
		log.WithFields(log.Fields{
			"milestone": result,
			"error":     err,
		}).Errorf("could not save milestone")
	}

	return result, nil
}

// createIssue .
func (f flowImpl) createIssue(title, desc, relatedBranch string, milestoneID int) (*gitlabop.CreateIssueResult, error) {
	ctx := context.Background()
	result, err := f.gitlabOperator.CreateIssue(ctx, &gitlabop.CreateIssueRequest{
		Title:         title,
		Desc:          desc,
		RelatedBranch: relatedBranch,
		MilestoneID:   milestoneID,
		ProjectID:     f.ctx.Project.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create Issue failed")
	}

	if err = f.repo.SaveIssue(&repository.IssueDO{
		IssueIID:      result.IID,
		Title:         title,
		Desc:          desc,
		ProjectID:     f.ctx.Project.ID,
		MilestoneID:   milestoneID,
		RelatedBranch: relatedBranch,
		WebURL:        result.WebURL,
	}); err != nil {
		log.
			WithFields(log.Fields{
				"issue": result,
				"error": err,
			}).
			Errorf("save Issue failed")
	}

	return result, nil
}

// CreateMergeRequest create MergeRequest
func (f flowImpl) createMergeRequest(
	title, desc string, milestoneID, issueIID int, srcBranch, targetBranch string,
) (*gitlabop.CreateMergeResult, error) {
	ctx := context.Background()
	// MergeRequest is still Work in progress
	title = "WIP: " + title
	// Closes related issue
	if issueIID != 0 {
		desc = fmt.Sprintf("Closes #%d\n", issueIID) + desc
	}

	result, err := f.gitlabOperator.CreateMergeRequest(ctx, &gitlabop.CreateMergeRequest{
		Title:        title,
		Desc:         desc,
		SrcBranch:    srcBranch,
		TargetBranch: targetBranch,
		MilestoneID:  milestoneID,
		IssueIID:     issueIID,
		ProjectID:    f.ctx.Project.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create MR failed")
	}
	log.
		WithFields(log.Fields{
			"id":     result.ID,
			"source": srcBranch,
			"target": targetBranch,
			"url":    result.WebURL,
			"title":  title,
		}).
		Debug("create mr success")

	if err = f.repo.SaveMergeRequest(&repository.MergeRequestDO{
		ProjectID:      f.ctx.Project.ID,
		MilestoneID:    milestoneID,
		IssueIID:       issueIID,
		MergeRequestID: result.ID,
		SourceBranch:   srcBranch,
		TargetBranch:   targetBranch,
		WebURL:         result.WebURL,
	}); err != nil {
		log.
			WithFields(log.Fields{
				"error": err,
			}).
			Errorf("save MR failed")
	}

	return result, nil
}

// featureProcessMR is a process for creating a merge request for feature branch to target branch
func (f flowImpl) featureProcessMR(featureBranchName string, target types.BranchTyp) error {
	if featureBranchName == "" {
		featureBranchName, _ = f.gitOperator.CurrentBranch()
	}

	featureBranchName = genFeatureBranchName(featureBranchName)
	featureBranch, err := f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  f.ctx.Project.ID,
		BranchName: featureBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate feature branch failed")
	}

	milestone, err := f.repo.QueryMilestone(&repository.MilestoneDO{
		MilestoneID: featureBranch.MilestoneID})
	if err != nil {
		return errors.Wrap(err, "locate milestone failed")
	}

	// create MR
	targetBranch := target.String()
	title := genMRTitle(featureBranchName, targetBranch)
	if _, err = f.createMergeRequest(
		title, milestone.Desc, milestone.MilestoneID, 0, featureBranch.BranchName, targetBranch); err != nil {
		return errors.Wrapf(err, "featureProcessMR failed to create merge request")
	}

	return nil
}

const _printTpl = `
	😬 Title: %s
	😬 URL	: %s
`

// printAndOpenBrowser print WebURL into stdout and open web browser.
func (f flowImpl) printAndOpenBrowser(title, url string) {
	if len(title) == 0 && len(url) == 0 {
		log.Warn("could not execute printAndOpenBrowser with empty title and url")
		return
	}
	if !strings.HasPrefix(url, "http") {
		log.Warnf("invalid url format: %s", url)
		return
	}

	var (
		err1, err2 error
	)

	_, err1 = fmt.Fprint(os.Stdout, fmt.Sprintf(_printTpl, title, url))
	if f.ctx.Conf.OpenBrowser {
		err2 = pkg.OpenBrowser(url)
	}
	log.WithFields(log.Fields{
		"err1": err1,
		"err2": err2,
	}).Debugf("printAndOpenBrowser")
}
