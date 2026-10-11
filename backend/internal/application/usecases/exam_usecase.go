package usecases

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type ExamUseCase struct {
	examRepo    repositories.ExamRepository
	attemptRepo repositories.ExamAttemptRepository
	grantRepo   repositories.ExtraAttemptGrantRepository
	courseRepo  repositories.CourseRepository
}

func NewExamUseCase(
	examRepo repositories.ExamRepository,
	attemptRepo repositories.ExamAttemptRepository,
	grantRepo repositories.ExtraAttemptGrantRepository,
	courseRepo repositories.CourseRepository,
) *ExamUseCase {
	return &ExamUseCase{
		examRepo:    examRepo,
		attemptRepo: attemptRepo,
		grantRepo:   grantRepo,
		courseRepo:  courseRepo,
	}
}

func (uc *ExamUseCase) CreateExam(ctx context.Context, courseID, orgID uuid.UUID, req dto.CreateExamRequest) (*dto.ExamDTO, error) {
	if _, err := uc.courseRepo.FindByID(ctx, courseID, orgID); err != nil {
		return nil, err
	}

	if !req.MCQEnabled && !req.FileEnabled {
		return nil, apperrors.ValidationError("at least one section (MCQ or file upload) must be enabled")
	}

	if req.MCQEnabled && len(req.Questions) == 0 {
		return nil, apperrors.ValidationError("MCQ section requires at least 1 question")
	}

	maxAttempts := req.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	questions := buildExamQuestions(uuid.Nil, req.Questions)

	exam := &entities.Exam{
		ID:              uuid.New(),
		CourseID:        courseID,
		OrgID:           orgID,
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		MaxAttempts:     maxAttempts,
		DueDate:         req.DueDate,
		MCQEnabled:      req.MCQEnabled,
		MCQPoints:       req.MCQPoints,
		FileEnabled:     req.FileEnabled,
		FilePoints:      req.FilePoints,
		Questions:       questions,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := uc.examRepo.Create(ctx, exam); err != nil {
		return nil, err
	}
	return toExamDTO(exam, true), nil
}

func (uc *ExamUseCase) GetExam(ctx context.Context, id, orgID uuid.UUID) (*dto.ExamDTO, error) {
	exam, err := uc.examRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	return toExamDTO(exam, true), nil
}

func (uc *ExamUseCase) ListExams(ctx context.Context, courseID, orgID uuid.UUID) ([]*dto.ExamListItemDTO, error) {
	exams, err := uc.examRepo.FindByCourse(ctx, courseID, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.ExamListItemDTO, len(exams))
	for i, e := range exams {
		result[i] = toExamListItemDTO(e)
	}
	return result, nil
}

func (uc *ExamUseCase) UpdateExam(ctx context.Context, id, orgID uuid.UUID, req dto.UpdateExamRequest) (*dto.ExamDTO, error) {
	existing, err := uc.examRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		existing.DurationMinutes = *req.DurationMinutes
	}
	if req.MaxAttempts != nil {
		existing.MaxAttempts = *req.MaxAttempts
	}
	if req.DueDate != nil {
		existing.DueDate = req.DueDate
	}
	if req.MCQEnabled != nil {
		existing.MCQEnabled = *req.MCQEnabled
	}
	if req.MCQPoints != nil {
		existing.MCQPoints = *req.MCQPoints
	}
	if req.FileEnabled != nil {
		existing.FileEnabled = *req.FileEnabled
	}
	if req.FilePoints != nil {
		existing.FilePoints = *req.FilePoints
	}

	if !existing.MCQEnabled && !existing.FileEnabled {
		return nil, apperrors.ValidationError("at least one section (MCQ or file upload) must be enabled")
	}

	if existing.MCQEnabled {
		if req.Questions != nil {
			if len(req.Questions) == 0 {
				return nil, apperrors.ValidationError("MCQ section requires at least 1 question")
			}
			existing.Questions = buildExamQuestions(existing.ID, req.Questions)
		}
	} else {
		existing.Questions = nil
	}

	existing.UpdatedAt = time.Now()
	if err := uc.examRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return toExamDTO(existing, true), nil
}

func (uc *ExamUseCase) DeleteExam(ctx context.Context, id, orgID uuid.UUID) error {
	return uc.examRepo.Delete(ctx, id, orgID)
}

