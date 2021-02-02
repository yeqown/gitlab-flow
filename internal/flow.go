package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yeqown/gitlab-flow/internal/repository"
	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"github.com/yeqown/log"
)

// IFlow to control branches, MRs, milestones and issues.
type IFlow interface {
	// FeatureBegin open a milestone and related to a feature branch,
	// then CLI would automate fetch origin branches and pull them to local.
	// Of course, flow would save data in local storage.
	FeatureBegin(title, desc string) error
	// FeatureDebugging open a MergeRequest of feature branch and types.DevBranch branch.
	FeatureDebugging(featureBranchName string) error
	// FeatureTest open a MergeRequest of feature branch and types.TestBranch branch.
	FeatureTest(featureBranchName string) error
	// FeatureRelease open a MergeRequest of feature branch and types.MasterBranch branch.
	FeatureRelease(featureBranchName string) error
	// DONE(@yeqown) this would be useful while you merge feature into master but there is conflict.
	// FeatureResolveConflict will checkout a new branch from target branch,
	// then create a merge request from current feature branch to the new branch.
	// newBranch = "resolve-conflict/featureBranchName-to-master"
	FeatureResolveConflict(featureBranchName string, targetBranch types.BranchTyp) error

	// FeatureBeginIssue checkout a issue branch from feature branch, also open a merge request
	// which is from issue branch to feature branch.
	FeatureBeginIssue(featureBranchName string, title, desc string) error
	// FeatureFinishIssue open the WebURL of merge request which is from issue branch to feature branch.
	FeatureFinishIssue(featureBranchName, issueBranchName string) error

	// HotfixStart checkout a hotfix branch from types.MasterBranch, also open a merge request
	// which is from hotfix branch to types.MasterBranch.
	HotfixBegin(title, desc string) error
	// HotfixRelease open the WebURL of merge request which is from hotfix branch to types.MasterBranch.
	HotfixFinish(hotfixBranchName string) error

	// SyncMilestone synchronize remote repository milestone and
	// related issues / merge requests to local.
	SyncMilestone(milestoneID int, interact bool) error
}

var (
	_closePattern = "[Cc]loses? #(\\d+)"

	_closeReg *regexp.Regexp
)

func init() {
	_closeReg = regexp.MustCompile(_closePattern)
}

// parseIssueIIDFromMergeRequestIssue .
func parseIssueIIDFromMergeRequestIssue(desc string) (issueIID int) {
	data := _closeReg.FindSubmatch([]byte(desc))
	if len(data) == 0 {
		return
	}

	d, err := strconv.Atoi(string(data[1]))
	if err != nil {
		log.
			WithField("find", string(data[1])).
			Warnf("parse issue iid from desc: %v", err)
		return
	}

	return d
}

// notBuiltinBranch judge whether branchName is builtin or not.
// true means not builtin name, false is instead.
func notBuiltinBranch(branchName string) bool {
	switch branchName {
	case types.DevBranch.String(), types.TestBranch.String(), types.MasterBranch.String():
		return false
	}
	return true
}

const (
	FeatureBranchPrefix         = "feature/"
	HotfixBranchPrefix          = "hotfix/"
	ConflictResolveBranchPrefix = "conflict-resolve/"
)

// genFeatureBranchName
func genFeatureBranchName(name string) string {
	if strings.HasPrefix(name, FeatureBranchPrefix) {
		return name
	}

	return FeatureBranchPrefix + name
}

// genHotfixBranchName .
func genHotfixBranchName(name string) string {
	if strings.HasPrefix(name, HotfixBranchPrefix) {
		return name
	}

	return HotfixBranchPrefix + name
}

// genMRTitle
func genMRTitle(srcBranch, targetBranch string) string {
	return fmt.Sprintf("Merge %s to %s", srcBranch, targetBranch)
}

// genIssueBranchName .
// @result = 1-milestoneTitle as default
// fmt.Sprintf("%d-%s", issue.IID, milestone.Title)
func genIssueBranchName(name string, issueIID int) string {
	if strings.HasPrefix(name, strconv.Itoa(issueIID)+"-") {
		return name
	}

	return fmt.Sprintf("%d-%s", issueIID, name)
}

func parseFeaturenameFromIssueName(issueName string) string {
	idx := strings.Index(issueName, "-")
	if idx == -1 {
		//	not errCouldNotFound
		return ""
	}

	return issueName[idx+1:]
}

// chooseOneProjectInteractively if there are not only one project matched from local or remote,
// then let user know and do the choice.
func chooseOneProjectInteractively(projects []*repository.ProjectDO) (*repository.ProjectDO, error) {
	if len(projects) == 0 {
		return nil, errors.New("no project to choose")
	}

	if len(projects) == 1 {
		// if only one project found, then use this as target project
		return projects[0], nil
	}

	projectOptions := make([]string, len(projects))
	for idx, v := range projects {
		projectOptions[idx] = fmt.Sprintf("%d::%s::%d::%s", idx, v.ProjectName, v.ProjectID, v.WebURL)
	}

	qs := []*survey.Question{
		{
			Name: "projects",
			Prompt: &survey.Select{
				Message: "choose one project",
				Options: projectOptions,
			},
		},
	}
	r := struct {
		Idx int `survey:"projects"`
	}{}
	if err := survey.Ask(qs, &r); err != nil {
		return nil, errors.Wrap(err, "survey.Ask failed")
	}

	return projects[r.Idx], nil
}

// chooseOneMilestoneInteractively if there are not only one milestone matched from local or remote,
// then let user know and make a decision.
func chooseOneMilestoneInteractively(milestones []*repository.MilestoneDO) (*repository.MilestoneDO, error) {
	if len(milestones) == 0 {
		return nil, errors.New("no milestone to choose")
	}

	if len(milestones) == 1 {
		// if only one project found, then use this as target project
		return milestones[0], nil

	}

	milestoneOptions := make([]string, len(milestones))
	for idx, v := range milestones {
		milestoneOptions[idx] = v.Title
	}

	qs := []*survey.Question{
		{
			Name: "milestones",
			Prompt: &survey.Select{
				Message: "choose one milestone",
				Options: milestoneOptions,
			},
		},
	}
	r := struct {
		Idx int `survey:"milestones"`
	}{}
	if err := survey.Ask(qs, &r); err != nil {
		return nil, errors.Wrap(err, "survey.Ask failed")
	}

	return milestones[r.Idx], nil
}
