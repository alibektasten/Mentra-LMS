package usecases

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type AttendanceUseCase struct {
	attendanceRepo repositories.AttendanceRepository
	groupRepo      repositories.GroupRepository
	userRepo       repositories.UserRepository
}

func NewAttendanceUseCase(attendanceRepo repositories.AttendanceRepository, groupRepo repositories.GroupRepository, userRepo repositories.UserRepository) *AttendanceUseCase {
	return &AttendanceUseCase{attendanceRepo: attendanceRepo, groupRepo: groupRepo, userRepo: userRepo}
}

func (uc *AttendanceUseCase) GetSheet(ctx context.Context, courseID, groupID, orgID uuid.UUID, date time.Time) (*dto.AttendanceSheetDTO, error) {
	if err := uc.requireCourseGroup(ctx, courseID, groupID, orgID); err != nil {
		return nil, err
	}
	members, err := uc.groupRepo.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	studentIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		studentIDs = append(studentIDs, member.StudentID)
	}
	users, err := uc.userRepo.FindByIDs(ctx, studentIDs)
	if err != nil {
		return nil, err
	}
	usersByID := make(map[uuid.UUID]entities.User, len(users))
	for _, user := range users {
		usersByID[user.ID] = user
	}
	records, err := uc.attendanceRepo.ListByGroupAndDate(ctx, groupID, attendanceDate(date))
	if err != nil {
		return nil, err
	}
	statusByStudent := make(map[uuid.UUID]entities.AttendanceStatus, len(records))
	for _, record := range records {
		statusByStudent[record.StudentID] = record.Status
	}
	students := make([]dto.AttendanceStudentDTO, 0, len(studentIDs))
	for _, studentID := range studentIDs {
		user, ok := usersByID[studentID]
		if !ok {
			continue
		}
		status := statusByStudent[studentID]
		students = append(students, dto.AttendanceStudentDTO{StudentID: studentID.String(), Name: user.Name, Email: user.Email, Status: string(status)})
	}
	sort.Slice(students, func(i, j int) bool { return strings.ToLower(students[i].Name) < strings.ToLower(students[j].Name) })
	return &dto.AttendanceSheetDTO{GroupID: groupID.String(), CourseID: courseID.String(), Date: attendanceDate(date), Students: students}, nil
}

func (uc *AttendanceUseCase) SaveSheet(ctx context.Context, courseID, groupID, orgID, markerID uuid.UUID, req dto.SaveAttendanceRequest) (*dto.AttendanceSheetDTO, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(req.Date))
	if err != nil {
		return nil, apperrors.ValidationError("date must use YYYY-MM-DD")
	}
	if err := uc.requireCourseGroup(ctx, courseID, groupID, orgID); err != nil {
		return nil, err
	}
	members, err := uc.groupRepo.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	enrolled := make(map[uuid.UUID]struct{}, len(members))
	for _, member := range members {
		enrolled[member.StudentID] = struct{}{}
	}
	if len(req.Records) != len(enrolled) {
		return nil, apperrors.ValidationError("attendance must include every enrolled student exactly once")
	}
	seen := make(map[uuid.UUID]struct{}, len(req.Records))
	now := time.Now()
	records := make([]entities.AttendanceRecord, 0, len(req.Records))
	for _, input := range req.Records {
		studentID, err := uuid.Parse(input.StudentID)
		if err != nil {
			return nil, apperrors.ValidationError("invalid student_id")
		}
		if _, ok := enrolled[studentID]; !ok {
			return nil, apperrors.ForbiddenError("student is not enrolled in this group")
		}
		if _, duplicated := seen[studentID]; duplicated {
			return nil, apperrors.ValidationError("each student can be marked only once")
		}
		status, err := attendanceStatus(input.Status)
		if err != nil {
			return nil, err
		}
		seen[studentID] = struct{}{}
		records = append(records, entities.AttendanceRecord{ID: uuid.New(), GroupID: groupID, StudentID: studentID, OrgID: orgID, Date: attendanceDate(date), Status: status, MarkedBy: markerID, MarkedAt: now, CreatedAt: now, UpdatedAt: now})
	}
	if err := uc.attendanceRepo.UpsertMany(ctx, records); err != nil {
		return nil, err
	}
	return uc.GetSheet(ctx, courseID, groupID, orgID, date)
}

func (uc *AttendanceUseCase) requireCourseGroup(ctx context.Context, courseID, groupID, orgID uuid.UUID) error {
	group, err := uc.groupRepo.GetGroupByID(ctx, groupID, orgID)
	if err != nil {
		return err
	}
	if group.CourseID == nil || *group.CourseID != courseID {
		return apperrors.ForbiddenError("group does not belong to this course")
	}
	return nil
}

func attendanceDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func attendanceStatus(value string) (entities.AttendanceStatus, error) {
	switch entities.AttendanceStatus(strings.ToLower(strings.TrimSpace(value))) {
	case entities.AttendancePresent:
		return entities.AttendancePresent, nil
	case entities.AttendanceLate:
		return entities.AttendanceLate, nil
	case entities.AttendanceAbsent:
		return entities.AttendanceAbsent, nil
	default:
		return "", apperrors.ValidationError("attendance status must be present, late, or absent")
	}
}
