package internal

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/yeqown/gitlab-flow/internal/conf"
	gitop "github.com/yeqown/gitlab-flow/internal/git-operator"
	gitlabop "github.com/yeqown/gitlab-flow/internal/gitlab-operator"
	"github.com/yeqown/gitlab-flow/internal/repository"
	"github.com/yeqown/gitlab-flow/internal/repository/impl"
	"github.com/yeqown/gitlab-flow/internal/types"
	"github.com/yeqown/gitlab-flow/pkg"

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

var (
	errInvalidFeatureName = errors.New("feature branch could not be empty")
)

// checkOAuthAccessToken check access token is valid or not. If access token becomes invalid
// then refresh it, if refresh failed, it leads to  re-authorize.
func checkOAuthAccessToken(ctx *types.FlowContext) {
	c, _ := conf.Load(ctx.ConfPath(), nil)
	oauth := gitlabop.NewOAuth2Support(&gitlabop.OAuth2Config{
		Host:         c.GitlabHost,
		ServeAddr:    "", // use default address
		AccessToken:  c.OAuth.AccessToken,
		RefreshToken: c.OAuth.RefreshToken,
	})
	if err := oauth.Enter(c.OAuth.RefreshToken); err != nil {
		log.
			WithFields(log.Fields{
				"config": c,
			}).
			Errorf("NewGitlabOperator could not renew token: %v", err)
		panic(err)
	}

	accessToken, refreshToken := oauth.Load()

	c.OAuth.AccessToken = accessToken
	c.OAuth.RefreshToken = refreshToken
	// update context oauth configuration
	ctx.GetOAuth().AccessToken = accessToken
	ctx.GetOAuth().RefreshToken = refreshToken

	if err := conf.Save(ctx.ConfPath(), c, nil); err != nil {
		log.
			WithFields(log.Fields{
				"config": c,
				"error":  err,
			}).
			Error("checkOAuthAccessToken failed to save")
	}
}

func NewFlow(ctx *types.FlowContext) IFlow {
	if ctx == nil {
		log.Fatal("empty FlowContext initialized")
		panic("can not reach")
	}

	log.
		WithField("context", ctx).
		Debugf("constructing flow")

	checkOAuthAccessToken(ctx)

	flow := &flowImpl{
		ctx:            ctx,
		gitlabOperator: gitlabop.NewGitlabOperator(ctx.GetOAuth().AccessToken, ctx.APIEndpoint()),
		gitOperator:    gitop.NewBasedCmd(ctx.CWD()),
		repo:           impl.NewBasedSqlite3(impl.ConnectDB(ctx.ConfPath(), ctx.IsDebug())),
	}

	// if flowContext has NONE project information, so we need to fill it.
	if err := flow.fillContextWithProject(); err != nil {
		log.
			Fatalf("could not locate project(%s): %v", ctx.ProjectName(), err)
	}

	return flow
}

