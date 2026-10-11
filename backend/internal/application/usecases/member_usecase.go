package usecases

import (
	"context"
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type MemberUseCase struct {
	userRepo   repositories.UserRepository
	memberRepo repositories.MembershipRepository
}

func NewMemberUseCase(userRepo repositories.UserRepository, memberRepo repositories.MembershipRepository) *MemberUseCase {
	return &MemberUseCase{userRepo: userRepo, memberRepo: memberRepo}
}

func (uc *MemberUseCase) ListMembers(ctx context.Context, orgID uuid.UUID, role string, page, pageSize int) ([]dto.MemberDTO, int64, error) {
	memberships, total, err := uc.memberRepo.FindByOrgFiltered(ctx, orgID, role, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(memberships) == 0 {
		return []dto.MemberDTO{}, 0, nil
	}
	userIDs := make([]uuid.UUID, len(memberships))
	for i, m := range memberships {
		userIDs[i] = m.UserID
	}
	users, err := uc.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}
	userMap := make(map[string]entities.User, len(users))
	for _, u := range users {
		userMap[u.ID.String()] = u
	}
	result := make([]dto.MemberDTO, len(memberships))
	for i, m := range memberships {
		u := userMap[m.UserID.String()]
		result[i] = dto.MemberDTO{
			ID:       m.ID.String(),
			UserID:   m.UserID.String(),
			OrgID:    m.OrgID.String(),
			Role:     string(m.Role),
			Name:     u.Name,
			Email:    u.Email,
			JoinedAt: m.CreatedAt,
		}
	}
	return result, total, nil
}

func (uc *MemberUseCase) InviteMember(ctx context.Context, orgID uuid.UUID, req dto.InviteMemberRequest) (*dto.MemberDTO, error) {
	existing, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	var user *entities.User
	now := time.Now()
	if existing != nil {
		user = existing
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperrors.InternalError("failed to hash password")
		}
		user = &entities.User{
			ID:           uuid.New(),
			Email:        req.Email,
			PasswordHash: string(hash),
			Name:         req.Name,
			Role:         req.Role,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := uc.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
	}

	existingMembership, _ := uc.memberRepo.FindByUserAndOrg(ctx, user.ID, orgID)
	if existingMembership != nil {
		return nil, apperrors.ConflictError("user is already a member of this organization")
	}

	m := &entities.Membership{
		ID:        uuid.New(),
		UserID:    user.ID,
		OrgID:     orgID,
		Role:      entities.Role(req.Role),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.memberRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return &dto.MemberDTO{
		ID:       m.ID.String(),
		UserID:   user.ID.String(),
		OrgID:    orgID.String(),
		Role:     string(m.Role),
		Name:     user.Name,
		Email:    user.Email,
		JoinedAt: m.CreatedAt,
	}, nil
}

func (uc *MemberUseCase) BulkImportCSV(ctx context.Context, orgID uuid.UUID, r io.Reader) (*dto.CSVImportResultDTO, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, apperrors.ValidationError("invalid CSV: cannot read header")
	}
	// Normalize header
	colIndex := map[string]int{"name": -1, "email": -1, "role": -1}
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		if _, ok := colIndex[key]; ok {
			colIndex[key] = i
		}
	}
	for col, idx := range colIndex {
		if idx < 0 {
			return nil, apperrors.ValidationError(fmt.Sprintf("missing required CSV column: %s", col))
		}
	}

	result := &dto.CSVImportResultDTO{
		Imported: []dto.CSVImportedUser{},
		Errors:   []dto.CSVRowError{},
	}
	rowNum := 1
	now := time.Now()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			result.Errors = append(result.Errors, dto.CSVRowError{Row: rowNum, Error: "invalid row"})
			continue
		}

		name := strings.TrimSpace(record[colIndex["name"]])
		email := strings.TrimSpace(record[colIndex["email"]])
		role := strings.TrimSpace(record[colIndex["role"]])

		if name == "" || email == "" || role == "" {
			result.Errors = append(result.Errors, dto.CSVRowError{Row: rowNum, Email: email, Error: "name, email, and role are required"})
			continue
		}
		if role != "admin" && role != "teacher" && role != "student" {
			result.Errors = append(result.Errors, dto.CSVRowError{Row: rowNum, Email: email, Error: "role must be admin, teacher, or student"})
			continue
		}

		var user *entities.User
		var plainPassword string
		existing, _ := uc.userRepo.FindByEmail(ctx, email)
		if existing != nil {
			user = existing
		} else {
			plainPassword = randomPassword(12)
			hash, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
			user = &entities.User{
				ID:           uuid.New(),
				Email:        email,
				PasswordHash: string(hash),
				Name:         name,
				Role:         role,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := uc.userRepo.Create(ctx, user); err != nil {
				result.Errors = append(result.Errors, dto.CSVRowError{Row: rowNum, Email: email, Error: "failed to create user"})
				continue
			}
		}

		existingMembership, _ := uc.memberRepo.FindByUserAndOrg(ctx, user.ID, orgID)
		if existingMembership != nil {
			result.Errors = append(result.Errors, dto.CSVRowError{Row: rowNum, Email: email, Error: "already a member"})
			continue
		}

		m := &entities.Membership{
			ID:        uuid.New(),
			UserID:    user.ID,
			OrgID:     orgID,
			Role:      entities.Role(role),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := uc.memberRepo.Create(ctx, m); err != nil {
			result.Errors = append(result.Errors, dto.CSVRowError{Row: rowNum, Email: email, Error: "failed to create membership"})
			continue
		}

		result.Imported = append(result.Imported, dto.CSVImportedUser{
			Name:     user.Name,
			Email:    user.Email,
			Role:     role,
			Password: plainPassword,
		})
	}
	return result, nil
}

