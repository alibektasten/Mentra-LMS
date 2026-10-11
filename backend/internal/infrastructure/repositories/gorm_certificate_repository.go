package repositories

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/database"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMCertificateRepository struct{ db *gorm.DB }

func NewGORMCertificateRepository(db *gorm.DB) *GORMCertificateRepository {
	return &GORMCertificateRepository{db: db}
}

func (r *GORMCertificateRepository) Create(ctx context.Context, c *entities.Certificate) error {
	return r.db.WithContext(ctx).Create(&database.CertificateModel{ID: c.ID.String(), UserID: c.UserID.String(), CourseID: c.CourseID.String(), OrgID: c.OrgID.String(), VerificationCode: c.VerificationCode, IssuedAt: c.IssuedAt}).Error
}

func (r *GORMCertificateRepository) FindByUserAndCourse(ctx context.Context, userID, courseID, orgID uuid.UUID) (*entities.Certificate, error) {
	var model database.CertificateModel
	err := r.db.WithContext(ctx).First(&model, "user_id = ? AND course_id = ? AND org_id = ?", userID.String(), courseID.String(), orgID.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toCertificateEntity(&model), nil
}

func (r *GORMCertificateRepository) ListByUser(ctx context.Context, userID, orgID uuid.UUID) ([]entities.Certificate, error) {
	var models []database.CertificateModel
	if err := r.db.WithContext(ctx).Where("user_id = ? AND org_id = ?", userID.String(), orgID.String()).Order("issued_at DESC").Find(&models).Error; err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]entities.Certificate, len(models))
	for i := range models {
		result[i] = *toCertificateEntity(&models[i])
	}
	return result, nil
}

func (r *GORMCertificateRepository) FindByVerificationCode(ctx context.Context, code string) (*entities.Certificate, error) {
	var model database.CertificateModel
	err := r.db.WithContext(ctx).First(&model, "verification_code = ?", code).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("certificate", code)
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toCertificateEntity(&model), nil
}

func toCertificateEntity(m *database.CertificateModel) *entities.Certificate {
	id, _ := uuid.Parse(m.ID)
	userID, _ := uuid.Parse(m.UserID)
	courseID, _ := uuid.Parse(m.CourseID)
	orgID, _ := uuid.Parse(m.OrgID)
	return &entities.Certificate{ID: id, UserID: userID, CourseID: courseID, OrgID: orgID, VerificationCode: m.VerificationCode, IssuedAt: m.IssuedAt}
}
