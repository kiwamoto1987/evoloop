package domain

// PatchResult holds the output from an LLM patch generation.
type PatchResult struct {
	PatchContent string `json:"patch_content"`
	RawOutput    string `json:"raw_output"`
}
