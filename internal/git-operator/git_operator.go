package gitop

// IGitOperator supports to manage the local git repository.
type IGitOperator interface {
	// Checkout loacl branch
	Checkout(branchName string, b bool) error

	// FetchOrigin fetch origin branches
	FetchOrigin() error
}

func New(repoDir string) IGitOperator {
	return NewBasedCmd(repoDir)
}
