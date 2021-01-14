package gitop

// IGitOperator supports to manage the local git repository.
type IGitOperator interface {
	// Checkout local branch
	Checkout(branchName string, b bool) error

	// FetchOrigin fetch origin branches
	FetchOrigin() error

	CurrentBranch() (string, error)
}
