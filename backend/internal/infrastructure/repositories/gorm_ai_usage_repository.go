package repositories

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMAIUsageRepository struct{ db *gorm.DB }

func NewGORMAIUsageRepository(db *gorm.DB) *GORMAIUsageRepository {
	return &GORMAIUsageRepository{db: db}
}

// The conflict update is conditional on the current count. PostgreSQL locks
// the conflicting row while evaluating that predicate, so concurrent app
// instances cannot both consume the last available unit.
func (r *GORMAIUsageRepository) Consume(ctx context.Context, orgID, userID uuid.UUID, day time.Time, limit int) (int, bool, error) {
	if limit <= 0 {
		return 0, true, nil
	}
	var row struct{ Count int }
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO ai_daily_usage (org_id, user_id, usage_date, count, updated_at)
		VALUES (?, ?, ?, 1, NOW())
		ON CONFLICT (org_id, user_id, usage_date)
		DO UPDATE SET count = ai_daily_usage.count + 1, updated_at = NOW()
		WHERE ai_daily_usage.count < ?
		RETURNING count
	`, orgID.String(), userID.String(), day.UTC().Format("2006-01-02"), limit).Scan(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, apperrors.InternalError("AI usage could not be recorded")
	}
	if row.Count == 0 {
		return 0, false, nil
	}
	return limit - row.Count, true, nil
}
