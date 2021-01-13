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
}

// NewContext be generated with non project information.
// Do not use Project directly!!!
func NewContext(conf *Config, cwd string) *FlowContext {
	return &FlowContext{
		Conf:    conf,
		CWD:     cwd,
		Project: nil,
	}
}

// ProjectBasics contains basic attributes of project.
type ProjectBasics struct {
	ID     int
	Name   string
	WebURL string
}