// fillContextWithProject .
// FlowContext with null project information, so we need to fill it.
func (f flowImpl) fillContextWithProject() error {
	// DONE(@yeqown): fill project information from local repository or remote gitlab repository.
	// DONE(@yeqown): projectName would be different from project path, use git repository name as project name.
	var (
		projectName = f.ctx.ProjectName()
		projects    []*repository.ProjectDO
		err         error
	)

	// if user specify locating project from remote directly, so skip the step of getting from local.
	if f.ctx.ForceRemote() {
		goto locateFromRemote
	}

	// get from local
	projects, err = f.repo.QueryProjects(&repository.ProjectDO{ProjectName: projectName})
	if err == nil && len(projects) != 0 {
		// locate project from local, and there are maybe more than one project.
		matched, err2 := chooseOneProjectInteractively(projects)
		if err2 == nil {
			f.ctx.InjectProject(&types.ProjectBasics{
				ID:     matched.ProjectID,
				Name:   matched.ProjectName,
				WebURL: matched.WebURL,
			})
			return nil
		}
	}

	log.
		WithFields(log.Fields{"project": projectName}).
		Warnf("could not found from local: %v", err)

locateFromRemote:
	// query from remote repository.
	result, err := f.gitlabOperator.ListProjects(context.Background(), &gitlabop.ListProjectRequest{
		Page:        1,
		PerPage:     20,
		ProjectName: projectName,
	})
	if err != nil {
		return errors.Wrap(err, "requests remote repository failed")
	}

	// found and match
	// DONE(@yeqown): if remote(gitlab) has not only one project with projectName, then choose one as target.
	remoteMatched := make([]*repository.ProjectDO, 0, 5)
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

			// DONE(@yeqown): save into local database
			projectDO := repository.ProjectDO{
				ProjectName: projectName,
				ProjectID:   v.ID,
				LocalDir:    f.ctx.CWD(),
				WebURL:      v.WebURL,
			}
			remoteMatched = append(remoteMatched, &projectDO)
			continue
		}
	}

	matched, err := chooseOneProjectInteractively(remoteMatched)
	if err == nil {
		if err = f.repo.SaveProject(matched); err != nil {
			log.
				WithField("project", matched).
				Warn("could not save project")
		}

		f.ctx.InjectProject(&types.ProjectBasics{
			ID:     matched.ProjectID,
			Name:   matched.ProjectName,
			WebURL: matched.WebURL,
		})
		return nil
	}

	// could not match
	return fmt.Errorf("could not match project(%s) from remote: %v", projectName, err)
}

func (f flowImpl) FeatureBegin(opc *types.OpFeatureContext, title, desc string) error {
	log.
		WithFields(log.Fields{
			"title": title,
			"desc":  desc,
		}).
		Debug("FeatureBegin called")

	if err := blockingNamePrefix(title); err != nil {
		return errors.Wrap(err, "blocking name prefix detected")
	}

	// create milestone
	result, err := f.createMilestone(title, desc)
	if err != nil {
		return errors.Wrap(err, "CreateMilestone failed")
	}

	// create feature branch, branchName is generated from title "feature/title" as default
	featureBranchName := genFeatureBranchName(title)
	if len(opc.FeatureBranchName) != 0 {
		featureBranchName = genFeatureBranchName(opc.FeatureBranchName)
	}
	_, err = f.createBranch(featureBranchName, types.MasterBranch.String(), result.ID, 0)
	if err != nil {
		return err
	}

	return nil
}

// extractFeatureBranchName return feature branch name with rule and input.
// input:	opc.FeatureBranchName = "aaaa"
// output	"feature/aaaa"
// if input is empty string, flowImpl would use the project's current git branch name,
// error would be returned when no feature branch is found.
func (f flowImpl) extractFeatureBranchName(opc *types.OpFeatureContext) (string, error) {
	if opc.FeatureBranchName == "" {
		opc.FeatureBranchName, _ = f.gitOperator.CurrentBranch()
	}
	if opc.FeatureBranchName == "" {
		return "", errInvalidFeatureName
	}
	opc.FeatureBranchName = genFeatureBranchName(opc.FeatureBranchName)
	return opc.FeatureBranchName, nil
}

func (f flowImpl) FeatureDebugging(opc *types.OpFeatureContext) (err error) {
	if opc.FeatureBranchName, err = f.extractFeatureBranchName(opc); err != nil {
		return err
	}
	return f.featureProcessMR(opc.FeatureBranchName, types.DevBranch, opc.ForceCreateMergeRequest)
}

func (f flowImpl) FeatureTest(opc *types.OpFeatureContext) (err error) {
	if opc.FeatureBranchName, err = f.extractFeatureBranchName(opc); err != nil {
		return err
	}
	return f.featureProcessMR(opc.FeatureBranchName, types.TestBranch, opc.ForceCreateMergeRequest)
}

