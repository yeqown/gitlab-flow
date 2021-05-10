package internal

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	gitop "github.com/yeqown/gitlab-flow/internal/git-operator"
	"github.com/yeqown/gitlab-flow/internal/repository"
	"github.com/yeqown/gitlab-flow/internal/repository/impl"
	"github.com/yeqown/gitlab-flow/internal/types"
	"github.com/yeqown/gitlab-flow/pkg"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/yeqown/log"
)

type dashImpl struct {
	ctx         *types.FlowContext
	repo        repository.IFlowRepository
	gitOperator gitop.IGitOperator
}

func NewDash(ctx *types.FlowContext) IDash {
	if ctx == nil {
		log.Fatal("empty FlowContext initialized")
		panic("can not reach")
	}

	log.
		WithField("context", ctx).
		Debugf("constructing dash")

	dash := dashImpl{
		ctx:         ctx,
		repo:        impl.NewBasedSqlite3(impl.ConnectDB(ctx.ConfPath(), ctx.Conf.DebugMode)),
		gitOperator: gitop.NewBasedCmd(ctx.CWD),
	}

	// DONE(@yeqown): need load project info from local database.
	if err := dash.fillContextWithProject(); err != nil {
		log.
			Fatalf("could not locate project(%s): %v", ctx.ProjectName(), err)
	}

	return dash
}

var (
	detailTpl        *template.Template
	detailTplPattern = `
🚗 Project's Name	:		{{.projectTitle}} (ID:{{.projectID}})
🚕 Project's URL	:		{{.projectURL}}
🚌 Milestone Title	:		{{.milestoneTitle}} (ID:{{.milestoneID}})
🎯 Milestone Desc	:		{{.milestoneDesc}}
🤡 Feature Branch	:		{{.featureBranch}}
👽 Milestone URL	:		{{.milestoneWebURL}}
`
	_featureDetailTblHeader      = []string{"MR#Src", "MR#Target", "MR#WebURL", "Issue#IID", "Issue#Desc"}
	_featureDetailIssueTblHeader = []string{"Issue#IID", "Issue#Title", "Issue#Desc", "Issue#WebURL"}
)

func init() {
	detailTpl = template.Must(
		template.New("detail").Parse(detailTplPattern))
}

func (d dashImpl) fillContextWithProject() error {
	// DONE(@yeqown): fill project information from local repository or remote gitlab repository.
	// DONE(@yeqown): projectName would be different from project path, use git repository name as project name.
	projectName := d.ctx.ProjectName()

	// get from local
	projects, err := d.repo.QueryProjects(&repository.ProjectDO{ProjectName: projectName})
	if err == nil && len(projects) != 0 {
		// should let user to choose one
		matched, err := chooseOneProjectInteractively(projects)
		if err == nil {
			d.ctx.Project = &types.ProjectBasics{
				ID:     matched.ProjectID,
				Name:   matched.ProjectName,
				WebURL: matched.WebURL,
			}
			return nil
		}
	}

	log.
		WithFields(log.Fields{"project": projectName}).
		Fatalf("could not found from local: %v", err)
	return fmt.Errorf("could not match project(%s) from remote: %v", projectName, err)
}

// FeatureDetail get feature detail of current project:
// * basic information to current milestone.
// * all merge request and its related issue created in current milestone.
// * all issues created in current milestone with web url.
func (d dashImpl) FeatureDetail(featureBranchName string) ([]byte, error) {
	if featureBranchName == "" {
		featureBranchName, _ = d.gitOperator.CurrentBranch()
	}
	if featureBranchName == "" {
		return nil, errInvalidFeatureName
	}
	featureBranchName = genFeatureBranchName(featureBranchName)

	// locate branch
	bm, err := d.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  d.ctx.Project.ID,
		BranchName: featureBranchName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not locate branch")
	}

	// query milestone
	milestone, err := d.repo.QueryMilestone(&repository.MilestoneDO{
		ProjectID:   d.ctx.Project.ID,
		MilestoneID: bm.MilestoneID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "dashImpl.FeatureDetail query milestone")
	}

	// query all merge requests related to milestone.
	mrs, err := d.repo.QueryMergeRequests(&repository.MergeRequestDO{
		ProjectID:   d.ctx.Project.ID,
		MilestoneID: bm.MilestoneID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "dashImpl.FeatureDetail query mergeRequest")
	}

	// query all issues related to milestone.
	issues, err := d.repo.QueryIssues(&repository.IssueDO{
		ProjectID:   d.ctx.Project.ID,
		MilestoneID: bm.MilestoneID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "dashImpl.FeatureDetail query issues")
	}

	// rework issue
	issueCache := make(map[int]*repository.IssueDO)
	issueTblData := make([][]string, len(issues))
	for idx, v := range issues {
		issueCache[v.IssueIID] = v
		issueTblData[idx] = []string{
			strconv.Itoa(v.IssueIID),
			v.Title,
			v.Desc,
			v.WebURL,
		}
	}

	mrTblData := make([][]string, len(mrs))
	for idx, mr := range mrs {
		issue, ok := issueCache[mr.IssueIID]
		if !ok {
			log.
				WithFields(log.Fields{
					"mergeRequestURL": mr.WebURL,
				}).
				Warn("no issues found with merge request")
			issue = new(repository.IssueDO)
		}

		mrTblData[idx] = []string{
			mr.SourceBranch,              //
			mr.TargetBranch,              //
			mr.WebURL,                    // MR-URL
			strconv.Itoa(issue.IssueIID), // issue.issueIID
			issue.Desc,                   // issue.Title
		}
	}

	log.
		WithFields(log.Fields{"mrTblData": mrTblData}).
		Debug("mrTblData is conducted")

	data := map[string]interface{}{
		"projectTitle":    d.ctx.Project.Name,
		"projectID":       d.ctx.Project.ID,
		"projectURL":      d.ctx.Project.WebURL,
		"milestoneTitle":  milestone.Title,
		"milestoneID":     milestone.MilestoneID,
		"milestoneDesc":   milestone.Desc,
		"featureBranch":   bm.BranchName,
		"milestoneWebURL": milestone.WebURL,
	}
	buf := bytes.NewBuffer(nil)
	if err = detailTpl.Execute(buf, data); err != nil {
		return nil, errors.Wrap(err, "detailTpl.Execute")
	}

	buf.WriteString("All Merge Requests:\n")

	// output all merge request into table
	w := tablewriter.NewWriter(buf)
	w.SetHeader(_featureDetailTblHeader)
	w.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	w.SetAlignment(tablewriter.ALIGN_LEFT)
	for _, row := range mrTblData {
		// if master merge request
		if row[1] == types.MasterBranch.String() {
			w.Rich(row, []tablewriter.Colors{
				{tablewriter.Bold, tablewriter.FgHiRedColor},
				{tablewriter.Bold, tablewriter.FgHiRedColor},
				{},
				{tablewriter.Bold, tablewriter.FgBlackColor},
			})
			continue
		}

		if row[1] == types.TestBranch.String() {
			w.Rich(row, []tablewriter.Colors{
				{tablewriter.Bold, tablewriter.FgHiGreenColor},
				{tablewriter.Bold, tablewriter.FgHiGreenColor},
				{},
				{tablewriter.Bold, tablewriter.FgBlackColor},
			})
			continue
		}

		w.Append(row)
	}
	w.Render()

	buf.WriteString("All Issues:\n")

	// output all issues into detail
	w2 := tablewriter.NewWriter(buf)
	w2.SetHeader(_featureDetailIssueTblHeader)
	w2.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	w2.SetAlignment(tablewriter.ALIGN_LEFT)
	w2.AppendBulk(issueTblData)
	w2.Render()

	return buf.Bytes(), nil
}

