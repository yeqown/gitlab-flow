package gitop

// IGitOperator supports to manage the local git repository.
type IGitOperator interface {
	// Checkout local branch
	Checkout(branchName string, b bool) error

	// FetchOrigin fetch origin branches
	FetchOrigin() error

	// CurrentBranch just get current branch of your target repository.
	CurrentBranch() (string, error)

	// Merge would merge source into target branch. If current branch is not your target branch,
	// this function would automatically checkout, then execute the merge command.
	Merge(source, target string) error
}
