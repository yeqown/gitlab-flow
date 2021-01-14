package internal

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/yeqown/gitlab-flow/internal/conf"
	"github.com/yeqown/gitlab-flow/internal/types"

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
	// TODO(@yeqown) this would be useful while you merge feature into master but there is conflict.
	// FeatureResolveConflict will checkout a new branch from target branch,
	// then create a merge request from current feature branch to the new branch.
	// newBranch = "resolve-conflict/featureBranchName-to-master"
	// FeatureResolveConflict(featureBranchName string) error

	// FeatureBeginIssue checkout a issue branch from feature branch, also open a merge request
	// which is from issue branch to feature branch.
	FeatureBeginIssue(featureBranchName string, params ...string) error
	// FeatureFinishIssue open the WebURL of merge request which is from issue branch to feature branch.
	FeatureFinishIssue(featureBranchName, issueBranchName string) error

	// HotfixStart checkout a hotfix branch from types.MasterBranch, also open a merge request
	// which is from hotfix branch to types.MasterBranch.
	HotfixBegin(title, desc string) error
	// HotfixRelease open the WebURL of merge request which is from hotfix branch to types.MasterBranch.
	HotfixFinish(hotfixBranchName string) error

	// SyncMilestone synchronize remote repository milestone and
	// related issues / merge requests to local.
	SyncMilestone(milestoneID int) error
}

// extractProjectNameFromCWD get project name from current working directory.
func extractProjectNameFromCWD(cwd string) string {
	splited := strings.Split(cwd, string(filepath.Separator))
	return splited[len(splited)-1]
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

func notBuiltinBranch(branchName string) bool {
	switch branchName {
	case types.DevBranch.String(), types.TestBranch.String(), types.MasterBranch.String():
		return true
	}
	return false
}

// NewContext be generated with non project information.
// Do not use Project directly!!!
func NewContext(cwd, configFilePath string) *types.FlowContext {
	c, err := conf.Load(configFilePath, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"cwd":            cwd,
			"configFilePath": configFilePath,
		}).Fatalf("load config failed: %v", err)
	}

	return &types.FlowContext{
		Conf:           c,
		CWD:            cwd,
		Project:        nil,
		ConfigFilePath: configFilePath,
		DBFilePath:     "", // TODO(@yeqown): fill the logic to calc the db filepath
	}
}
