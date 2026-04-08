package domain

// PromptContext holds all context needed to generate a patch proposal.
type PromptContext struct {
	ProjectRootPath string `json:"project_root_path"`

	IssueId          string `json:"issue_id"`
	IssueTitle       string `json:"issue_title"`
	IssueDescription string `json:"issue_description"`

	AcceptanceCriteria   []string          `json:"acceptance_criteria,omitempty"`
	TargetPaths          []string          `json:"target_paths,omitempty"`
	RelevantFileContents map[string]string `json:"relevant_file_contents,omitempty"`

	ClaudeContextContent string `json:"claude_context_content,omitempty"`
}
