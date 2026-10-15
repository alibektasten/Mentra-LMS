package usecases

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/google/uuid"
)

type CourseUseCase struct {
	courseRepo repositories.CourseRepository
	memberRepo repositories.MembershipRepository
	groupRepo  repositories.GroupRepository
	ctRepo     repositories.CourseTeacherRepository
	examRepo   repositories.ExamRepository
}

func NewCourseUseCase(
	courseRepo repositories.CourseRepository,
	memberRepo repositories.MembershipRepository,
	groupRepo repositories.GroupRepository,
	ctRepo repositories.CourseTeacherRepository,
	extras ...interface{},
) *CourseUseCase {
	var examRepo repositories.ExamRepository
	if len(extras) > 0 {
		examRepo, _ = extras[0].(repositories.ExamRepository)
	}
	return &CourseUseCase{courseRepo: courseRepo, memberRepo: memberRepo, groupRepo: groupRepo, ctRepo: ctRepo, examRepo: examRepo}
}

func (uc *CourseUseCase) CreateCourse(ctx context.Context, req dto.CreateCourseRequest, creatorID, orgID uuid.UUID) (*dto.CourseDTO, error) {
	status := entities.StatusDraft
	if req.Status == string(entities.StatusPublished) {
		status = entities.StatusPublished
	}
	course := &entities.Course{
		ID:          uuid.New(),
		OrgID:       orgID,
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		CreatedBy:   creatorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	// A teacher must immediately see and be able to manage the course they
	// create. Admin-created courses also keep their creator as a collaborator.
	teacher := &entities.CourseTeacher{
		ID:         uuid.New(),
		CourseID:   course.ID,
		TeacherID:  creatorID,
		OrgID:      orgID,
		AssignedAt: time.Now(),
	}
	var assessments []*entities.Exam
	if uc.examRepo != nil {
		assessments = buildAssessmentPlan(course)
	}
	if setupRepo, ok := uc.courseRepo.(repositories.CourseSetupRepository); ok {
		if err := setupRepo.CreateWithAcademicSetup(ctx, course, teacher, assessments); err != nil {
			return nil, err
		}
		return toCourseDTO(course), nil
	}

	// Fallback for non-database adapters. Production GORM repositories use the
	// transaction above; this path keeps legacy in-memory/test adapters usable.
	if err := uc.courseRepo.Create(ctx, course); err != nil {
		return nil, err
	}
	if err := uc.ctRepo.Add(ctx, teacher); err != nil {
		return nil, err
	}
	for _, assessment := range assessments {
		if err := uc.examRepo.Create(ctx, assessment); err != nil {
			return nil, err
		}
	}
	return toCourseDTO(course), nil
}

// createAssessmentPlan provisions the required assessments as editable
// placeholders. They remain unavailable to students until the teacher assigns
// a due date, so a newly created course has a complete grading plan without
// exposing unfinished exams.
func buildAssessmentPlan(course *entities.Course) []*entities.Exam {
	now := time.Now()
	plan := []struct {
		assessmentType string
		title          string
		description    string
	}{
		{"midterm", "Midterm examination", "Required midterm assessment. Set a date and add instructions or questions before scheduling."},
		{"endterm", "Endterm examination", "Required endterm assessment. Set a date and add instructions or questions before scheduling."},
		{"final", "Final examination", "Required final assessment. Set a date and add instructions or questions before scheduling."},
	}
	exams := make([]*entities.Exam, 0, len(plan))
	for _, item := range plan {
		exam := &entities.Exam{
			ID:              uuid.New(),
			CourseID:        course.ID,
			OrgID:           course.OrgID,
			AssessmentType:  item.assessmentType,
			Title:           item.title,
			Description:     item.description,
			DurationMinutes: 120,
			MaxAttempts:     1,
			FileEnabled:     true,
			FilePoints:      100,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		exams = append(exams, exam)
	}
	return exams
}

func (uc *CourseUseCase) GetCourse(ctx context.Context, id, orgID uuid.UUID) (*dto.CourseDTO, error) {
	course, err := uc.courseRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	return toCourseDTO(course), nil
}

func (uc *CourseUseCase) ListCourses(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]dto.CourseDTO, int64, error) {
	courses, total, err := uc.courseRepo.FindByOrg(ctx, orgID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.CourseDTO, len(courses))
	for i, c := range courses {
		result[i] = *toCourseDTO(&c)
	}
	return result, total, nil
}

func (uc *CourseUseCase) ListCoursesForRole(ctx context.Context, userID, orgID uuid.UUID, role string, page, pageSize int) ([]dto.CourseDTO, int64, error) {
	switch role {
	case string(entities.RoleAdmin), string(entities.RoleSuperAdmin):
		return uc.ListCourses(ctx, orgID, page, pageSize)
	case string(entities.RoleTeacher):
		ids, err := uc.ctRepo.FindCourseIDsByTeacher(ctx, userID, orgID)
		if err != nil {
			return nil, 0, err
		}
		courses, total, err := uc.courseRepo.FindByIDs(ctx, ids, orgID, page, pageSize)
		if err != nil {
			return nil, 0, err
		}
		result := make([]dto.CourseDTO, len(courses))
		for i, c := range courses {
			result[i] = *toCourseDTO(&c)
		}
		return result, total, nil
	case string(entities.RoleStudent):
		ids, err := uc.groupRepo.FindCourseIDsByStudent(ctx, userID, orgID)
		if err != nil {
			return nil, 0, err
		}
		courses, total, err := uc.courseRepo.FindPublishedByIDs(ctx, ids, orgID, page, pageSize)
		if err != nil {
			return nil, 0, err
		}
		result := make([]dto.CourseDTO, len(courses))
		for i, c := range courses {
			result[i] = *toCourseDTO(&c)
		}
		return result, total, nil
	default:
		return uc.ListCourses(ctx, orgID, page, pageSize)
	}
}

func (uc *CourseUseCase) UpdateCourse(ctx context.Context, id, orgID uuid.UUID, req dto.UpdateCourseRequest) (*dto.CourseDTO, error) {
	course, err := uc.courseRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		course.Title = req.Title
	}
	if req.Description != "" {
		course.Description = req.Description
	}
	if req.Status == string(entities.StatusDraft) || req.Status == string(entities.StatusPublished) || req.Status == string(entities.StatusArchived) {
		course.Status = entities.CourseStatus(req.Status)
	}
	course.UpdatedAt = time.Now()
	if err := uc.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}
	return toCourseDTO(course), nil
}

