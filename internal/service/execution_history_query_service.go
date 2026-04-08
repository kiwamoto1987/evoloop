package service

import (
	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/repository"
)

// HistorySummary holds combined history data for display.
type HistorySummary struct {
	Issues      []*domain.ImplementationIssue `json:"issues"`
	Executions  []*domain.ExecutionRecord     `json:"executions"`
	Evaluations []*domain.EvaluationReport    `json:"evaluations"`
}

// ExecutionHistoryQueryService queries execution history.
type ExecutionHistoryQueryService struct {
	issueRepo      *repository.ImplementationIssueRepository
	executionRepo  *repository.ExecutionHistoryRepository
	evaluationRepo *repository.EvaluationReportRepository
}

// NewExecutionHistoryQueryService creates a new query service.
func NewExecutionHistoryQueryService(
	issueRepo *repository.ImplementationIssueRepository,
	executionRepo *repository.ExecutionHistoryRepository,
	evaluationRepo *repository.EvaluationReportRepository,
) *ExecutionHistoryQueryService {
	return &ExecutionHistoryQueryService{
		issueRepo:      issueRepo,
		executionRepo:  executionRepo,
		evaluationRepo: evaluationRepo,
	}
}

// QueryAll retrieves all history data.
func (s *ExecutionHistoryQueryService) QueryAll() (*HistorySummary, error) {
	issues, err := s.issueRepo.FindAll()
	if err != nil {
		return nil, err
	}

	executions, err := s.executionRepo.FindAll()
	if err != nil {
		return nil, err
	}

	evaluations, err := s.evaluationRepo.FindAll()
	if err != nil {
		return nil, err
	}

	return &HistorySummary{
		Issues:      issues,
		Executions:  executions,
		Evaluations: evaluations,
	}, nil
}
