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
	FeatureBegin(opc *types.OpFeatureContext, title, desc string) error
	// FeatureDebugging open a MergeRequest of feature branch and types.DevBranch branch.
	FeatureDebugging(opc *types.OpFeatureContext) error
	// FeatureTest open a MergeRequest of feature branch and types.TestBranch branch.
	FeatureTest(opc *types.OpFeatureContext) error
	// FeatureRelease open a MergeRequest of feature branch and types.MasterBranch branch.
	FeatureRelease(opc *types.OpFeatureContext) error
	// DONE(@yeqown) this would be useful while you merge feature into master but there is conflict.

	// FeatureResolveConflict will checkout a new branch from target branch,
	// then create a merge request from current feature branch to the new branch.
	// newBranch = "resolve-conflict/featureBranchName-to-master"
	FeatureResolveConflict(opc *types.OpFeatureContext, targetBranch types.BranchTyp) error

	// FeatureBeginIssue checkout a issue branch from feature branch, also open a merge request
	// which is from issue branch to feature branch.
	FeatureBeginIssue(opc *types.OpFeatureContext, title, desc string) error
	// FeatureFinishIssue open the WebURL of merge request which is from issue branch to feature branch.
	FeatureFinishIssue(opc *types.OpFeatureContext, issueBranchName string) error

	// HotfixBegin checkout a hotfix branch from types.MasterBranch, also open a merge request
	// which is from hotfix branch to types.MasterBranch.
	HotfixBegin(title, desc string) error
	// HotfixFinish open the WebURL of merge request which is from hotfix branch to types.MasterBranch.
	HotfixFinish(hotfixBranchName string) error

	// SyncProject synchronize project information from remote gitlab server.
	SyncProject() error
	// SyncMilestone synchronize remote repository milestone and related issues / merge requests to local.
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
	IssueBranchPrefix           = "issue/"
)

// genFeatureBranchName
func genFeatureBranchName(name string) string {
	if strings.HasPrefix(name, FeatureBranchPrefix) {
		return name
	}

	return FeatureBranchPrefix + name
}

// tryParseFeatureNameFrom try parse feature name from issue name or other cases.
// if branchName has no prefix, which means tryParseFeatureNameFrom could judge it.
// otherwise, branchName must have prefix, which was one of flow branch prefixes.
func tryParseFeatureNameFrom(branchName string, comptaible bool) (string, bool) {
	arr := strings.Split(branchName, "/")
	if len(arr) < 2 {
		return "", false
	}

	prefix := arr[0] + "/"
	switch prefix {
	case FeatureBranchPrefix:
		// pass
	case HotfixBranchPrefix:
		// pass
	case ConflictResolveBranchPrefix:
		// pass
	case IssueBranchPrefix:
		out := strings.Join(arr[1:], "/")
		if out = parseFeatureFromIssueName(out, comptaible); out != "" {
			return out, true
		}
	case "":
		fallthrough
	default:
		return "", false
	}

	return strings.Join(arr[1:], "/"), true
}

// isFeatureName judge whether branchName is feature branch or not.
// strings.HasPrefix(name, FeatureBranchPrefix)
func isFeatureName(name string) bool {
	return strings.HasPrefix(name, FeatureBranchPrefix)
}

// genHotfixBranchName .
func genHotfixBranchName(name string) string {
	if strings.HasPrefix(name, HotfixBranchPrefix) {
		return name
	}

	return HotfixBranchPrefix + name
}

// genMergeRequestName generate merge request name.
func genMergeRequestName(srcBranch, targetBranch string) string {
	return fmt.Sprintf("Merge %s into %s", srcBranch, targetBranch)
}

// genIssueBranchName .
// @result = issue/milestoneTitle-1 as default
func genIssueBranchName(name string, issueIID int) string {
	if strings.HasPrefix(name, IssueBranchPrefix) {
		return name
	}

	return IssueBranchPrefix + name + "-" + strconv.Itoa(issueIID)
}

// parseFeatureFromIssueName parse issue name to feature name, there are
// two different cases:
// 1. "1-milestoneName"
// 2. "issue/milestoneName-1"
//
// TODO(@yeqown): comptaible with old.
func parseFeatureFromIssueName(issueName string, compatible bool) string {
	// if comptaible, try parse "1-milestoneName"
	if compatible {
		// DONE(@yeqown): support "1-milestoneName"
		idx := strings.Index(issueName, "-")
		if idx == -1 {
			return ""
		}

		return issueName[idx+1:]
	}

	issueName = strings.TrimPrefix(issueName, IssueBranchPrefix)
	idx := strings.LastIndex(issueName, "-")
	if idx == -1 {
		//	not errCouldNotFound
		return ""
	}

	return issueName[:idx]
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
