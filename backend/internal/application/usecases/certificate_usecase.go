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
	certificateRepo repositories.CertificateRepository
	groupRepo       repositories.GroupRepository
	courseRepo      repositories.CourseRepository
	moduleRepo      repositories.ModuleRepository
	lessonRepo      repositories.LessonRepository
	progressRepo    repositories.LessonProgressRepository
}

func NewCertificateUseCase(certificateRepo repositories.CertificateRepository, groupRepo repositories.GroupRepository, courseRepo repositories.CourseRepository, moduleRepo repositories.ModuleRepository, lessonRepo repositories.LessonRepository, progressRepo repositories.LessonProgressRepository) *CertificateUseCase {
	return &CertificateUseCase{certificateRepo: certificateRepo, groupRepo: groupRepo, courseRepo: courseRepo, moduleRepo: moduleRepo, lessonRepo: lessonRepo, progressRepo: progressRepo}
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
		course, err := uc.courseRepo.FindByID(ctx, certificate.CourseID, orgID)
		if err != nil {
			return nil, err
		}
		result = append(result, toCertificateDTO(&certificate, course.Title))
	}
	return result, nil
}

func (uc *CertificateUseCase) Verify(ctx context.Context, code string) (*dto.CertificateDTO, error) {
	certificate, err := uc.certificateRepo.FindByVerificationCode(ctx, code)
	if err != nil {
		return nil, err
	}
	course, err := uc.courseRepo.FindByID(ctx, certificate.CourseID, certificate.OrgID)
	if err != nil {
		return nil, err
	}
	result := toCertificateDTO(certificate, course.Title)
	return &result, nil
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

func toCertificateDTO(certificate *entities.Certificate, courseTitle string) dto.CertificateDTO {
	return dto.CertificateDTO{ID: certificate.ID.String(), CourseID: certificate.CourseID.String(), CourseTitle: courseTitle, VerificationCode: certificate.VerificationCode, IssuedAt: certificate.IssuedAt}
}
