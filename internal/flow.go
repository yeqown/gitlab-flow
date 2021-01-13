package internal

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/yeqown/log"
)

// IFlow to control branches, MRs, milestones and issues.
type IFlow interface {
	FeatureStart(title, desc string) error
	FeatureDebugging(featureBranchName string) error
	FeatureTest(featureBranchName string) error
	FeatureRelease(featureBranchName string) error

	FeatureStartIssue(featureBranchName string, params ...string) error
	FeatureFinishIssue(featureBranchName, issueBranchName string) error

	HotfixStart(title, desc string) error
	HotfixRelease(hotfixBranchName string) error

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
		log.WithField("find", string(data[1])).
			Warn("parse issue iid from desc: %v", err)
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