func (uc *ExamUseCase) StartAttempt(ctx context.Context, examID, studentID, orgID uuid.UUID) (*dto.StartAttemptResponse, error) {
	exam, err := uc.examRepo.FindByID(ctx, examID, orgID)
	if err != nil {
		return nil, err
	}

	if exam.DueDate != nil && time.Now().After(*exam.DueDate) {
		return nil, apperrors.ValidationError("exam due date has passed")
	}

	// Idempotent: return existing in_progress attempt
	active, err := uc.attemptRepo.FindActiveAttempt(ctx, examID, studentID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return &dto.StartAttemptResponse{
			AttemptID:    active.ID.String(),
			ExamID:       exam.ID.String(),
			StartedAt:    active.StartedAt,
			ExpiresAt:    active.ExpiresAt,
			SavedAnswers: toExamMCQAnswerInputs(active.MCQAnswers),
			Exam:         toExamDTO(exam, false),
		}, nil
	}

	// Check attempt limit
	count, err := uc.attemptRepo.CountByExamAndStudent(ctx, examID, studentID)
	if err != nil {
		return nil, err
	}
	grants, err := uc.grantRepo.SumByExamAndStudent(ctx, examID, studentID)
	if err != nil {
		return nil, err
	}
	if count >= exam.MaxAttempts+grants {
		return nil, apperrors.ValidationError("maximum attempts reached")
	}

	now := time.Now()
	attempt := &entities.ExamAttempt{
		ID:         uuid.New(),
		ExamID:     examID,
		StudentID:  studentID,
		OrgID:      orgID,
		Status:     "in_progress",
		StartedAt:  now,
		ExpiresAt:  now.Add(time.Duration(exam.DurationMinutes) * time.Minute),
		FilePoints: exam.FilePoints,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.attemptRepo.Create(ctx, attempt); err != nil {
		return nil, err
	}

	return &dto.StartAttemptResponse{
		AttemptID:    attempt.ID.String(),
		ExamID:       exam.ID.String(),
		StartedAt:    attempt.StartedAt,
		ExpiresAt:    attempt.ExpiresAt,
		SavedAnswers: []dto.ExamMCQAnswerInput{},
		Exam:         toExamDTO(exam, false),
	}, nil
}

// SaveDraft persists an in-progress MCQ attempt without grading it. Reloading
// the page therefore cannot consume an attempt or finalise student work.
func (uc *ExamUseCase) SaveDraft(ctx context.Context, attemptID, studentID uuid.UUID, answers []dto.ExamMCQAnswerInput) error {
	attempt, err := uc.attemptRepo.FindByID(ctx, attemptID)
	if err != nil {
		return err
	}
	if attempt.StudentID != studentID {
		return apperrors.ForbiddenError("not your attempt")
	}
	if attempt.Status != "in_progress" {
		return apperrors.ValidationError("attempt is not in progress")
	}
	if time.Now().After(attempt.ExpiresAt) {
		attempt.Status = "expired"
		attempt.UpdatedAt = time.Now()
		_ = uc.attemptRepo.Update(ctx, attempt)
		return apperrors.ValidationError("exam time has expired")
	}

	exam, err := uc.examRepo.FindByID(ctx, attempt.ExamID, attempt.OrgID)
	if err != nil {
		return err
	}
	if err := validateExamAnswers(exam, answers); err != nil {
		return err
	}
	attempt.MCQAnswers = toEntityExamMCQAnswers(answers)
	attempt.UpdatedAt = time.Now()
	return uc.attemptRepo.Update(ctx, attempt)
}

func (uc *ExamUseCase) SubmitAttempt(ctx context.Context, attemptID, studentID uuid.UUID, mcqAnswers []dto.ExamMCQAnswerInput, filePath string) (*dto.ExamAttemptDTO, error) {
	attempt, err := uc.attemptRepo.FindByID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	if attempt.StudentID != studentID {
		return nil, apperrors.ForbiddenError("not your attempt")
	}
	if attempt.Status != "in_progress" {
		return nil, apperrors.ValidationError("attempt is not in progress")
	}

	now := time.Now()
	grace := attempt.ExpiresAt.Add(60 * time.Second)
	if now.After(grace) {
		attempt.Status = "expired"
		attempt.UpdatedAt = now
		_ = uc.attemptRepo.Update(ctx, attempt)
		return nil, apperrors.ValidationError("exam time has expired")
	}

	exam, err := uc.examRepo.FindByID(ctx, attempt.ExamID, attempt.OrgID)
	if err != nil {
		return nil, err
	}
	if err := validateExamAnswers(exam, mcqAnswers); err != nil {
		return nil, err
	}

	// Convert and store MCQ answers
	attempt.MCQAnswers = toEntityExamMCQAnswers(mcqAnswers)

	// Auto-grade MCQ
	if exam.MCQEnabled && len(exam.Questions) > 0 {
		correctCount := 0
		answerMap := make(map[string]string)
		for _, a := range mcqAnswers {
			answerMap[a.QuestionID] = a.AnswerID
		}
		for _, q := range exam.Questions {
			selectedAnswerID, ok := answerMap[q.ID.String()]
			if !ok {
				continue
			}
			for _, a := range q.Answers {
				if a.ID.String() == selectedAnswerID && a.IsCorrect {
					correctCount++
					break
				}
			}
		}
		numQuestions := len(exam.Questions)
		mcqScore := 0
		if numQuestions > 0 {
			mcqScore = (correctCount * exam.MCQPoints) / numQuestions
		}
		attempt.MCQScore = &mcqScore
		attempt.MCQMaxScore = exam.MCQPoints
	}

	if filePath != "" {
		attempt.FilePath = filePath
		attempt.FilePoints = exam.FilePoints
	}

	attempt.Status = "submitted"
	attempt.SubmittedAt = &now
	attempt.UpdatedAt = now

	// Compute total if file section not needed or already graded
	if !exam.FileEnabled || filePath == "" {
		if attempt.MCQScore != nil {
			total := *attempt.MCQScore
			attempt.TotalScore = &total
		}
	}

	if err := uc.attemptRepo.Update(ctx, attempt); err != nil {
		return nil, err
	}
	return toExamAttemptDTO(attempt), nil
}

func (uc *ExamUseCase) MyAttempts(ctx context.Context, examID, studentID uuid.UUID) ([]*dto.ExamAttemptDTO, error) {
	attempts, err := uc.attemptRepo.FindByExamAndStudent(ctx, examID, studentID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.ExamAttemptDTO, len(attempts))
	for i, a := range attempts {
		result[i] = toExamAttemptDTO(a)
	}
	return result, nil
}

func (uc *ExamUseCase) ListAttempts(ctx context.Context, examID uuid.UUID) ([]*dto.ExamAttemptDTO, error) {
	attempts, err := uc.attemptRepo.FindByExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.ExamAttemptDTO, len(attempts))
	for i, a := range attempts {
		result[i] = toExamAttemptDTO(a)
	}
	return result, nil
}

