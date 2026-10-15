package database

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openMigrationIntegrationDB provisions one uniquely named database from the
// explicitly supplied test DSN. It never reads DATABASE_URL and validates the
// source name before deriving an admin connection for this isolated cluster.
func openMigrationIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	rawDSN := os.Getenv("MENTRA_TEST_DATABASE_URL")
	if rawDSN == "" {
		t.Skip("MENTRA_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(rawDSN)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(strings.Trim(parsed.Path, "/"), "_test"), "MENTRA_TEST_DATABASE_URL must point to a database ending in _test")

	databaseName := "mentra_migration_" + strings.ReplaceAll(uuid.NewString(), "-", "") + "_test"
	adminURL := *parsed
	adminURL.Path = "/postgres"
	admin, err := gorm.Open(postgres.Open(adminURL.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, admin.Exec("CREATE DATABASE "+databaseName).Error)

	testURL := *parsed
	testURL.Path = "/" + databaseName
	db, err := gorm.Open(postgres.Open(testURL.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, closeErr := db.DB()
		if closeErr == nil {
			_ = sqlDB.Close()
		}
		// This name is generated in this test and is never the configured test
		// database itself, so cleanup cannot target user academic data.
		_ = admin.Exec("DROP DATABASE IF EXISTS " + databaseName + " WITH (FORCE)").Error
	})
	return db
}

func TestRunMigrations_CleanDatabaseAndRepeat(t *testing.T) {
	db := openMigrationIntegrationDB(t)
	require.NoError(t, RunMigrations(db))
	require.NoError(t, RunMigrations(db), "the migration must be idempotent on a clean database")
	assert.True(t, db.Migrator().HasTable(&AIDailyUsageModel{}))
	assert.True(t, postgresIndexExists(t, db, "idx_ai_daily_usage_org_user_day"))
	assert.True(t, db.Migrator().HasTable(&CourseGradeChangeModel{}))
	assert.True(t, postgresIndexExists(t, db, "idx_exam_attempt_one_in_progress"))
}

func TestRunMigrations_LegacyDuplicateActiveAttemptsRequireExplicitResolution(t *testing.T) {
	db := openMigrationIntegrationDB(t)
	// This represents the pre-index schema. Both rows are intentionally still
	// valid, so expiring either one automatically would be an academic choice.
	require.NoError(t, db.Exec(`
		CREATE TABLE exam_attempts (
			id uuid PRIMARY KEY,
			exam_id uuid NOT NULL,
			student_id uuid NOT NULL,
			org_id uuid NOT NULL,
			status text NOT NULL DEFAULT 'in_progress',
			started_at timestamptz,
			expires_at timestamptz,
			submitted_at timestamptz,
			mcq_answers jsonb,
			mcq_score integer,
			mcq_max_score integer,
			file_path text,
			file_feedback text,
			file_score integer,
			file_points integer,
			total_score integer,
			graded_by uuid,
			graded_at timestamptz,
			created_at timestamptz,
			updated_at timestamptz
		)
	`).Error)

	firstID, secondID, examID, studentID, orgID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	for _, attemptID := range []uuid.UUID{firstID, secondID} {
		require.NoError(t, db.Exec(`
			INSERT INTO exam_attempts (id, exam_id, student_id, org_id, status, started_at, expires_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, 'in_progress', ?, ?, ?, ?)
		`, attemptID.String(), examID.String(), studentID.String(), orgID.String(), now, now.Add(time.Hour), now, now).Error)
	}

	err := RunMigrations(db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "require explicit academic resolution")
	assert.False(t, postgresIndexExists(t, db, "idx_exam_attempt_one_in_progress"))
	var activeCount int64
	require.NoError(t, db.Table("exam_attempts").Where("exam_id = ? AND student_id = ? AND status = ?", examID.String(), studentID.String(), "in_progress").Count(&activeCount).Error)
	assert.EqualValues(t, 2, activeCount, "migration must not choose or discard an active academic attempt")

	// A human-approved resolution explicitly selects the attempt to close. The
	// upgrade can then continue and remains idempotent on subsequent runs.
	require.NoError(t, db.Exec("UPDATE exam_attempts SET status = 'expired', updated_at = ? WHERE id = ?", now, firstID.String()).Error)
	require.NoError(t, RunMigrations(db))
	require.NoError(t, RunMigrations(db))
	assert.True(t, postgresIndexExists(t, db, "idx_exam_attempt_one_in_progress"))
}

func postgresIndexExists(t *testing.T, db *gorm.DB, name string) bool {
	t.Helper()
	var count int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM pg_indexes WHERE schemaname = current_schema() AND indexname = ?`, name).Scan(&count).Error)
	return count == 1
}
