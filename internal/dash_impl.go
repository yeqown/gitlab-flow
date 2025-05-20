package internal

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/yeqown/log"

	gitop "github.com/yeqown/gitlab-flow/internal/git-operator"
	"github.com/yeqown/gitlab-flow/internal/repository"
	"github.com/yeqown/gitlab-flow/internal/repository/impl"
	"github.com/yeqown/gitlab-flow/internal/types"
	"github.com/yeqown/gitlab-flow/pkg"
)

type dashImpl struct {
	ctx         *types.FlowContext
	repo        repository.IFlowRepository
	gitOperator gitop.IGitOperator
}

func NewDash(ctx *types.FlowContext, ch IConfigHelper) IDash {
	if ctx == nil {
		log.Fatal("empty FlowContext initialized")
		panic("can not reach")
	}

	log.
		WithField("context", ctx).
		Debugf("constructing dash")

	dash := dashImpl{
		ctx:         ctx,
		repo:        impl.NewBasedSqlite3(impl.ConnectDB(ch.Context().GlobalConfPath, ctx.IsDebug())),
		gitOperator: gitop.NewBasedCmd(ctx.CWD()),
	}

	// DONE(@yeqown): need load project info from a local database.
	if err := dash.fillContextWithProject(); err != nil {
		log.Fatalf("could not locate project(%s): %v", ctx.ProjectName(), err)
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

// fillContextWithProject
// DONE(@yeqown): fill project information from local repository or remote gitlab repository.
// DONE(@yeqown): projectName would be different from project path, use git repository name as project name.
func (d dashImpl) fillContextWithProject() error {
	projectName := d.ctx.ProjectName()

	// get from local
	injected, err := injectProjectIntoContext(d.repo, d.ctx, projectName, d.ctx.CWD())
	if err == nil && injected {
		return nil
	}

	// err != nil or not injected
	log.
		WithFields(log.Fields{"cwd": d.ctx.CWD(), "injected": injected}).
		Fatalf("could not found project(%s) from local: %v", projectName, err)

	return fmt.Errorf("could not found project(%s) from local: %v", projectName, err)
}

// FeatureDetail get feature detail of current project:
// * basic information to current milestone.
// * all merge request and its related issue created in current milestone.
// * all issues created in current milestone with web url.
func (d dashImpl) FeatureDetail(branchName string) ([]byte, error) {
	if branchName == "" {
		out, err := d.gitOperator.CurrentBranch()
		if out == "" || err != nil {
			log.
				WithFields(log.Fields{
					"branch": branchName,
					"err":    err,
				}).
				Debug("dashImpl.FeatureDetail get branch name failed")
			return nil, errors.Wrap(errInvalidFeatureName, "dashImpl.FeatureDetail.CurrentBranch")
		}
		branchName = out

		// if current branch name could be parsed to feature branch name, then use it.
		if !isFeatureName(branchName) {
			out, ok := tryParseFeatureNameFrom(branchName, false)
			if !ok {
				log.
					WithFields(log.Fields{
						"branchName": branchName,
						"out":        out,
						"ok":         ok,
					}).
					Debug("dashImpl.FeatureDetail could not parse branch name by default")
				return nil, errors.Wrap(errInvalidFeatureName, "dashImpl.FeatureDetail.tryParseFeatureNameFrom")
			}
			branchName = out
		}
	}

	branchName = genFeatureBranchName(branchName)
	// locate branch
	branch, err := d.repo.QueryBranch(&repository.BranchDO{
		ProjectID:  d.ctx.Project().ID,
		BranchName: branchName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not locate branch:"+branchName)
	}

	// query milestone
	milestone, err := d.repo.QueryMilestone(&repository.MilestoneDO{
		ProjectID:   d.ctx.Project().ID,
		MilestoneID: branch.MilestoneID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "dashImpl.FeatureDetail query milestone")
	}

	// query all merge requests related to milestone.
	mrs, err := d.repo.QueryMergeRequests(&repository.MergeRequestDO{
		ProjectID:   d.ctx.Project().ID,
		MilestoneID: branch.MilestoneID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "dashImpl.FeatureDetail query mergeRequest")
	}

	// query all issues related to milestone.
	issues, err := d.repo.QueryIssues(&repository.IssueDO{
		ProjectID:   d.ctx.Project().ID,
		MilestoneID: branch.MilestoneID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "dashImpl.FeatureDetail query issues")
	}

	return d.dealDataIntoFeatureDetail(branch, milestone, issues, mrs)
}

// dealDataIntoFeatureDetail deal all data related to feature branch into template and tables.
func (d dashImpl) dealDataIntoFeatureDetail(
	bm *repository.BranchDO, milestone *repository.MilestoneDO,
	issues []*repository.IssueDO, mrs []*repository.MergeRequestDO,
) ([]byte, error) {
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
		"projectTitle":    d.ctx.Project().Name,
		"projectID":       d.ctx.Project().ID,
		"projectURL":      d.ctx.Project().WebURL,
		"milestoneTitle":  milestone.Title,
		"milestoneID":     milestone.MilestoneID,
		"milestoneDesc":   milestone.Desc,
		"featureBranch":   bm.BranchName,
		"milestoneWebURL": milestone.WebURL,
	}
	buf := bytes.NewBuffer(nil)
	if err := detailTpl.Execute(buf, data); err != nil {
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
	_milestoneOverviewTblHeader = []string{"🏝Project", "MR#Action", "🏕MR#WebURL"}
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
		featureBranchName, _ := d.gitOperator.CurrentBranch()
		featureBranchName = genFeatureBranchName(featureBranchName)
		// DONE(@yeqown): optimize this logic, using repo method.
		milestone, err := d.repo.QueryMilestoneByBranchName(d.ctx.Project().ID, featureBranchName)
		if milestone != nil {
			milestoneName = milestone.Title
		}

		log.
			WithFields(log.Fields{
				"featureBranchName": featureBranchName,
				"error":             err,
			}).
			Debugf("locate milestone of current branch")
	}

	if milestoneName == "" {
		return nil, errors.New("you must specify a milestone name or " +
			"sure you are using a branch which could get milestone")
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

		// insert all merge requests of current project into tblData
		for _, mr := range mrs {
			tblData = append(tblData, []string{
				project.ProjectName,
				fmt.Sprintf("%s => %s", mr.SourceBranch, mr.TargetBranch),
				mr.WebURL,
			})
		}
	}

	buf := bytes.NewBuffer(nil)
	w := tablewriter.NewWriter(buf)
	w.SetHeader(_milestoneOverviewTblHeader)
	w.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		tablewriter.Colors{},
		tablewriter.Colors{},
	)
	w.SetRowLine(true)
	w.SetAutoMergeCells(true)
	w.AppendBulk(tblData)
	w.Render()

	return buf.Bytes(), nil
}

func (d dashImpl) ProjectDetail(module string) ([]byte, error) {
	switch module {
	case "home":
		d.printAndOpenBrowser(d.ctx.Project().Name, d.ctx.Project().WebURL)
	case "branch":
		d.printAndOpenBrowser("branches", genProjectURL(d.ctx.Project().WebURL, "/-/branches"))
	case "tag":
		d.printAndOpenBrowser("tags", genProjectURL(d.ctx.Project().WebURL, "/-/tags"))
	case "commit":
		d.printAndOpenBrowser("commits", genProjectURL(d.ctx.Project().WebURL, "/commits/master"))
	}

	return nil, nil
}

func genProjectURL(base, suffix string) string {
	return base + suffix
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

	_, err1 = fmt.Fprintf(os.Stdout, _printTpl, title, url)
	if d.ctx.ShouldOpenBrowser() {
		err2 = pkg.OpenBrowser(url)
	}
	log.WithFields(log.Fields{
		"err1": err1,
		"err2": err2,
	}).Debugf("printAndOpenBrowser")
}

type projectQueryHelper interface {
	QueryProjects(*repository.ProjectDO) ([]*repository.ProjectDO, error)
}

type projectInjectIntoContextHelper interface {
	InjectProject(*types.ProjectBasics)
}

// injectProjectIntoContext query project from local persistence repository and inject into context.
//
// First, query project from local repository with projectName,
// if not found, then query with localDir.
//
// Then, let user choose one project from matched projects if there are more than one.
//
// Finally, inject project into context.
func injectProjectIntoContext(
	q projectQueryHelper, injector projectInjectIntoContextHelper, projectName, localDir string) (bool, error) {
	filter := &repository.ProjectDO{ProjectName: projectName}
	tries := 1

query:
	// if tries > 2, means we have tried twice, tried with projectName and localDir.
	// but still not found, then return false.
	if tries > 2 {
		return false, nil
	}

	log.WithFields(log.Fields{"filter": filter, "tries": tries}).
		Debug("injectProjectIntoContext.query")
	projects, err := q.QueryProjects(filter)
	log.WithFields(log.Fields{"projects": projects, "got": len(projects)}).
		Debugf("injectProjectIntoContext.query result, err=%v", err)
	if err != nil {
		return false, errors.Wrap(err, "injectProjectIntoContext.QueryProjects")
	}
	if len(projects) == 0 {
		// try with localDir
		tries++
		filter = &repository.ProjectDO{LocalDir: localDir}
		goto query
	}

	if len(projects) != 0 {
		// let user choose one
		matched, err2 := chooseOneProjectInteractively(projects)
		if err2 != nil {
			return true, errors.Wrap(err2, "injectProjectIntoContext.chooseOneProjectInteractively")
		}

		injector.InjectProject(&types.ProjectBasics{
			ID:     matched.ProjectID,
			Name:   matched.ProjectName,
			WebURL: matched.WebURL,
		})
	}

	return true, nil
}