func (uc *ExamUseCase) GradeFileSection(ctx context.Context, attemptID, graderID uuid.UUID, req dto.GradeExamFileRequest) (*dto.ExamAttemptDTO, error) {
	attempt, err := uc.attemptRepo.FindByID(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if attempt.Status != "submitted" {
		return nil, apperrors.ValidationError("attempt has not been submitted")
	}

	exam, err := uc.examRepo.FindByID(ctx, attempt.ExamID, attempt.OrgID)
	if err != nil {
		return nil, err
	}
	if req.Score < 0 || req.Score > exam.FilePoints {
		return nil, apperrors.ValidationError("score must be between 0 and the file section maximum")
	}

	now := time.Now()
	attempt.FileScore = &req.Score
	attempt.FileFeedback = req.Feedback
	attempt.GradedBy = &graderID
	attempt.GradedAt = &now
	attempt.UpdatedAt = now

	// Recompute total score
	if exam.MCQEnabled && exam.FileEnabled {
		if attempt.MCQScore != nil {
			total := *attempt.MCQScore + req.Score
			attempt.TotalScore = &total
		}
	} else if exam.FileEnabled {
		attempt.TotalScore = &req.Score
	}

	if err := uc.attemptRepo.Update(ctx, attempt); err != nil {
		return nil, err
	}
	return toExamAttemptDTO(attempt), nil
}

func (uc *ExamUseCase) GrantExtraAttempt(ctx context.Context, examID, grantedByID, orgID uuid.UUID, req dto.GrantExtraAttemptRequest) error {
	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return apperrors.ValidationError("invalid student id")
	}

	grant := &entities.ExtraAttemptGrant{
		ID:         uuid.New(),
		ExamID:     examID,
		StudentID:  studentID,
		OrgID:      orgID,
		GrantedBy:  grantedByID,
		ExtraCount: req.ExtraCount,
		CreatedAt:  time.Now(),
	}
	return uc.grantRepo.Create(ctx, grant)
}