func (uc *CourseUseCase) DeleteCourse(ctx context.Context, id, orgID uuid.UUID) error {
	return uc.courseRepo.Delete(ctx, id, orgID)
}

func (uc *CourseUseCase) PublishCourse(ctx context.Context, id, orgID uuid.UUID) error {
	course, err := uc.courseRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return err
	}
	if err := course.Publish(); err != nil {
		return err
	}
	course.UpdatedAt = time.Now()
	return uc.courseRepo.Update(ctx, course)
}

func toCourseDTO(c *entities.Course) *dto.CourseDTO {
	return &dto.CourseDTO{
		ID:          c.ID.String(),
		OrgID:       c.OrgID.String(),
		Title:       c.Title,
		Description: c.Description,
		Status:      string(c.Status),
		CreatedBy:   c.CreatedBy.String(),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toModuleDTO(m *entities.Module) *dto.ModuleDTO {
	return &dto.ModuleDTO{
		ID:        m.ID.String(),
		CourseID:  m.CourseID.String(),
		OrgID:     m.OrgID.String(),
		Title:     m.Title,
		Position:  m.Position,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toLessonDTO(l *entities.Lesson) *dto.LessonDTO {
	return &dto.LessonDTO{
		ID:        l.ID.String(),
		ModuleID:  l.ModuleID.String(),
		OrgID:     l.OrgID.String(),
		Title:     l.Title,
		Content:   l.Content,
		Type:      string(l.Type),
		VideoURL:  l.VideoURL,
		LinkURL:   l.LinkURL,
		FileURL:   l.FileURL,
		Position:  l.Position,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

// exported for use in other packages
var ToCourseDTO = toCourseDTO
var ToModuleDTO = toModuleDTO
var ToLessonDTO = toLessonDTO
