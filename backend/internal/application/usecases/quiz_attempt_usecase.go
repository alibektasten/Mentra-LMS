package usecases

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/domain/services"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type QuizAttemptUseCase struct {
	attemptRepo repositories.QuizAttemptRepository
	quizRepo    repositories.QuizRepository
	lessonRepo  repositories.LessonRepository
	aiService   services.AIService
}

func NewQuizAttemptUseCase(
	attemptRepo repositories.QuizAttemptRepository,
	quizRepo repositories.QuizRepository,
	lessonRepo repositories.LessonRepository,
	aiService services.AIService,
) *QuizAttemptUseCase {
	return &QuizAttemptUseCase{
		attemptRepo: attemptRepo,
		quizRepo:    quizRepo,
		lessonRepo:  lessonRepo,
		aiService:   aiService,
	}
}

func (uc *QuizAttemptUseCase) Submit(ctx context.Context, quizID, studentID, orgID uuid.UUID, answers []dto.QuizAnswerInput) (*dto.QuizAttemptResultDTO, error) {
	quiz, err := uc.quizRepo.FindByID(ctx, quizID, orgID)
	if err != nil {
		return nil, err
	}

	// Check deadline
	if quiz.DueDate != nil && time.Now().After(*quiz.DueDate) && !quiz.AllowLateSubmission {
		return nil, apperrors.ValidationError("quiz deadline has passed")
	}

	// Check if already attempted
	existing, err := uc.attemptRepo.FindByQuizAndStudent(ctx, quizID, studentID)
	if err == nil && existing != nil {
		return toQuizAttemptResultDTO(existing), nil
	}

	// Build answer map for fast lookup
	answerMap := make(map[string]string)
	for _, a := range answers {
		answerMap[a.QuestionID] = a.AnswerID
	}

	// Auto-grade
	correct := 0
	for _, q := range quiz.Questions {
		selectedAnswerID := answerMap[q.ID.String()]
		for _, a := range q.Answers {
			if a.ID.String() == selectedAnswerID && a.IsCorrect {
				correct++
				break
			}
		}
	}

	maxScore := len(quiz.Questions)
	attemptAnswers := make([]entities.QuizAttemptAnswer, len(answers))
	for i, a := range answers {
		attemptAnswers[i] = entities.QuizAttemptAnswer{
			QuestionID: a.QuestionID,
			AnswerID:   a.AnswerID,
		}
	}

	attempt := &entities.QuizAttempt{
		ID:          uuid.New(),
		QuizID:      quizID,
		StudentID:   studentID,
		OrgID:       orgID,
		Score:       correct,
		MaxScore:    maxScore,
		Answers:     attemptAnswers,
		SubmittedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.attemptRepo.Create(ctx, attempt); err != nil {
		return nil, err
	}
	return toQuizAttemptResultDTO(attempt), nil
}

func (uc *QuizAttemptUseCase) GetMyAttempt(ctx context.Context, quizID, studentID uuid.UUID) (*dto.QuizAttemptResultDTO, error) {
	attempt, err := uc.attemptRepo.FindByQuizAndStudent(ctx, quizID, studentID)
	if err != nil {
		return nil, err
	}
	return toQuizAttemptResultDTO(attempt), nil
}

func (uc *QuizAttemptUseCase) GetRemediation(ctx context.Context, quizID, studentID, orgID uuid.UUID) (*dto.RemediationDTO, error) {
	attempt, err := uc.attemptRepo.FindByQuizAndStudent(ctx, quizID, studentID)
	if err != nil {
		return nil, err
	}

	quiz, err := uc.quizRepo.FindByID(ctx, quizID, orgID)
	if err != nil {
		return nil, err
	}

	lesson, err := uc.lessonRepo.FindByID(ctx, quiz.LessonID, orgID)
	if err != nil {
		return nil, err
	}

	pct := 0.0
	if attempt.MaxScore > 0 {
		pct = float64(attempt.Score) / float64(attempt.MaxScore) * 100
	}

	answerMap := make(map[string]string, len(attempt.Answers))
	for _, a := range attempt.Answers {
		answerMap[a.QuestionID] = a.AnswerID
	}

	var wrongQuestions []string
	for _, q := range quiz.Questions {
		selectedID := answerMap[q.ID.String()]
		isCorrect := false
		for _, a := range q.Answers {
			if a.ID.String() == selectedID && a.IsCorrect {
				isCorrect = true
				break
			}
		}
		if !isCorrect {
			wrongQuestions = append(wrongQuestions, q.Question)
		}
	}

	remediation := "You answered all questions correctly. Great job!"
	if len(wrongQuestions) > 0 {
		remediation, err = uc.aiService.GenerateRemediation(ctx, lesson.Content, wrongQuestions)
		if err != nil {
			if isAIBudgetError(err) {
				return nil, err
			}
			remediation = "Review the lesson material again, paying close attention to the topics you found challenging."
		}
	}

	return &dto.RemediationDTO{
		QuizID:      quizID.String(),
		LessonTitle: lesson.Title,
		Score:       attempt.Score,
		MaxScore:    attempt.MaxScore,
		Percentage:  pct,
		Remediation: remediation,
		WeakTopics:  wrongQuestions,
	}, nil
}

func toQuizAttemptResultDTO(a *entities.QuizAttempt) *dto.QuizAttemptResultDTO {
	var pct float64
	if a.MaxScore > 0 {
		pct = float64(a.Score) / float64(a.MaxScore) * 100
	}
	return &dto.QuizAttemptResultDTO{
		ID:          a.ID.String(),
		QuizID:      a.QuizID.String(),
		Score:       a.Score,
		MaxScore:    a.MaxScore,
		Percentage:  pct,
		SubmittedAt: a.SubmittedAt,
	}
}
