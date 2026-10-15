package usecases

import (
	"context"
	"strings"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/domain/services"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type AIUseCase struct {
	lessonRepo     repositories.LessonRepository
	quizRepo       repositories.QuizRepository
	assignmentRepo repositories.AssignmentRepository
	aiService      services.AIService
}

func NewAIUseCase(
	lessonRepo repositories.LessonRepository,
	quizRepo repositories.QuizRepository,
	assignmentRepo repositories.AssignmentRepository,
	aiService services.AIService,
) *AIUseCase {
	return &AIUseCase{lessonRepo: lessonRepo, quizRepo: quizRepo, assignmentRepo: assignmentRepo, aiService: aiService}
}

func (uc *AIUseCase) SummarizeLesson(ctx context.Context, lessonID, orgID uuid.UUID) (*dto.SummarizeLessonResponse, error) {
	lesson, err := uc.lessonRepo.FindByID(ctx, lessonID, orgID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(lesson.Content) == "" {
		return nil, apperrors.ValidationError("lesson has no text content to summarize")
	}

	summary, err := uc.aiService.GenerateLessonSummary(ctx, lesson.Content)
	if err != nil {
		return nil, err
	}

	return &dto.SummarizeLessonResponse{Summary: summary}, nil
}

func (uc *AIUseCase) GenerateQuiz(ctx context.Context, lessonID, orgID uuid.UUID, numQuestions int) (*dto.QuizDTO, error) {
	lesson, err := uc.lessonRepo.FindByID(ctx, lessonID, orgID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(lesson.Content) == "" {
		return nil, apperrors.ValidationError("lesson has no text content to generate a quiz")
	}

	questions, err := uc.aiService.GenerateQuiz(ctx, lesson.Content, numQuestions)
	if err != nil {
		return nil, err
	}

	// Convert domain questions to QuizDTO questions inline
	questionDTOs := make([]dto.QuestionDTO, len(questions))
	for i, q := range questions {
		answers := make([]dto.AnswerDTO, len(q.Answers))
		for j, a := range q.Answers {
			answers[j] = dto.AnswerDTO{
				ID:        a.ID.String(),
				Answer:    a.Answer,
				IsCorrect: a.IsCorrect,
			}
		}
		questionDTOs[i] = dto.QuestionDTO{
			ID:       q.ID.String(),
			Question: q.Question,
			Position: q.Position,
			Answers:  answers,
		}
	}

	return &dto.QuizDTO{
		LessonID:  lessonID.String(),
		OrgID:     orgID.String(),
		Title:     "AI Generated Quiz for " + lesson.Title,
		Questions: questionDTOs,
	}, nil
}

func (uc *AIUseCase) GetAssignmentFeedback(ctx context.Context, submissionID, userID, orgID uuid.UUID) (*dto.AssignmentFeedbackResponse, error) {
	submission, err := uc.assignmentRepo.FindSubmissionByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	if submission.StudentID != userID {
		return nil, apperrors.ForbiddenError("not your submission")
	}
	if submission.TextContent == "" {
		return nil, apperrors.ValidationError("submission has no text content to review")
	}
	assignment, err := uc.assignmentRepo.FindByID(ctx, submission.AssignmentID, orgID)
	if err != nil {
		return nil, err
	}
	feedback, err := uc.aiService.GenerateAssignmentFeedback(ctx, assignment.Title, assignment.Description, submission.TextContent)
	if err != nil {
		return nil, err
	}
	return &dto.AssignmentFeedbackResponse{
		Strengths:    feedback.Strengths,
		Gaps:         feedback.Gaps,
		Improvements: feedback.Improvements,
		Overall:      feedback.Overall,
	}, nil
}

func (uc *AIUseCase) GenerateFlashcards(ctx context.Context, lessonID, orgID uuid.UUID, numCards int) ([]dto.FlashcardDTO, error) {
	lesson, err := uc.lessonRepo.FindByID(ctx, lessonID, orgID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(lesson.Content) == "" {
		return nil, apperrors.ValidationError("lesson has no text content to generate flashcards")
	}
	cards, err := uc.aiService.GenerateFlashcards(ctx, lesson.Content, numCards)
	if err != nil {
		return nil, err
	}
	result := make([]dto.FlashcardDTO, len(cards))
	for i, c := range cards {
		result[i] = dto.FlashcardDTO{Term: c.Term, Definition: c.Definition}
	}
	return result, nil
}

func (uc *AIUseCase) GenerateTeachingPlan(ctx context.Context, topic, audience, language string, durationMinutes int) (*dto.GenerateTeachingPlanResponse, error) {
	plan, err := uc.aiService.GenerateTeachingPlan(ctx, strings.TrimSpace(topic), strings.TrimSpace(audience), language, durationMinutes)
	if err != nil {
		return nil, err
	}
	return &dto.GenerateTeachingPlanResponse{Plan: plan}, nil
}
