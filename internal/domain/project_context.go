package domain

// ProjectContext holds the inspection result for a Git project.
type ProjectContext struct {
	ProjectRootPath string `json:"project_root_path"`
	CurrentBranch   string `json:"current_branch"`
	IsGitRepository bool   `json:"is_git_repository"`
	IsDirty         bool   `json:"is_dirty"`

	TestCommand      string `json:"test_command,omitempty"`
	LintCommand      string `json:"lint_command,omitempty"`
	TypeCheckCommand string `json:"typecheck_command,omitempty"`
}
