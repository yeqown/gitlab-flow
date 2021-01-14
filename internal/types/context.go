package types

// FlowContext contains all necessary parameters of flow command to execute.
type FlowContext struct {
	// Conf of flow CLI.
	Conf *Config
	// CWD current working directory.
	CWD string
	// project of current working directory, normally, get from current working directory.
	Project *ProjectBasics

	//// Branch current branch name.
	//ProjectBranch string

	ConfigFilePath string
	DBFilePath     string
}

// ProjectBasics contains basic attributes of project.
type ProjectBasics struct {
	ID     int
	Name   string
	WebURL string
}