func (f flowImpl) FeatureRelease(opc *types.OpFeatureContext) (err error) {
	if opc.FeatureBranchName, err = f.extractFeatureBranchName(opc); err != nil {
		return err
	}
	return f.featureProcessMR(opc.FeatureBranchName, types.MasterBranch, opc.ForceCreateMergeRequest)
}

func (f flowImpl) FeatureResolveConflict(opc *types.OpFeatureContext, targetBranch types.BranchTyp) (err error) {
	if opc.FeatureBranchName, err = f.extractFeatureBranchName(opc); err != nil {
		return err
	}

	// feature/branch-1 => conflict-resolve/branch-1
	resolveConflictBranch := strings.Replace(opc.FeatureBranchName, FeatureBranchPrefix, ConflictResolveBranchPrefix, 1)

	// locate feature branch
	featureBranch, err := f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  f.ctx.Project().ID,
		BranchName: opc.FeatureBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate feature branch failed")
	}

	// create resolve conflict branch.
	// Notice: createBranch would create branch which checkout to new branch automatically.
	if _, err = f.createBranch(
		resolveConflictBranch, targetBranch.String(), featureBranch.MilestoneID, 0); err != nil {
		return err
	}

	// check out one branch from target branch, and open a MergeRequest.
	if err = f.featureProcessMR(resolveConflictBranch, targetBranch, opc.ForceCreateMergeRequest); err != nil {
		return err
	}

	// then local git command to use git merge
	//  --no-ff `featureBranchName`
	return f.gitOperator.Merge(opc.FeatureBranchName, resolveConflictBranch)
}

func (f flowImpl) FeatureBeginIssue(opc *types.OpFeatureContext, title, desc string) error {
	// DONE(@yeqown): is featureBranchName is empty, use current branch name.
	if opc.FeatureBranchName == "" {
		opc.FeatureBranchName, _ = f.gitOperator.CurrentBranch()
	}
	if opc.FeatureBranchName == "" {
		return errInvalidFeatureName
	}
	opc.FeatureBranchName = genFeatureBranchName(opc.FeatureBranchName)

	featureBranch, err := f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  f.ctx.Project().ID,
		BranchName: opc.FeatureBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate feature branch from local failed")
	}

	// query milestone
	milestone, err := f.repo.QueryMilestone(&repository.MilestoneDO{
		ProjectID:   f.ctx.Project().ID,
		MilestoneID: featureBranch.MilestoneID,
	})
	if err != nil {
		return errors.Wrap(err, "locate milestone failed")
	}

	if len(title) == 0 {
		title = milestone.Title
	} else {
		if err := blockingNamePrefix(title); err != nil {
			return errors.Wrap(err, "blocking name prefix detected")
		}
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
	_, err = f.createBranch(issueBranchName, featureBranch.BranchName, milestone.MilestoneID, issue.IID)
	if err != nil {
		return errors.Wrap(err, "create branch failed")
	}

	f.printAndOpenBrowser("Open Issue", issue.WebURL)

	return nil
}

