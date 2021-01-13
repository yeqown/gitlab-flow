package types

// BranchTyp
type BranchTyp string

func (b BranchTyp) String() string {
	return string(b)
}

const (
	MasterBranch BranchTyp = "master"
	DevBranch    BranchTyp = "develop"
	TestBranch   BranchTyp = "test"
)

const (
	FeatureBranchPrefix = "feature/"
)