func buildExamQuestions(examID uuid.UUID, reqs []dto.CreateExamQuestionRequest) []entities.ExamQuestion {
	questions := make([]entities.ExamQuestion, len(reqs))
	for i, q := range reqs {
		qID := uuid.New()
		answers := make([]entities.ExamAnswer, len(q.Answers))
		for j, a := range q.Answers {
			answers[j] = entities.ExamAnswer{
				ID:         uuid.New(),
				QuestionID: qID,
				Answer:     a.Answer,
				IsCorrect:  a.IsCorrect,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
		}
		questions[i] = entities.ExamQuestion{
			ID:        qID,
			ExamID:    examID,
			Question:  q.Question,
			Position:  q.Position,
			Answers:   answers,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	return questions
}

func toExamDTO(e *entities.Exam, withCorrect bool) *dto.ExamDTO {
	questions := make([]dto.ExamQuestionDTO, len(e.Questions))
	for i, q := range e.Questions {
		answers := make([]dto.ExamAnswerDTO, len(q.Answers))
		for j, a := range q.Answers {
			isCorrect := false
			if withCorrect {
				isCorrect = a.IsCorrect
			}
			answers[j] = dto.ExamAnswerDTO{
				ID:        a.ID.String(),
				Answer:    a.Answer,
				IsCorrect: isCorrect,
			}
		}
		questions[i] = dto.ExamQuestionDTO{
			ID:       q.ID.String(),
			Question: q.Question,
			Position: q.Position,
			Answers:  answers,
		}
	}
	return &dto.ExamDTO{
		ID:              e.ID.String(),
		CourseID:        e.CourseID.String(),
		OrgID:           e.OrgID.String(),
		Title:           e.Title,
		Description:     e.Description,
		DurationMinutes: e.DurationMinutes,
		MaxAttempts:     e.MaxAttempts,
		TotalPoints:     e.MCQPoints + e.FilePoints,
		DueDate:         e.DueDate,
		MCQEnabled:      e.MCQEnabled,
		MCQPoints:       e.MCQPoints,
		FileEnabled:     e.FileEnabled,
		FilePoints:      e.FilePoints,
		Questions:       questions,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func toExamListItemDTO(e *entities.Exam) *dto.ExamListItemDTO {
	return &dto.ExamListItemDTO{
		ID:              e.ID.String(),
		CourseID:        e.CourseID.String(),
		OrgID:           e.OrgID.String(),
		Title:           e.Title,
		Description:     e.Description,
		DurationMinutes: e.DurationMinutes,
		MaxAttempts:     e.MaxAttempts,
		TotalPoints:     e.MCQPoints + e.FilePoints,
		DueDate:         e.DueDate,
		MCQEnabled:      e.MCQEnabled,
		MCQPoints:       e.MCQPoints,
		FileEnabled:     e.FileEnabled,
		FilePoints:      e.FilePoints,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func toExamAttemptDTO(a *entities.ExamAttempt) *dto.ExamAttemptDTO {
	mcqAnswers := toExamMCQAnswerInputs(a.MCQAnswers)
	return &dto.ExamAttemptDTO{
		ID:           a.ID.String(),
		ExamID:       a.ExamID.String(),
		StudentID:    a.StudentID.String(),
		Status:       a.Status,
		StartedAt:    a.StartedAt,
		ExpiresAt:    a.ExpiresAt,
		SubmittedAt:  a.SubmittedAt,
		MCQAnswers:   mcqAnswers,
		MCQScore:     a.MCQScore,
		MCQMaxScore:  a.MCQMaxScore,
		FilePath:     a.FilePath,
		FileScore:    a.FileScore,
		FilePoints:   a.FilePoints,
		FileFeedback: a.FileFeedback,
		TotalScore:   a.TotalScore,
		GradedAt:     a.GradedAt,
	}
}

func toEntityExamMCQAnswers(answers []dto.ExamMCQAnswerInput) []entities.ExamMCQAnswer {
	result := make([]entities.ExamMCQAnswer, len(answers))
	for i, answer := range answers {
		result[i] = entities.ExamMCQAnswer{QuestionID: answer.QuestionID, AnswerID: answer.AnswerID}
	}
	return result
}

func toExamMCQAnswerInputs(answers []entities.ExamMCQAnswer) []dto.ExamMCQAnswerInput {
	result := make([]dto.ExamMCQAnswerInput, len(answers))
	for i, answer := range answers {
		result[i] = dto.ExamMCQAnswerInput{QuestionID: answer.QuestionID, AnswerID: answer.AnswerID}
	}
	return result
}

func validateExamAnswers(exam *entities.Exam, answers []dto.ExamMCQAnswerInput) error {
	validAnswers := make(map[string]map[string]bool, len(exam.Questions))
	for _, question := range exam.Questions {
		validAnswers[question.ID.String()] = make(map[string]bool, len(question.Answers))
		for _, answer := range question.Answers {
			validAnswers[question.ID.String()][answer.ID.String()] = true
		}
	}
	seen := make(map[string]bool, len(answers))
	for _, answer := range answers {
		if seen[answer.QuestionID] || !validAnswers[answer.QuestionID][answer.AnswerID] {
			return apperrors.ValidationError("invalid exam answer")
		}
		seen[answer.QuestionID] = true
	}
	return nil
}