// FeatureFinishIssue implements IFlow.FeatureFinishIssue
// DONE(@yeqown): issue merge request should be called here, rather than FeatureBeginIssue
func (f flowImpl) FeatureFinishIssue(opc *types.OpFeatureContext, issueBranchName string) (err error) {
	// DONE(@yeqown): if issueBranchName is empty, make current branch name as default.
	if issueBranchName == "" {
		issueBranchName, _ = f.gitOperator.CurrentBranch()
	}
	if issueBranchName == "" {
		return errors.New("issue branch could not be empty")
	}

	var (
		milestoneID = 0
		issueIID    = 0
	)
	// locate issue branch.
	if b, err := f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  f.ctx.Project().ID,
		BranchName: issueBranchName,
	}); err != nil {
		return errors.Wrapf(err, "locate issue branch(%s) failed", issueBranchName)
	} else {
		milestoneID = b.MilestoneID
		issueIID = b.IssueIID
	}

	// DONE(@yeqown) get feature branch name from issueBranchName
	if opc.FeatureBranchName == "" {
		opc.FeatureBranchName = parseFeatureFromIssueName(issueBranchName, opc.ParseIssueCompatible)
	}
	if opc.FeatureBranchName == "" {
		return errInvalidFeatureName
	}
	opc.FeatureBranchName = genFeatureBranchName(opc.FeatureBranchName)
	//if _, err := f.repo.QueryBranch(&repository.BranchDO{
	//	ProjectID:   f.ctx.Project().ID,
	//	BranchName:  featureBranchName,
	//	MilestoneID: milestoneID,
	//}); err != nil {
	//	return errors.Wrapf(err, "locate feature branch(%s) failed", featureBranchName)
	//}

	var mr *repository.MergeRequestDO
	if opc.ForceCreateMergeRequest {
		goto issueCreateMR
	}

	// locate MR
	mr, err = f.repo.QueryMergeRequest(&repository.MergeRequestDO{
		ProjectID:    f.ctx.Project().ID,
		IssueIID:     issueIID,
		MilestoneID:  milestoneID,
		SourceBranch: issueBranchName,
		TargetBranch: opc.FeatureBranchName,
	})
	if err != nil && !repository.IsErrNotFound(err) {
		log.
			WithFields(log.Fields{
				"projectID":    f.ctx.Project().ID,
				"issueIID":     issueIID,
				"milestoneID":  milestoneID,
				"sourceBranch": issueBranchName,
				"targetBranch": opc.FeatureBranchName,
			}).
			Errorf("locate MR failed: %v", err)
		return errors.Wrap(err, "locate MR failed")
	}

	// got merge request from local
	if mr != nil {
		log.
			WithFields(log.Fields{
				"featureBranch":   opc.FeatureBranchName,
				"issueBranch":     issueBranchName,
				"mergeRequestURL": mr.WebURL,
			}).
			Debug("issue info")

		f.printAndOpenBrowser("Issue Merge Request", mr.WebURL)
		return nil
	}
issueCreateMR:
	// not hit, so create one
	title := genMergeRequestName(issueBranchName, opc.FeatureBranchName)
	desc := ""
	result, err := f.createMergeRequest(title, desc, milestoneID, issueIID, issueBranchName, opc.FeatureBranchName)
	if err != nil {
		return errors.Wrap(err, "create issue merge request failed")
	}

	log.
		WithFields(log.Fields{
			"issueBranchName":   issueBranchName,
			"featureBranchName": opc.FeatureBranchName,
			"mergeRequestURL":   result.WebURL,
		}).
		Debug("create issue merge request finished")

	f.printAndOpenBrowser("Issue Merge Request", result.WebURL)

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
		ProjectID:  f.ctx.Project().ID,
		BranchName: hotfixBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate hotfix branch failed")
	}

	// locate issue
	issue, err := f.repo.QueryIssue(&repository.IssueDO{
		ProjectID:     f.ctx.Project().ID,
		RelatedBranch: hotfixBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate issue failed")
	}

	// locate MR first
	mr, err := f.repo.QueryMergeRequest(&repository.MergeRequestDO{
		ProjectID:   f.ctx.Project().ID,
		MilestoneID: issue.MilestoneID,
		IssueIID:    issue.IssueIID,
		//SourceBranch:   "",
		TargetBranch: types.MasterBranch.String(),
	})
	if err != nil && !repository.IsErrNotFound(err) {
		return errors.Wrap(err, "query database failed")
	}

	// hit hotfix merge request
	if mr != nil {
		f.printAndOpenBrowser("Hotfix Merge Request", mr.WebURL)
		return nil
	}

	// then create MR to master
	title := genMergeRequestName(hotfixBranchName, types.MasterBranch.String())
	result, err := f.createMergeRequest(
		title, issue.Desc, 0, issue.IssueIID, hotfixBranchName, types.MasterBranch.String())
	if err != nil {
		return errors.Wrap(err, "create hotfix MR failed")
	}

	f.printAndOpenBrowser("Hotfix Merge Request", result.WebURL)

	log.
		WithFields(log.Fields{
			"issue":        issue,
			"mergeRequest": result,
		}).
		Debug("hotfix finish done")

	return nil
}

