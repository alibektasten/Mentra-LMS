package repositories

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// lockExamConfiguration is shared by structural exam edits and first attempt
// admission. PostgreSQL keeps the advisory lock for the surrounding database
// transaction, so the related read/check/write steps cannot interleave.
func lockExamConfiguration(tx *gorm.DB, examID uuid.UUID) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	return tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", fmt.Sprintf("exam:%s", examID)).Error
}