func (uc *MemberUseCase) RemoveMember(ctx context.Context, membershipID, orgID uuid.UUID) error {
	membership, err := uc.memberRepo.FindByID(ctx, membershipID, orgID)
	if err != nil {
		return err
	}
	if membership.Role == entities.RoleAdmin {
		_, admins, err := uc.memberRepo.FindByOrgFiltered(ctx, orgID, string(entities.RoleAdmin), 1, 1)
		if err != nil {
			return err
		}
		if admins <= 1 {
			return apperrors.ValidationError("an organization must retain at least one administrator")
		}
	}
	return uc.memberRepo.Delete(ctx, membershipID, orgID)
}

func (uc *MemberUseCase) UpdateRole(ctx context.Context, membershipID, orgID uuid.UUID, req dto.UpdateMemberRoleRequest) (*dto.MemberDTO, error) {
	m, err := uc.memberRepo.FindByID(ctx, membershipID, orgID)
	if err != nil {
		return nil, err
	}
	newRole := entities.Role(req.Role)
	if m.Role == entities.RoleAdmin && newRole != entities.RoleAdmin {
		_, admins, err := uc.memberRepo.FindByOrgFiltered(ctx, orgID, string(entities.RoleAdmin), 1, 1)
		if err != nil {
			return nil, err
		}
		if admins <= 1 {
			return nil, apperrors.ValidationError("an organization must retain at least one administrator")
		}
	}
	m.Role = newRole
	m.UpdatedAt = time.Now()
	if err := uc.memberRepo.Update(ctx, m); err != nil {
		return nil, err
	}
	user, err := uc.userRepo.FindByID(ctx, m.UserID)
	if err != nil {
		return nil, err
	}
	return &dto.MemberDTO{
		ID:       m.ID.String(),
		UserID:   m.UserID.String(),
		OrgID:    m.OrgID.String(),
		Role:     string(m.Role),
		Name:     user.Name,
		Email:    user.Email,
		JoinedAt: m.CreatedAt,
	}, nil
}

const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomPassword(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(passwordChars))))
		b[i] = passwordChars[n.Int64()]
	}
	return string(b)
}
