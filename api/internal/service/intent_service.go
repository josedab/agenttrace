package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type IntentService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	intents map[string]*domain.IntentDeclaration // intentID -> declaration
}

func NewIntentService(logger *zap.Logger) *IntentService {
	return &IntentService{
		logger:  logger,
		intents: make(map[string]*domain.IntentDeclaration),
	}
}

func (s *IntentService) DeclareIntent(ctx context.Context, projectID uuid.UUID, input domain.IntentInput) (*domain.IntentDeclaration, error) {
	decl := &domain.IntentDeclaration{
		ID:              uuid.New(),
		ProjectID:       projectID,
		TraceID:         input.TraceID,
		AgentName:       input.AgentName,
		DeclaredIntent:  input.DeclaredIntent,
		DeclaredActions: input.DeclaredActions,
		Status:          "pending",
		AlignmentScore:  0,
		DeclaredAt:      time.Now(),
	}

	s.mu.Lock()
	s.intents[decl.ID.String()] = decl
	s.mu.Unlock()

	s.logger.Info("intent declared", zap.String("intentId", decl.ID.String()), zap.String("agent", input.AgentName))
	return decl, nil
}

func (s *IntentService) VerifyIntent(ctx context.Context, intentID uuid.UUID, actualActions []string) (*domain.IntentDeclaration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	decl, ok := s.intents[intentID.String()]
	if !ok {
		return nil, fmt.Errorf("intent declaration not found: %s", intentID)
	}

	decl.ActualActions = actualActions
	now := time.Now()
	decl.VerifiedAt = &now

	// Compare declared vs actual actions
	matched := 0
	var misalignments []domain.IntentMisalignment

	declaredSet := make(map[string]bool)
	for _, a := range decl.DeclaredActions {
		declaredSet[strings.ToLower(a)] = true
	}

	for _, actual := range actualActions {
		found := false
		for _, declared := range decl.DeclaredActions {
			if strings.EqualFold(actual, declared) || strings.Contains(strings.ToLower(actual), strings.ToLower(declared)) {
				found = true
				matched++
				break
			}
		}
		if !found {
			severity := "minor"
			if !declaredSet[strings.ToLower(actual)] {
				severity = "major"
			}
			misalignments = append(misalignments, domain.IntentMisalignment{
				DeclaredAction: "",
				ActualAction:   actual,
				Severity:       severity,
				Description:    fmt.Sprintf("action '%s' was not declared in intent", actual),
			})
		}
	}

	// Check for declared actions that weren't performed
	for _, declared := range decl.DeclaredActions {
		found := false
		for _, actual := range actualActions {
			if strings.EqualFold(declared, actual) || strings.Contains(strings.ToLower(actual), strings.ToLower(declared)) {
				found = true
				break
			}
		}
		if !found {
			misalignments = append(misalignments, domain.IntentMisalignment{
				DeclaredAction: declared,
				ActualAction:   "",
				Severity:       "minor",
				Description:    fmt.Sprintf("declared action '%s' was not performed", declared),
			})
		}
	}

	total := len(decl.DeclaredActions) + len(actualActions)
	if total > 0 {
		decl.AlignmentScore = float64(matched*2) / float64(total)
	} else {
		decl.AlignmentScore = 1.0
	}

	decl.Misalignments = misalignments
	if decl.AlignmentScore >= 0.8 {
		decl.Status = "verified"
	} else {
		decl.Status = "misaligned"
	}

	return decl, nil
}

func (s *IntentService) GetVerification(ctx context.Context, intentID uuid.UUID) (*domain.IntentDeclaration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	decl, ok := s.intents[intentID.String()]
	if !ok {
		return nil, fmt.Errorf("intent declaration not found: %s", intentID)
	}
	return decl, nil
}

func (s *IntentService) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.IntentVerificationStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &domain.IntentVerificationStats{
		ProjectID:            projectID,
		MisalignmentsByAgent: make(map[string]int),
	}

	verified := 0
	for _, decl := range s.intents {
		if decl.ProjectID != projectID {
			continue
		}
		if decl.Status == "pending" {
			continue
		}
		stats.TotalVerifications++
		if decl.Status == "verified" {
			verified++
		} else {
			stats.MisalignmentsByAgent[decl.AgentName]++
		}
	}

	if stats.TotalVerifications > 0 {
		stats.AlignmentRate = float64(verified) / float64(stats.TotalVerifications)
	}

	return stats, nil
}
