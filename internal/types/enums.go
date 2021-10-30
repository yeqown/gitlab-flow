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

// SyncBranchSetting reset builtin branch enums manually.
func SyncBranchSetting(master, dev, test BranchTyp) {
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
