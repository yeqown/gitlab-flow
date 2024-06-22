package types

// BranchTyp is builtin branch type to limit parameter passing
type BranchTyp string

func (b BranchTyp) String() string {
	return string(b)
}

// DONE(@yeqown) MasterBranch, DevBranch and TestBranch could be customized.
var (
	MasterBranch BranchTyp = "master"
	DevBranch    BranchTyp = "develop"
	TestBranch   BranchTyp = "test"
)

// SetBranchSetting reset builtin branch enums manually.
func SetBranchSetting(master, dev, test BranchTyp) {
	if master != "" {
		MasterBranch = master
	}
	if dev != "" {
		DevBranch = dev
	}
	if test != "" {
		TestBranch = test
	}
}

var (
	FeatureBranchPrefix         = "feature/"
	HotfixBranchPrefix          = "hotfix/"
	ConflictResolveBranchPrefix = "conflict-resolve/"
	IssueBranchPrefix           = "issue/"
)

func SetBranchPrefix(feature, hotfix, conflictResolve, issue string) {
	if feature != "" {
		FeatureBranchPrefix = feature
	}
	if hotfix != "" {
		HotfixBranchPrefix = hotfix
	}
	if conflictResolve != "" {
		ConflictResolveBranchPrefix = conflictResolve
	}
	if issue != "" {
		IssueBranchPrefix = issue
	}
}
