package llm

import "github.com/kiwamoto1987/evoloop/internal/domain"

// LanguageModelClient abstracts LLM interaction for patch generation.
type LanguageModelClient interface {
	GeneratePatch(input *domain.PromptContext) (*domain.PatchResult, error)
}
