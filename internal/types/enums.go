package types

// BranchTyp is builtin branch type to limit parameter passing
type BranchTyp string

func (b BranchTyp) String() string {
	return string(b)
}

// TODO(@yeqown) MasterBranch, DevBranch and TestBranch could be customized.
var (
	MasterBranch BranchTyp = "master"
	DevBranch    BranchTyp = "develop"
	TestBranch   BranchTyp = "test"
)
