package repositories

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ailms/backend/internal/infrastructure/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGORMAIUsageRepository_ConcurrentRequestsShareOneAtomicLimit(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	orgID, userID := uuid.New(), uuid.New()
	day := time.Now().UTC()
	t.Cleanup(func() {
		_ = db.Where("org_id = ? AND user_id = ?", orgID.String(), userID.String()).Delete(&database.AIDailyUsageModel{}).Error
	})

	const attempts, limit = 20, 5
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	errors := make(chan error, attempts)
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Separate repository instances model concurrent backend replicas.
			_, ok, err := NewGORMAIUsageRepository(db).Consume(context.Background(), orgID, userID, day, limit)
			if err != nil {
				errors <- err
				return
			}
			if ok {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	require.Equal(t, limit, allowed)
	var count int64
	require.NoError(t, db.Model(&database.AIDailyUsageModel{}).Where("org_id = ? AND user_id = ?", orgID.String(), userID.String()).Select("count").Scan(&count).Error)
	require.EqualValues(t, limit, count)
	// A fresh repository instance sees the same persisted counter.
	_, permitted, err := NewGORMAIUsageRepository(db).Consume(context.Background(), orgID, userID, day, limit)
	require.NoError(t, err)
	require.False(t, permitted)
}