// SyncMilestone rebuilds local data related to `milestoneID`
//
// 1. pull milestone + MergeRequest + Issues by `milestoneID`.
// 2. parse `IssueID` from MR description.
// 3. handle and save data.
func (f flowImpl) SyncMilestone(milestoneID int, interact bool) error {
	ctx := context.Background()
	projectId := f.ctx.Project().ID

	// parameter checking
	if milestoneID == 0 && !interact {
		return errors.New("milestoneID could not be zero")
	}

	// interact mode
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

		milestones := make([]*repository.MilestoneDO, len(result.Data))
		for idx, v := range result.Data {
			milestones[idx] = &repository.MilestoneDO{
				ProjectID:   projectId,
				MilestoneID: v.ID,
				Title:       v.Name,
				Desc:        v.Description,
				WebURL:      v.WebURL,
			}
		}

		milestone, err := chooseOneMilestoneInteractively(milestones)
		if err != nil {
			return errors.Wrap(err, "chooseOneMilestoneInteractively failed")
		}
		milestoneID = milestone.MilestoneID
	}

	log.Info("Querying remote repository data")
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

	// format data into DO
	i, mr, b, branchName := f.
		syncFormatResultIntoDO(milestoneResult, milestoneMRsResult.Data, milestoneIssuesResult.Data)
	log.WithFields(log.Fields{
		"milestoneResult":       milestoneResult,
		"milestoneMRsResult":    milestoneMRsResult,
		"milestoneIssuesResult": milestoneIssuesResult,
		"featureBranchName":     branchName,
	}).Debugf("syncFormatResultIntoDO calling")

	// save data models
	log.Info("Saving remote repository data into local database...")
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

	log.Info("Fetching remote branches...")
	_ = f.gitOperator.FetchOrigin()

	return nil
}

