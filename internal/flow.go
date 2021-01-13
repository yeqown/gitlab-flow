package internal

import (
	"path/filepath"
	"strings"
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
