package usecases

import (
	"context"
	"sort"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/google/uuid"
)

type GradeUseCase struct {
	assignmentRepo  repositories.AssignmentRepository
	attemptRepo     repositories.QuizAttemptRepository
	quizRepo        repositories.QuizRepository
	memberRepo      repositories.MembershipRepository
	examRepo        repositories.ExamRepository
	examAttemptRepo repositories.ExamAttemptRepository
}

func NewGradeUseCase(
	assignmentRepo repositories.AssignmentRepository,
	attemptRepo repositories.QuizAttemptRepository,
	quizRepo repositories.QuizRepository,
	memberRepo repositories.MembershipRepository,
	extras ...interface{},
) *GradeUseCase {
	var examRepo repositories.ExamRepository
	var examAttemptRepo repositories.ExamAttemptRepository
	if len(extras) > 0 {
		examRepo, _ = extras[0].(repositories.ExamRepository)
	}
	if len(extras) > 1 {
		examAttemptRepo, _ = extras[1].(repositories.ExamAttemptRepository)
	}
	return &GradeUseCase{
		assignmentRepo:  assignmentRepo,
		attemptRepo:     attemptRepo,
		quizRepo:        quizRepo,
		memberRepo:      memberRepo,
		examRepo:        examRepo,
		examAttemptRepo: examAttemptRepo,
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

	assignmentsPercent := percentagePointer(totalEarned, totalPossible)
	var midtermPercent, endtermPercent, finalExamPercent *float64
	if uc.examRepo != nil && uc.examAttemptRepo != nil {
		exams, err := uc.examRepo.FindByCourse(ctx, courseID, orgID)
		if err != nil {
			return nil, err
		}
		for _, exam := range exams {
			if exam.AssessmentType != "midterm" && exam.AssessmentType != "endterm" && exam.AssessmentType != "final" {
				continue
			}
			percent, score, submittedAt, gradedAt, err := uc.examPercentage(ctx, exam.ID, studentID, exam.MCQPoints+exam.FilePoints)
			if err != nil {
				return nil, err
			}
			item := dto.GradeItemDTO{ItemID: exam.ID.String(), ItemType: exam.AssessmentType, Title: exam.Title, MaxPoints: exam.MCQPoints + exam.FilePoints, Score: score, Status: "not_submitted", SubmittedAt: submittedAt, GradedAt: gradedAt}
			if submittedAt != nil {
				item.Status = "submitted"
			}
			if score != nil {
				item.Status = "graded"
			}
			items = append(items, item)
			switch exam.AssessmentType {
			case "midterm":
				midtermPercent = percent
			case "endterm":
				endtermPercent = percent
			case "final":
				finalExamPercent = percent
			}
		}
	}

	midtermRegister := combinedRegister(assignmentsPercent, midtermPercent)
	endtermRegister := combinedRegister(assignmentsPercent, endtermPercent)
	finalGrade := weightedGrade(assignmentsPercent, midtermPercent, endtermPercent, finalExamPercent)
	isReady := assignmentsPercent != nil && midtermPercent != nil && endtermPercent != nil && finalExamPercent != nil
	percentage := 0.0
	if assignmentsPercent != nil {
		percentage = *assignmentsPercent
	}
	if uc.examRepo != nil && uc.examAttemptRepo != nil {
		percentage = finalGrade
	}

	return &dto.StudentGradeDTO{
		StudentID:          studentID.String(),
		Items:              items,
		TotalEarned:        totalEarned,
		TotalPossible:      totalPossible,
		Percentage:         percentage,
		AssignmentsPercent: assignmentsPercent,
		MidtermPercent:     midtermPercent,
		EndtermPercent:     endtermPercent,
		FinalExamPercent:   finalExamPercent,
		MidtermRegister:    midtermRegister,
		EndtermRegister:    endtermRegister,
		FinalGrade:         finalGrade,
		IsFinalGradeReady:  isReady,
	}, nil
}

func percentagePointer(earned, possible int) *float64 {
	if possible <= 0 {
		return nil
	}
	value := float64(earned) / float64(possible) * 100
	return &value
}

func combinedRegister(assignments, exam *float64) *float64 {
	if assignments == nil || exam == nil {
		return nil
	}
	value := (*assignments + *exam) / 2
	return &value
}

func weightedGrade(assignments, midterm, endterm, finalExam *float64) float64 {
	value := 0.0
	if assignments != nil {
		value += *assignments * 0.25
	}
	if midterm != nil {
		value += *midterm * 0.125
	}
	if endterm != nil {
		value += *endterm * 0.125
	}
	if finalExam != nil {
		value += *finalExam * 0.5
	}
	return value
}

func (uc *GradeUseCase) examPercentage(ctx context.Context, examID, studentID uuid.UUID, maxPoints int) (*float64, *int, *time.Time, *time.Time, error) {
	attempts, err := uc.examAttemptRepo.FindByExamAndStudent(ctx, examID, studentID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var bestScore *int
	var submittedAt, gradedAt *time.Time
	for _, attempt := range attempts {
		if attempt.SubmittedAt != nil && (submittedAt == nil || attempt.SubmittedAt.After(*submittedAt)) {
			submittedAt = attempt.SubmittedAt
		}
		if attempt.TotalScore != nil && (bestScore == nil || *attempt.TotalScore > *bestScore) {
			bestScore = attempt.TotalScore
			gradedAt = attempt.GradedAt
		}
	}
	if bestScore == nil || maxPoints <= 0 {
		return nil, bestScore, submittedAt, gradedAt, nil
	}
	percent := float64(*bestScore) / float64(maxPoints) * 100
	return &percent, bestScore, submittedAt, gradedAt, nil
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

	// Scheduled weighted assessments belong alongside regular assignments in a
	// student's deadline list. This makes the course timeline reflect the full
	// grading plan rather than only module assignments.
	if uc.examRepo != nil && uc.examAttemptRepo != nil {
		exams, err := uc.examRepo.FindByCourse(ctx, courseID, orgID)
		if err != nil {
			return nil, err
		}
		for _, exam := range exams {
			if exam.DueDate == nil || exam.DueDate.Before(now) || !isScheduledAssessment(exam.AssessmentType) {
				continue
			}

			attempts, err := uc.examAttemptRepo.FindByExamAndStudent(ctx, exam.ID, studentID)
			if err != nil {
				return nil, err
			}
			submitted := false
			for _, attempt := range attempts {
				if attempt.SubmittedAt != nil || attempt.Status == "submitted" {
					submitted = true
					break
				}
			}

			deadlines = append(deadlines, dto.DeadlineItemDTO{
				ItemID:    exam.ID.String(),
				ItemType:  exam.AssessmentType,
				Title:     exam.Title,
				DueDate:   *exam.DueDate,
				Submitted: submitted,
			})
		}
	}

	sort.Slice(deadlines, func(i, j int) bool {
		return deadlines[i].DueDate.Before(deadlines[j].DueDate)
	})

	return deadlines, nil
}

func isScheduledAssessment(value string) bool {
	return value == "midterm" || value == "endterm" || value == "final"
}
