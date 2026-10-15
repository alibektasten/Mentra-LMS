package usecases

import (
	"context"
	"fmt"
	"sort"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/google/uuid"
)

// NotificationUseCase builds a single, role-aware inbox from real course
// events. It deliberately exposes only courses the current member can access.
type NotificationUseCase struct {
	announcements repositories.AnnouncementRepository
	courses       repositories.CourseRepository
	groups        repositories.GroupRepository
	teachers      repositories.CourseTeacherRepository
	modules       repositories.ModuleRepository
	lessons       repositories.LessonRepository
	assignments   repositories.AssignmentRepository
	exams         repositories.ExamRepository
	courseGrades  repositories.CourseGradeRepository
}

func NewNotificationUseCase(
	announcements repositories.AnnouncementRepository,
	courses repositories.CourseRepository,
	groups repositories.GroupRepository,
	teachers repositories.CourseTeacherRepository,
	modules repositories.ModuleRepository,
	lessons repositories.LessonRepository,
	assignments repositories.AssignmentRepository,
	exams repositories.ExamRepository,
	courseGrades repositories.CourseGradeRepository,
) *NotificationUseCase {
	return &NotificationUseCase{
		announcements: announcements, courses: courses, groups: groups, teachers: teachers,
		modules: modules, lessons: lessons, assignments: assignments, exams: exams, courseGrades: courseGrades,
	}
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
		if err := uc.appendCourseEvents(ctx, &items, course, orgID); err != nil {
			return nil, err
		}
		if role == string(entities.RoleStudent) {
			if err := uc.appendGradePublication(ctx, &items, course, userID, orgID); err != nil {
				return nil, err
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if len(items) > 30 {
		items = items[:30]
	}
	return items, nil
}

func (uc *NotificationUseCase) appendCourseEvents(ctx context.Context, items *[]dto.NotificationDTO, course *entities.Course, orgID uuid.UUID) error {
	assignments, err := uc.assignments.FindByCourse(ctx, course.ID, orgID)
	if err != nil {
		return err
	}
	for _, assignment := range assignments {
		body := "A new assignment is available."
		if assignment.DueDate != nil {
			body = fmt.Sprintf("Due %s.", assignment.DueDate.Format("2 Jan 2006, 15:04"))
		}
		*items = append(*items, dto.NotificationDTO{
			ID: assignment.ID.String(), Type: "assignment", Title: "Assignment: " + assignment.Title, Body: body,
			CourseID: course.ID.String(), CourseTitle: course.Title, DueAt: assignment.DueDate, CreatedAt: assignment.CreatedAt,
		})
	}

	exams, err := uc.exams.FindByCourse(ctx, course.ID, orgID)
	if err != nil {
		return err
	}
	for _, exam := range exams {
		// Shell assessments are created with every course and should not notify
		// students until the teacher has scheduled or configured them.
		if exam.DueDate == nil && len(exam.Questions) == 0 {
			continue
		}
		body := "A new assessment is available."
		if exam.DueDate != nil {
			body = fmt.Sprintf("Scheduled for %s.", exam.DueDate.Format("2 Jan 2006, 15:04"))
		}
		*items = append(*items, dto.NotificationDTO{
			ID: exam.ID.String(), Type: "exam", Title: "Assessment: " + exam.Title, Body: body,
			CourseID: course.ID.String(), CourseTitle: course.Title, DueAt: exam.DueDate, CreatedAt: exam.CreatedAt,
		})
	}

	modules, err := uc.modules.FindByCourse(ctx, course.ID, orgID)
	if err != nil {
		return err
	}
	for _, module := range modules {
		lessons, err := uc.lessons.FindByModule(ctx, module.ID, orgID)
		if err != nil {
			return err
		}
		for _, lesson := range lessons {
			*items = append(*items, dto.NotificationDTO{
				ID: lesson.ID.String(), Type: "lesson", Title: "New material: " + lesson.Title, Body: "A new lesson is available.",
				CourseID: course.ID.String(), CourseTitle: course.Title, CreatedAt: lesson.CreatedAt,
			})
		}
	}
	return nil
}

func (uc *NotificationUseCase) appendGradePublication(ctx context.Context, items *[]dto.NotificationDTO, course *entities.Course, studentID, orgID uuid.UUID) error {
	grade, err := uc.courseGrades.Find(ctx, course.ID, studentID, orgID)
	if err != nil || grade == nil || grade.PublishedAt == nil {
		return err
	}
	*items = append(*items, dto.NotificationDTO{
		ID: "grade-" + grade.ID.String(), Type: "grade", Title: "Results published", Body: "Your final course result is now available.",
		CourseID: course.ID.String(), CourseTitle: course.Title, CreatedAt: *grade.PublishedAt,
	})
	return nil
}
