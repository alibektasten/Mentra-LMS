package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AIUsageRepository consumes one daily request in a single atomic database
// operation. A rejected or failed database operation never authorizes a call.
type AIUsageRepository interface {
	Consume(ctx context.Context, orgID, userID uuid.UUID, day time.Time, limit int) (remaining int, allowed bool, err error)
}
