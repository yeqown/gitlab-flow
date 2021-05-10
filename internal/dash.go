package internal

// IDash is used to display useful data of current development stage,
// and also to analyze user developing data.
type IDash interface {
	// FeatureDetail get feature detail
	FeatureDetail(featureBranchName string) ([]byte, error)

	// MilestoneOverview get milestone detail
	MilestoneOverview(milestoneName, branchFilter string) ([]byte, error)

	// ProjectDetail display project detail， includes: project web URL
	ProjectDetail() ([]byte, error)
}