// SyncProject want to rebuild project manually while project is not exists in local database.
func (f flowImpl) SyncProject() error {
	// NewFlow get project information earlier than SyncProject method, so synchronize project only need to
	// set FlowContext.forceRemote as true which will read project information from remote rather than
	// load from local database.

	// forceRemote was set by `parseGlobalFlags` function.
	// lookup parseGlobalFlags for more detail.

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
		issueDO    = make([]*repository.IssueDO, 0, 10)
		mrDO       = make([]*repository.MergeRequestDO, 0, 10)
		branchDO   = make([]*repository.BranchDO, 0, 10)
		branchUniq = make(map[string]struct{})

		c                 = make(map[int]*repository.IssueDO)
		featureBranchName string
		projectID         = f.ctx.Project().ID
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
		log.
			WithFields(log.Fields{
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

		// featureBranchName
		if featureBranchName == "" && strings.HasPrefix(mr.SourceBranch, FeatureBranchPrefix) {
			featureBranchName = mr.SourceBranch
		}
		if featureBranchName == "" && strings.HasPrefix(mr.TargetBranch, FeatureBranchPrefix) {
			featureBranchName = mr.TargetBranch
		}

		if _, ok := branchUniq[mr.SourceBranch]; !ok {
			branchDO = append(branchDO, &repository.BranchDO{
				ProjectID:   projectID,
				MilestoneID: milestoneID,
				IssueIID:    issueIID,
				BranchName:  mr.SourceBranch,
			})
			branchUniq[mr.SourceBranch] = struct{}{}
		}
		// targetBranch should synchronize too.
		if _, ok := branchUniq[mr.TargetBranch]; !ok && notBuiltinBranch(mr.TargetBranch) {
			branchDO = append(branchDO, &repository.BranchDO{
				ProjectID:   projectID,
				MilestoneID: milestoneID,
				IssueIID:    issueIID,
				BranchName:  mr.TargetBranch,
			})
			branchUniq[mr.TargetBranch] = struct{}{}
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
		ProjectID:    f.ctx.Project().ID,
		//MilestoneID:  milestoneID,
		//IssueID:      issueIID,
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
			ProjectID:   f.ctx.Project().ID,
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
	title = strings.TrimSpace(title)
	desc = strings.TrimSpace(desc)

	ctx := context.Background()
	result, err := f.gitlabOperator.CreateMilestone(ctx, &gitlabop.CreateMilestoneRequest{
		Title:     title,
		Desc:      desc,
		ProjectID: f.ctx.Project().ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "CreateMilestone failed")
	}

	if err = f.repo.SaveMilestone(&repository.MilestoneDO{
		ProjectID:   f.ctx.Project().ID,
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
	title = strings.TrimSpace(title)
	desc = strings.TrimSpace(desc)

	ctx := context.Background()
	result, err := f.gitlabOperator.CreateIssue(ctx, &gitlabop.CreateIssueRequest{
		Title:         title,
		Desc:          desc,
		RelatedBranch: relatedBranch,
		MilestoneID:   milestoneID,
		ProjectID:     f.ctx.Project().ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create Issue failed")
	}

	if err = f.repo.SaveIssue(&repository.IssueDO{
		IssueIID:      result.IID,
		Title:         title,
		Desc:          desc,
		ProjectID:     f.ctx.Project().ID,
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
		ProjectID:    f.ctx.Project().ID,
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
		ProjectID:      f.ctx.Project().ID,
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

// featureProcessMR is a process for creating a merge request for feature branch to target branch. If
// forceCreateMR is true means skipping the logic which would query MergeRequest from local.
func (f flowImpl) featureProcessMR(
	featureBranchName string, targetBranchName types.BranchTyp, forceCreateMR bool) error {
	featureBranch, err := f.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  f.ctx.Project().ID,
		BranchName: featureBranchName,
	})
	if err != nil {
		return errors.Wrap(err, "locate feature branch failed")
	}

	var mr *repository.MergeRequestDO

	if forceCreateMR {
		goto featureCreateMR
	}
	// query feature MR first
	mr, err = f.repo.QueryMergeRequest(&repository.MergeRequestDO{
		ProjectID:    f.ctx.Project().ID,
		MilestoneID:  featureBranch.MilestoneID,
		IssueIID:     featureBranch.IssueIID,
		SourceBranch: featureBranchName,
		TargetBranch: targetBranchName.String(),
	})
	if err != nil && !repository.IsErrNotFound(err) {
		return errors.Wrap(err, "query merge request failed")
	}
	if mr != nil {
		f.printAndOpenBrowser("Feature Merge Request", mr.WebURL)
		return nil
	}

featureCreateMR:
	milestone, err := f.repo.QueryMilestone(&repository.MilestoneDO{
		MilestoneID: featureBranch.MilestoneID,
	})
	if err != nil {
		return errors.Wrap(err, "locate milestone failed")
	}
	// create MR
	targetBranch := targetBranchName.String()
	title := genMergeRequestName(featureBranchName, targetBranch)
	result, err := f.createMergeRequest(
		title, milestone.Desc, milestone.MilestoneID, 0, featureBranch.BranchName, targetBranch)
	if err != nil {
		return errors.Wrapf(err, "featureProcessMR failed to create merge request")
	}

	f.printAndOpenBrowser("Feature Merge Request", result.WebURL)

	return nil
}

const _printTpl = `
	👽 Title: %s
	🤡 URL	: %s
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

	_, err1 = fmt.Fprintf(os.Stdout, _printTpl, title, url)
	if f.ctx.ShouldOpenBrowser() {
		err2 = pkg.OpenBrowser(url)
	}
	log.WithFields(log.Fields{
		"err1": err1,
		"err2": err2,
	}).Debugf("printAndOpenBrowser")
}
