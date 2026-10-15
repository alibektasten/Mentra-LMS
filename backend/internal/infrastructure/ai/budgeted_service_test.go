package ai

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ailms/backend/internal/domain/services"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeUsage struct {
	calls   int
	allowed bool
	err     error
}

func (f *fakeUsage) Consume(_ context.Context, _, _ uuid.UUID, _ time.Time, _ int) (int, bool, error) {
	f.calls++
	return 0, f.allowed, f.err
}

func TestBudgetedServiceRejectsBeforeProviderWhenLimitReached(t *testing.T) {
	usage := &fakeUsage{allowed: false}
	provider := new(mocks.MockAIService)
	service := NewBudgetedService(provider, usage, 1)
	ctx := services.WithAIIdentity(context.Background(), uuid.New(), uuid.New())

	_, err := service.GenerateLessonSummary(ctx, "lesson")

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusTooManyRequests, appErr.HTTPStatus)
	assert.Equal(t, 1, usage.calls)
	provider.AssertNotCalled(t, "GenerateLessonSummary", mock.Anything, mock.Anything)
}

func TestBudgetedServiceDoesNotChargeWithoutVerifiedTenantIdentity(t *testing.T) {
	usage := &fakeUsage{allowed: true}
	provider := new(mocks.MockAIService)
	service := NewBudgetedService(provider, usage, 1)
	_, err := service.GenerateLessonSummary(context.Background(), "lesson")
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnauthorized, appErr.HTTPStatus)
	assert.Zero(t, usage.calls)
	provider.AssertNotCalled(t, "GenerateLessonSummary", mock.Anything, mock.Anything)
}

func TestBudgetedServiceDoesNotCallProviderOnDatabaseFailure(t *testing.T) {
	usage := &fakeUsage{err: errors.New("database unavailable")}
	provider := new(mocks.MockAIService)
	service := NewBudgetedService(provider, usage, 1)
	ctx := services.WithAIIdentity(context.Background(), uuid.New(), uuid.New())

	_, err := service.GenerateRemediation(ctx, "lesson", []string{"question"})

	require.Error(t, err)
	provider.AssertNotCalled(t, "GenerateRemediation", mock.Anything, mock.Anything, mock.Anything)
}

func TestBudgetedServiceChargesAttemptedProviderCall(t *testing.T) {
	usage := &fakeUsage{allowed: true}
	provider := new(mocks.MockAIService)
	provider.On("GenerateLessonSummary", mock.Anything, "lesson").Return("", errors.New("provider unavailable"))
	service := NewBudgetedService(provider, usage, 1)
	ctx := services.WithAIIdentity(context.Background(), uuid.New(), uuid.New())

	_, err := service.GenerateLessonSummary(ctx, "lesson")

	require.Error(t, err)
	assert.Equal(t, 1, usage.calls)
	provider.AssertExpectations(t)
}