var (
	_milestoneOverviewTblHeader = []string{"🏝Project🏖", "🎠MergeRequests🏕"}
)

func (d dashImpl) MilestoneOverview(milestoneName, branchFilter string) ([]byte, error) {
	log.
		WithFields(log.Fields{
			"milestoneName":      milestoneName,
			"targetBranchFilter": branchFilter,
		}).
		Debug("MilestoneOverview called")

	if milestoneName == "" {
		// get milestoneName from feature branchName
		// query milestone name automatically when no milestone name provided.
		// TODO(@yeqown): optimize this logic, using repo method.
		featureBranchName, _ := d.gitOperator.CurrentBranch()
		featureBranchName = genFeatureBranchName(featureBranchName)
		branch, err := d.repo.QueryBranch(&repository.BranchDO{
			ProjectID:  d.ctx.Project.ID,
			BranchName: featureBranchName,
		})
		if err == nil && branch != nil {
			milestone, _ := d.repo.QueryMilestone(&repository.MilestoneDO{MilestoneID: branch.MilestoneID})
			if milestone != nil {
				milestoneName = milestone.Title
			}
		}
		log.
			WithFields(log.Fields{
				"featureBranchName": featureBranchName,
			}).
			Debugf("could not locate branch of current branch: %v", err)
	}

	// query milestone by name (milestone-project), so there are many.
	projectMilestones, err := d.repo.QueryMilestones(&repository.MilestoneDO{
		Title: milestoneName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not got any projectMilestones")
	}

	log.
		WithFields(log.Fields{
			"count":             len(projectMilestones),
			"projectMilestones": projectMilestones,
		}).
		Debug("catching milestone result")

	// handle data and format output
	tblData := make([][]string, 0, 8)
	for _, v := range projectMilestones {
		// locate project info
		project, err := d.repo.QueryProject(&repository.ProjectDO{ProjectID: v.ProjectID})
		if err != nil {
			log.
				WithFields(log.Fields{
					"projectId": v.ProjectID,
				}).
				Warnf("could not locate project: %v", err)
		}

		// catching mergeRequest of each project
		mrs, err := d.repo.QueryMergeRequests(&repository.MergeRequestDO{
			ProjectID:    v.ProjectID,
			MilestoneID:  v.MilestoneID,
			TargetBranch: branchFilter,
		})
		if err != nil {
			log.
				WithFields(log.Fields{
					"projectId":    v.ProjectID,
					"milestoneID":  v.MilestoneID,
					"targetBranch": branchFilter,
				}).
				Warnf("could not locate project merge request: %v", err)
		}

		uris := make([]string, 0, len(mrs))
		for _, mr := range mrs {
			uris = append(uris, fmt.Sprintf("%s➡️%s	🧲%s", mr.SourceBranch, mr.TargetBranch, mr.WebURL))
		}
		tblData = append(tblData, []string{project.ProjectName, strings.Join(uris, "\n")})
	}

	buf := bytes.NewBuffer(nil)
	w := tablewriter.NewWriter(buf)
	w.SetHeader(_milestoneOverviewTblHeader)
	w.AppendBulk(tblData)
	w.Render()

	return buf.Bytes(), nil
}

func (d dashImpl) ProjectDetail() ([]byte, error) {
	d.printAndOpenBrowser(d.ctx.Project.Name, d.ctx.Project.WebURL)

	return nil, nil
}

// printAndOpenBrowser print WebURL into stdout and open web browser.
func (d dashImpl) printAndOpenBrowser(title, url string) {
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
	if d.ctx.Conf.OpenBrowser {
		err2 = pkg.OpenBrowser(url)
	}
	log.WithFields(log.Fields{
		"err1": err1,
		"err2": err2,
	}).Debugf("printAndOpenBrowser")
}
