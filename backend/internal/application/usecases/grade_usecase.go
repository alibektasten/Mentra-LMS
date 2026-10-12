package usecases

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/google/uuid"
)

type GradeUseCase struct {
	assignmentRepo repositories.AssignmentRepository
	attemptRepo    repositories.QuizAttemptRepository
	quizRepo       repositories.QuizRepository
	memberRepo     repositories.MembershipRepository
}

func NewGradeUseCase(
	assignmentRepo repositories.AssignmentRepository,
	attemptRepo repositories.QuizAttemptRepository,
	quizRepo repositories.QuizRepository,
	memberRepo repositories.MembershipRepository,
) *GradeUseCase {
	return &GradeUseCase{
		assignmentRepo: assignmentRepo,
		attemptRepo:    attemptRepo,
		quizRepo:       quizRepo,
		memberRepo:     memberRepo,
	}
}

func (uc *GradeUseCase) GetMyGrades(ctx context.Context, courseID, studentID, orgID uuid.UUID) (*dto.StudentGradeDTO, error) {
	assignments, err := uc.assignmentRepo.FindByCourse(ctx, courseID, orgID)
	if err != nil {
		return nil, err
	}

	var items []dto.GradeItemDTO
	totalEarned := 0
	totalPossible := 0

	for _, a := range assignments {
		item := dto.GradeItemDTO{
			ItemID:    a.ID.String(),
			ItemType:  "assignment",
			Title:     a.Title,
			MaxPoints: a.MaxPoints,
			Status:    "not_submitted",
		}
		totalPossible += a.MaxPoints
		sub, err := uc.assignmentRepo.FindSubmission(ctx, a.ID, studentID)
		if err == nil && sub != nil {
			item.SubmittedAt = &sub.SubmittedAt
			item.Feedback = sub.Feedback
			if sub.Score != nil {
				item.Score = sub.Score
				item.GradedAt = sub.GradedAt
				item.Status = "graded"
				totalEarned += *sub.Score
			} else {
				item.Status = "submitted"
			}
		}
		items = append(items, item)
	}

	var pct float64
	if totalPossible > 0 {
		pct = float64(totalEarned) / float64(totalPossible) * 100
	}

	return &dto.StudentGradeDTO{
		StudentID:     studentID.String(),
		Items:         items,
		TotalEarned:   totalEarned,
		TotalPossible: totalPossible,
		Percentage:    pct,
	}, nil
}

func (uc *GradeUseCase) GetUpcomingDeadlines(ctx context.Context, courseID, studentID, orgID uuid.UUID) ([]dto.DeadlineItemDTO, error) {
	assignments, err := uc.assignmentRepo.FindByCourse(ctx, courseID, orgID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var deadlines []dto.DeadlineItemDTO

	for _, a := range assignments {
		if a.DueDate == nil || a.DueDate.Before(now) {
			continue
		}
		submitted := false
		sub, err := uc.assignmentRepo.FindSubmission(ctx, a.ID, studentID)
		if err == nil && sub != nil {
			submitted = true
		}
		deadlines = append(deadlines, dto.DeadlineItemDTO{
			ItemID:    a.ID.String(),
			ItemType:  "assignment",
			Title:     a.Title,
			DueDate:   *a.DueDate,
			Submitted: submitted,
		})
	}

	return deadlines, nil
}
