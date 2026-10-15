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

type ProgressUseCase struct {
	progressRepo    repositories.LessonProgressRepository
	lessonRepo      repositories.LessonRepository
	aiService       services.AIService
	moduleRepo      repositories.ModuleRepository
	quizRepo        repositories.QuizRepository
	quizAttemptRepo repositories.QuizAttemptRepository
}

func NewProgressUseCase(
	progressRepo repositories.LessonProgressRepository,
	lessonRepo repositories.LessonRepository,
	aiService services.AIService,
	moduleRepo repositories.ModuleRepository,
	quizRepo repositories.QuizRepository,
	quizAttemptRepo repositories.QuizAttemptRepository,
) *ProgressUseCase {
	return &ProgressUseCase{
		progressRepo:    progressRepo,
		lessonRepo:      lessonRepo,
		aiService:       aiService,
		moduleRepo:      moduleRepo,
		quizRepo:        quizRepo,
		quizAttemptRepo: quizAttemptRepo,
	}
}

func (uc *ProgressUseCase) CompleteLesson(ctx context.Context, userID, lessonID, orgID uuid.UUID, score *float64) error {
	if _, err := uc.lessonRepo.FindByID(ctx, lessonID, orgID); err != nil {
		return err
	}

	existing, _ := uc.progressRepo.FindByUserAndLesson(ctx, userID, lessonID, orgID)
	now := time.Now()
	if existing != nil {
		existing.CompletedAt = &now
		existing.Score = score
		existing.UpdatedAt = now
		return uc.progressRepo.Update(ctx, existing)
	}

	progress := &entities.LessonProgress{
		ID:          uuid.New(),
		UserID:      userID,
		LessonID:    lessonID,
		OrgID:       orgID,
		CompletedAt: &now,
		Score:       score,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return uc.progressRepo.Create(ctx, progress)
}

func (uc *ProgressUseCase) GetStudentProgress(ctx context.Context, userID, orgID uuid.UUID) ([]dto.ProgressDTO, error) {
	progresses, err := uc.progressRepo.FindByUser(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProgressDTO, len(progresses))
	for i, p := range progresses {
		result[i] = toProgressDTO(&p)
	}
	return result, nil
}

func (uc *ProgressUseCase) GetStudentCourseProgress(ctx context.Context, userID, courseID, orgID uuid.UUID) ([]dto.ProgressDTO, error) {
	progresses, err := uc.progressRepo.FindByUserAndCourse(ctx, userID, courseID, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProgressDTO, len(progresses))
	for i, p := range progresses {
		result[i] = toProgressDTO(&p)
	}
	return result, nil
}

func (uc *ProgressUseCase) GetProgressInsights(ctx context.Context, userID, orgID uuid.UUID) (*dto.InsightsDTO, error) {
	progresses, err := uc.progressRepo.FindByUser(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}

	if len(progresses) == 0 {
		return nil, apperrors.NotFoundError("progress", userID.String())
	}

	insights, err := uc.aiService.GenerateProgressInsights(ctx, progresses)
	if err != nil {
		if isAIBudgetError(err) {
			return nil, err
		}
		insights = "Unable to generate insights at this time."
	}

	var totalScore float64
	var scoredCount int
	for _, p := range progresses {
		if p.Score != nil {
			totalScore += *p.Score
			scoredCount++
		}
	}
	avgScore := 0.0
	if scoredCount > 0 {
		avgScore = totalScore / float64(scoredCount)
	}

	return &dto.InsightsDTO{
		Insights:         insights,
		TotalLessons:     len(progresses),
		CompletedLessons: len(progresses),
		AverageScore:     avgScore,
	}, nil
}

func (uc *ProgressUseCase) GetCoursePacing(ctx context.Context, courseID, userID, orgID uuid.UUID) (*dto.CoursePacingDTO, error) {
	modules, err := uc.moduleRepo.FindByCourse(ctx, courseID, orgID)
	if err != nil {
		return nil, err
	}

	allProgress, err := uc.progressRepo.FindByUser(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	progressByLesson := make(map[uuid.UUID]struct{}, len(allProgress))
	for _, p := range allProgress {
		if p.CompletedAt != nil {
			progressByLesson[p.LessonID] = struct{}{}
		}
	}

	// Collect all lesson IDs per module
	moduleToLessons := make(map[uuid.UUID][]entities.Lesson, len(modules))
	var allLessonIDs []uuid.UUID
	for _, mod := range modules {
		lessons, err := uc.lessonRepo.FindByModule(ctx, mod.ID, orgID)
		if err != nil {
			return nil, err
		}
		moduleToLessons[mod.ID] = lessons
		for _, l := range lessons {
			allLessonIDs = append(allLessonIDs, l.ID)
		}
	}

	// Batch fetch quizzes for all lessons
	quizByLesson := make(map[uuid.UUID]entities.Quiz)
	if len(allLessonIDs) > 0 {
		quizzes, err := uc.quizRepo.FindByLessons(ctx, allLessonIDs, orgID)
		if err == nil {
			for _, q := range quizzes {
				quizByLesson[q.LessonID] = q
			}
		}
	}

	// Batch fetch quiz attempts for those quizzes
	var allQuizIDs []uuid.UUID
	for _, q := range quizByLesson {
		allQuizIDs = append(allQuizIDs, q.ID)
	}
	attemptByQuiz := make(map[uuid.UUID]*entities.QuizAttempt)
	if len(allQuizIDs) > 0 {
		attempts, err := uc.quizAttemptRepo.FindByStudentAndQuizzes(ctx, userID, allQuizIDs)
		if err == nil {
			for _, a := range attempts {
				attemptByQuiz[a.QuizID] = a
			}
		}
	}

	modulePacings := make([]dto.ModulePacingDTO, 0, len(modules))
	for _, mod := range modules {
		lessons := moduleToLessons[mod.ID]
		completed := 0
		var totalScore float64
		var scoredCount int

		for _, l := range lessons {
			if _, ok := progressByLesson[l.ID]; ok {
				completed++
			}
			if quiz, ok := quizByLesson[l.ID]; ok {
				if attempt, ok := attemptByQuiz[quiz.ID]; ok && attempt.MaxScore > 0 {
					totalScore += float64(attempt.Score) / float64(attempt.MaxScore) * 100
					scoredCount++
				}
			}
		}

		completionRate := 0.0
		if len(lessons) > 0 {
			completionRate = float64(completed) / float64(len(lessons)) * 100
		}
		avgScore := 0.0
		if scoredCount > 0 {
			avgScore = totalScore / float64(scoredCount)
		}

		modulePacings = append(modulePacings, dto.ModulePacingDTO{
			ModuleID:         mod.ID.String(),
			ModuleTitle:      mod.Title,
			TotalLessons:     len(lessons),
			CompletedLessons: completed,
			CompletionRate:   completionRate,
			AverageScore:     avgScore,
			HasQuizzes:       scoredCount > 0,
			Pace:             pacingSignal(completionRate, avgScore, scoredCount, len(lessons)),
		})
	}

	return &dto.CoursePacingDTO{
		CourseID:    courseID.String(),
		Modules:     modulePacings,
		OverallPace: overallPacingSignal(modulePacings),
	}, nil
}

func pacingSignal(completionRate, avgScore float64, scoredCount, totalLessons int) string {
	if totalLessons == 0 || completionRate == 0 {
		return "not_started"
	}
	if scoredCount > 0 && avgScore < 60 {
		return "struggling"
	}
	if completionRate >= 80 && (scoredCount == 0 || avgScore >= 80) {
		return "ahead"
	}
	return "on_track"
}

func overallPacingSignal(modules []dto.ModulePacingDTO) string {
	if len(modules) == 0 {
		return "not_started"
	}
	counts := map[string]int{}
	for _, m := range modules {
		counts[m.Pace]++
	}
	if counts["struggling"] > 0 {
		return "struggling"
	}
	if counts["not_started"] == len(modules) {
		return "not_started"
	}
	if counts["ahead"] == len(modules) {
		return "ahead"
	}
	return "on_track"
}

func toProgressDTO(p *entities.LessonProgress) dto.ProgressDTO {
	return dto.ProgressDTO{
		ID:          p.ID.String(),
		UserID:      p.UserID.String(),
		LessonID:    p.LessonID.String(),
		OrgID:       p.OrgID.String(),
		CompletedAt: p.CompletedAt,
		Score:       p.Score,
		CreatedAt:   p.CreatedAt,
	}
}
