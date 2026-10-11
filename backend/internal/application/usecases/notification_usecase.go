package usecases

import (
	"context"
	"sort"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/google/uuid"
)

// NotificationUseCase turns course announcements into a single, role-aware
// inbox. It deliberately exposes only courses the current member can access.
type NotificationUseCase struct {
	announcements repositories.AnnouncementRepository
	courses       repositories.CourseRepository
	groups        repositories.GroupRepository
	teachers      repositories.CourseTeacherRepository
}

func NewNotificationUseCase(announcements repositories.AnnouncementRepository, courses repositories.CourseRepository, groups repositories.GroupRepository, teachers repositories.CourseTeacherRepository) *NotificationUseCase {
	return &NotificationUseCase{announcements: announcements, courses: courses, groups: groups, teachers: teachers}
}

func (uc *NotificationUseCase) List(ctx context.Context, userID, orgID uuid.UUID, role string) ([]dto.NotificationDTO, error) {
	var courseIDs []uuid.UUID
	var err error
	switch role {
	case string(entities.RoleStudent):
		courseIDs, err = uc.groups.FindCourseIDsByStudent(ctx, userID, orgID)
	case string(entities.RoleTeacher):
		courseIDs, err = uc.teachers.FindCourseIDsByTeacher(ctx, userID, orgID)
	default:
		courses, _, listErr := uc.courses.FindByOrg(ctx, orgID, 1, 100)
		if listErr != nil {
			return nil, listErr
		}
		for _, course := range courses {
			courseIDs = append(courseIDs, course.ID)
		}
	}
	if err != nil {
		return nil, err
	}

	items := make([]dto.NotificationDTO, 0)
	for _, courseID := range courseIDs {
		course, err := uc.courses.FindByID(ctx, courseID, orgID)
		if err != nil {
			return nil, err
		}
		announcements, _, err := uc.announcements.ListByCourse(ctx, orgID, courseID, 10, 0)
		if err != nil {
			return nil, err
		}
		for _, announcement := range announcements {
			items = append(items, dto.NotificationDTO{ID: announcement.ID.String(), Type: "announcement", Title: announcement.Title, Body: announcement.Content, CourseID: courseID.String(), CourseTitle: course.Title, CreatedAt: announcement.CreatedAt})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if len(items) > 30 {
		items = items[:30]
	}
	return items, nil
}
