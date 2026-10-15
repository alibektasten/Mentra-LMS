package repositories

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/database"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMAttendanceRepository struct{ db *gorm.DB }

func NewGORMAttendanceRepository(db *gorm.DB) *GORMAttendanceRepository {
	return &GORMAttendanceRepository{db: db}
}

func (r *GORMAttendanceRepository) ListByGroupAndDate(ctx context.Context, groupID uuid.UUID, date time.Time) ([]entities.AttendanceRecord, error) {
	var models []database.AttendanceRecordModel
	err := r.db.WithContext(ctx).Where("group_id = ? AND date = ?", groupID.String(), attendanceDay(date)).Order("created_at ASC").Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]entities.AttendanceRecord, len(models))
	for i := range models {
		result[i] = *toAttendanceEntity(&models[i])
	}
	return result, nil
}

func (r *GORMAttendanceRepository) UpsertMany(ctx context.Context, records []entities.AttendanceRecord) error {
	if len(records) == 0 {
		return nil
	}
	models := make([]database.AttendanceRecordModel, len(records))
	for i, record := range records {
		models[i] = *toAttendanceModel(&record)
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "student_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "marked_by", "marked_at", "updated_at"}),
	}).Create(&models).Error
	if err != nil {
		return apperrors.InternalError(err.Error())
	}
	return nil
}

func attendanceDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func toAttendanceModel(record *entities.AttendanceRecord) *database.AttendanceRecordModel {
	return &database.AttendanceRecordModel{
		ID: record.ID.String(), GroupID: record.GroupID.String(), StudentID: record.StudentID.String(), OrgID: record.OrgID.String(),
		Date: attendanceDay(record.Date), Status: string(record.Status), MarkedBy: record.MarkedBy.String(), MarkedAt: record.MarkedAt, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}

func toAttendanceEntity(model *database.AttendanceRecordModel) *entities.AttendanceRecord {
	id, _ := uuid.Parse(model.ID)
	groupID, _ := uuid.Parse(model.GroupID)
	studentID, _ := uuid.Parse(model.StudentID)
	orgID, _ := uuid.Parse(model.OrgID)
	markedBy, _ := uuid.Parse(model.MarkedBy)
	return &entities.AttendanceRecord{ID: id, GroupID: groupID, StudentID: studentID, OrgID: orgID, Date: model.Date, Status: entities.AttendanceStatus(model.Status), MarkedBy: markedBy, MarkedAt: model.MarkedAt, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}
