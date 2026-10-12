package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/google/uuid"
)

type CertificateUseCase struct {
	certificateRepo   repositories.CertificateRepository
	groupRepo         repositories.GroupRepository
	courseRepo        repositories.CourseRepository
	moduleRepo        repositories.ModuleRepository
	lessonRepo        repositories.LessonRepository
	progressRepo      repositories.LessonProgressRepository
	userRepo          repositories.UserRepository
	organizationRepo  repositories.OrganizationRepository
	courseTeacherRepo repositories.CourseTeacherRepository
}

func NewCertificateUseCase(certificateRepo repositories.CertificateRepository, groupRepo repositories.GroupRepository, courseRepo repositories.CourseRepository, moduleRepo repositories.ModuleRepository, lessonRepo repositories.LessonRepository, progressRepo repositories.LessonProgressRepository, userRepo repositories.UserRepository, organizationRepo repositories.OrganizationRepository, courseTeacherRepo repositories.CourseTeacherRepository) *CertificateUseCase {
	return &CertificateUseCase{certificateRepo: certificateRepo, groupRepo: groupRepo, courseRepo: courseRepo, moduleRepo: moduleRepo, lessonRepo: lessonRepo, progressRepo: progressRepo, userRepo: userRepo, organizationRepo: organizationRepo, courseTeacherRepo: courseTeacherRepo}
}

// ListForStudent issues any newly earned certificates before returning them.
// A course qualifies only when it has lessons and every lesson is completed.
func (uc *CertificateUseCase) ListForStudent(ctx context.Context, userID, orgID uuid.UUID) ([]dto.CertificateDTO, error) {
	courseIDs, err := uc.groupRepo.FindCourseIDsByStudent(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	for _, courseID := range courseIDs {
		if err := uc.issueIfEligible(ctx, userID, courseID, orgID); err != nil {
			return nil, err
		}
	}
	certificates, err := uc.certificateRepo.ListByUser(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CertificateDTO, 0, len(certificates))
	for _, certificate := range certificates {
		certificateDTO, err := uc.toCertificateDTO(ctx, &certificate)
		if err != nil {
			return nil, err
		}
		result = append(result, certificateDTO)
	}
	return result, nil
}

func (uc *CertificateUseCase) Verify(ctx context.Context, code string) (*dto.CertificateDTO, error) {
	certificate, err := uc.certificateRepo.FindByVerificationCode(ctx, code)
	if err != nil {
		return nil, err
	}
	result, err := uc.toCertificateDTO(ctx, certificate)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (uc *CertificateUseCase) toCertificateDTO(ctx context.Context, certificate *entities.Certificate) (dto.CertificateDTO, error) {
	course, err := uc.courseRepo.FindByID(ctx, certificate.CourseID, certificate.OrgID)
	if err != nil {
		return dto.CertificateDTO{}, err
	}
	student, err := uc.userRepo.FindByID(ctx, certificate.UserID)
	if err != nil {
		return dto.CertificateDTO{}, err
	}
	organization, err := uc.organizationRepo.FindByID(ctx, certificate.OrgID)
	if err != nil {
		return dto.CertificateDTO{}, err
	}

	teacherName := "Mentra LMS Academic Office"
	teachers, err := uc.courseTeacherRepo.ListByCourse(ctx, certificate.CourseID, certificate.OrgID)
	if err != nil {
		return dto.CertificateDTO{}, err
	}
	if len(teachers) > 0 {
		teacher, err := uc.userRepo.FindByID(ctx, teachers[0].TeacherID)
		if err != nil {
			return dto.CertificateDTO{}, err
		}
		teacherName = teacher.Name
	}

	return dto.CertificateDTO{
		ID: certificate.ID.String(), CourseID: certificate.CourseID.String(), CourseTitle: course.Title,
		StudentName: student.Name, TeacherName: teacherName, OrganizationName: organization.Name,
		VerificationCode: certificate.VerificationCode, IssuedAt: certificate.IssuedAt,
	}, nil
}

func (uc *CertificateUseCase) issueIfEligible(ctx context.Context, userID, courseID, orgID uuid.UUID) error {
	existing, err := uc.certificateRepo.FindByUserAndCourse(ctx, userID, courseID, orgID)
	if err != nil || existing != nil {
		return err
	}
	course, err := uc.courseRepo.FindByID(ctx, courseID, orgID)
	if err != nil {
		return err
	}
	if course.Status != entities.StatusPublished {
		return nil
	}
	modules, err := uc.moduleRepo.FindByCourse(ctx, courseID, orgID)
	if err != nil {
		return err
	}
	lessonIDs := map[uuid.UUID]bool{}
	for _, module := range modules {
		lessons, err := uc.lessonRepo.FindByModule(ctx, module.ID, orgID)
		if err != nil {
			return err
		}
		for _, lesson := range lessons {
			lessonIDs[lesson.ID] = true
		}
	}
	if len(lessonIDs) == 0 {
		return nil
	}
	progress, err := uc.progressRepo.FindByUser(ctx, userID, orgID)
	if err != nil {
		return err
	}
	completed := map[uuid.UUID]bool{}
	for _, item := range progress {
		if item.CompletedAt != nil {
			completed[item.LessonID] = true
		}
	}
	for lessonID := range lessonIDs {
		if !completed[lessonID] {
			return nil
		}
	}
	now := time.Now()
	certificate := &entities.Certificate{ID: uuid.New(), UserID: userID, CourseID: courseID, OrgID: orgID, VerificationCode: fmt.Sprintf("MNT-%s", uuid.NewString()[:12]), IssuedAt: now}
	return uc.certificateRepo.Create(ctx, certificate)
}
