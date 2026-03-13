package service

import (
	"context"
	"math"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func newTestCostAttributionService() *CostAttributionService {
	logger, _ := zap.NewDevelopment()
	return NewCostAttributionService(logger)
}

func TestCostAttributionService_Attribute(t *testing.T) {
	tests := []struct {
		name             string
		hoursSaved       float64
		hourlyRate       float64
		wantValueSaved   float64
		wantCostPositive bool
		wantROI          float64
	}{
		{
			name:             "standard ROI calculation",
			hoursSaved:       10,
			hourlyRate:       100,
			wantValueSaved:   1000,
			wantCostPositive: true,
			wantROI:          566.67,
		},
		{
			name:             "zero hours saved - no division by zero",
			hoursSaved:       0,
			hourlyRate:       100,
			wantValueSaved:   0,
			wantCostPositive: false,
			wantROI:          0,
		},
		{
			name:             "negative hourly rate",
			hoursSaved:       10,
			hourlyRate:       -100,
			wantValueSaved:   -1000,
			wantCostPositive: false,
			wantROI:          566.67, // formula still applies: (v-c)/c*100
		},
		{
			name:             "very large values no overflow",
			hoursSaved:       1e10,
			hourlyRate:       1e5,
			wantValueSaved:   1e15,
			wantCostPositive: true,
			wantROI:          566.67,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestCostAttributionService()
			ctx := context.Background()
			projectID := uuid.New()

			attr, err := svc.Attribute(ctx, projectID, domain.AttributionInput{
				TraceID:    uuid.New().String(),
				IssueRef:   "ISSUE-1",
				IssueTitle: "Test Issue",
				Category:   "testing",
				HoursSaved: tt.hoursSaved,
				HourlyRate: tt.hourlyRate,
			})

			require.NoError(t, err)
			assert.Equal(t, tt.wantValueSaved, attr.EstimatedValueSaved)

			if tt.wantCostPositive {
				assert.True(t, attr.CostIncurred > 0)
			}

			assert.False(t, math.IsNaN(attr.ROI), "ROI should not be NaN")
			assert.False(t, math.IsInf(attr.ROI, 0), "ROI should not be Inf")

			if tt.hoursSaved != 0 && tt.hourlyRate > 0 {
				assert.InDelta(t, tt.wantROI, attr.ROI, 0.01)
			}
		})
	}
}

func TestCostAttributionService_GetReport(t *testing.T) {
	svc := newTestCostAttributionService()
	ctx := context.Background()
	projectID := uuid.New()

	// Add attributions in different categories
	_, err := svc.Attribute(ctx, projectID, domain.AttributionInput{
		TraceID:    uuid.New().String(),
		Category:   "code-review",
		HoursSaved: 5,
		HourlyRate: 100,
	})
	require.NoError(t, err)

	_, err = svc.Attribute(ctx, projectID, domain.AttributionInput{
		TraceID:    uuid.New().String(),
		Category:   "code-review",
		HoursSaved: 3,
		HourlyRate: 100,
	})
	require.NoError(t, err)

	_, err = svc.Attribute(ctx, projectID, domain.AttributionInput{
		TraceID:    uuid.New().String(),
		Category:   "testing",
		HoursSaved: 2,
		HourlyRate: 80,
	})
	require.NoError(t, err)

	report, err := svc.GetReport(ctx, projectID, domain.AttributionDateRange{})
	require.NoError(t, err)

	assert.Len(t, report.Attributions, 3)
	assert.True(t, report.TotalCost > 0)
	assert.True(t, report.TotalValueSaved > 0)
	assert.True(t, report.OverallROI > 0)

	// Check category aggregation
	assert.Len(t, report.ByCategory, 2)
	cr, ok := report.ByCategory["code-review"]
	assert.True(t, ok)
	assert.Equal(t, 2, cr.TraceCount)
	assert.True(t, cr.ROI > 0)

	tst, ok := report.ByCategory["testing"]
	assert.True(t, ok)
	assert.Equal(t, 1, tst.TraceCount)
}

func TestCostAttributionService_ListAttributions_NonExistentProject(t *testing.T) {
	svc := newTestCostAttributionService()
	ctx := context.Background()

	attrs, err := svc.ListAttributions(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, attrs)
}

func TestCostAttributionService_ConcurrentAccess(t *testing.T) {
	svc := newTestCostAttributionService()
	ctx := context.Background()
	projectID := uuid.New()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.Attribute(ctx, projectID, domain.AttributionInput{
				TraceID:    uuid.New().String(),
				Category:   "concurrent",
				HoursSaved: 1,
				HourlyRate: 50,
			})
		}()
	}
	wg.Wait()

	attrs, err := svc.ListAttributions(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, attrs, 50)
}
