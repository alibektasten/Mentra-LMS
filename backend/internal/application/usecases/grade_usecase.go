package usecases

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type GradeUseCase struct {
	assignmentRepo  repositories.AssignmentRepository
	attemptRepo     repositories.QuizAttemptRepository
	quizRepo        repositories.QuizRepository
	memberRepo      repositories.MembershipRepository
	examRepo        repositories.ExamRepository
	examAttemptRepo repositories.ExamAttemptRepository
	groupRepo       repositories.GroupRepository
	userRepo        repositories.UserRepository
	courseGradeRepo repositories.CourseGradeRepository
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
	var groupRepo repositories.GroupRepository
	var userRepo repositories.UserRepository
	var courseGradeRepo repositories.CourseGradeRepository
	if len(extras) > 2 {
		groupRepo, _ = extras[2].(repositories.GroupRepository)
	}
	if len(extras) > 3 {
		userRepo, _ = extras[3].(repositories.UserRepository)
	}
	if len(extras) > 4 {
		courseGradeRepo, _ = extras[4].(repositories.CourseGradeRepository)
	}
	return &GradeUseCase{
		assignmentRepo:  assignmentRepo,
		attemptRepo:     attemptRepo,
		quizRepo:        quizRepo,
		memberRepo:      memberRepo,
		examRepo:        examRepo,
		examAttemptRepo: examAttemptRepo,
		groupRepo:       groupRepo,
		userRepo:        userRepo,
		courseGradeRepo: courseGradeRepo,
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
	calculatedFinalGrade := weightedGrade(assignmentsPercent, midtermPercent, endtermPercent, finalExamPercent)
	isReady := calculatedFinalGrade != nil

	result := &dto.StudentGradeDTO{
		StudentID:            studentID.String(),
		Items:                items,
		TotalEarned:          totalEarned,
		TotalPossible:        totalPossible,
		Percentage:           calculatedFinalGrade,
		AssignmentsPercent:   assignmentsPercent,
		MidtermPercent:       midtermPercent,
		EndtermPercent:       endtermPercent,
		FinalExamPercent:     finalExamPercent,
		MidtermRegister:      midtermRegister,
		EndtermRegister:      endtermRegister,
		CalculatedFinalGrade: calculatedFinalGrade,
		FinalGrade:           calculatedFinalGrade,
		IsFinalGradeReady:    isReady,
	}
	if err := uc.applyCourseGrade(ctx, courseID, studentID, orgID, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPublishedMyGrades is the student-facing view of a course result. Source
// assessment marks remain visible to support ongoing learning, but the
// calculated final result and any teacher adjustment are withheld until the
// teacher publishes the gradebook.
func (uc *GradeUseCase) GetPublishedMyGrades(ctx context.Context, courseID, studentID, orgID uuid.UUID) (*dto.StudentGradeDTO, error) {
	result, err := uc.GetMyGrades(ctx, courseID, studentID, orgID)
	if err != nil || result.IsPublished {
		return result, err
	}

	result.Percentage = nil
	result.CalculatedFinalGrade = nil
	result.FinalGrade = nil
	result.OverrideGrade = nil
	result.OverrideNote = ""
	return result, nil
}

// GetGradebook returns one academic-register row for every student enrolled in
// one of the course's groups. Students in another organisation or an unrelated
// course cannot be included through this endpoint.
func (uc *GradeUseCase) GetGradebook(ctx context.Context, courseID, orgID uuid.UUID) (*dto.GradebookDTO, error) {
	if uc.groupRepo == nil || uc.userRepo == nil {
		return nil, apperrors.InternalError("gradebook dependencies are not configured")
	}
	studentIDs, err := uc.enrolledStudentIDs(ctx, courseID, orgID)
	if err != nil {
		return nil, err
	}
	users, err := uc.userRepo.FindByIDs(ctx, studentIDs)
	if err != nil {
		return nil, err
	}
	usersByID := make(map[uuid.UUID]entities.User, len(users))
	for _, user := range users {
		usersByID[user.ID] = user
	}

	students := make([]dto.GradebookStudentDTO, 0, len(studentIDs))
	publishedCount := 0
	for _, studentID := range studentIDs {
		user, ok := usersByID[studentID]
		if !ok {
			continue
		}
		grade, err := uc.GetMyGrades(ctx, courseID, studentID, orgID)
		if err != nil {
			return nil, err
		}
		if grade.IsPublished {
			publishedCount++
		}
		students = append(students, dto.GradebookStudentDTO{
			StudentID: studentID.String(), StudentName: user.Name, StudentEmail: user.Email, Grade: grade,
		})
	}
	sort.Slice(students, func(i, j int) bool {
		return strings.ToLower(students[i].StudentName) < strings.ToLower(students[j].StudentName)
	})
	return &dto.GradebookDTO{
		CourseID: courseID.String(), Students: students, PublishedCount: publishedCount, GeneratedAt: time.Now(),
	}, nil
}

// UpdateCourseGrade records a transparent final-grade adjustment. It never
// overwrites source assignment/exam scores, so the calculated result remains
// available alongside the teacher's correction.
func (uc *GradeUseCase) UpdateCourseGrade(ctx context.Context, courseID, studentID, orgID, actorID uuid.UUID, req dto.UpdateCourseGradeRequest) (*dto.StudentGradeDTO, error) {
	if uc.courseGradeRepo == nil {
		return nil, apperrors.InternalError("gradebook overrides are not configured")
	}
	if !req.ClearOverride && req.OverrideGrade == nil && req.OverrideNote == nil {
		return nil, apperrors.ValidationError("provide an override grade, note, or clear_override")
	}
	if req.OverrideGrade != nil && (*req.OverrideGrade < 0 || *req.OverrideGrade > 100) {
		return nil, apperrors.ValidationError("override grade must be between 0 and 100")
	}
	if err := uc.requireEnrolledStudent(ctx, courseID, studentID, orgID); err != nil {
		return nil, err
	}
	if !req.ClearOverride && (req.OverrideNote == nil || strings.TrimSpace(*req.OverrideNote) == "") {
		return nil, apperrors.ValidationError("a reason is required for a manual adjustment")
	}
	var normalizedNote *string
	if req.OverrideNote != nil {
		note := strings.TrimSpace(*req.OverrideNote)
		if len(note) > 2000 {
			return nil, apperrors.ValidationError("override note must be 2000 characters or fewer")
		}
		normalizedNote = &note
	}
	if _, err := uc.courseGradeRepo.UpdateOverride(ctx, courseID, studentID, orgID, actorID, req.OverrideGrade, normalizedNote, req.ClearOverride, time.Now().UTC()); err != nil {
		return nil, err
	}
	return uc.GetMyGrades(ctx, courseID, studentID, orgID)
}

// SetGradebookPublication publishes or withdraws every enrolled student's
// course result. A publication record is retained when withdrawn so the state
// is explicit rather than inferred from a missing row.
func (uc *GradeUseCase) SetGradebookPublication(ctx context.Context, courseID, orgID, actorID uuid.UUID, published bool, requestedStudentIDs []uuid.UUID) (*dto.GradebookPublicationResultDTO, error) {
	if uc.courseGradeRepo == nil {
		return nil, apperrors.InternalError("gradebook publication is not configured")
	}
	enrolledStudentIDs, err := uc.enrolledStudentIDs(ctx, courseID, orgID)
	if err != nil {
		return nil, err
	}
	studentIDs, err := selectEnrolledStudents(enrolledStudentIDs, requestedStudentIDs)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	grades := make([]*entities.CourseGrade, 0, len(studentIDs))
	for _, studentID := range studentIDs {
		record := &entities.CourseGrade{ID: uuid.New(), CourseID: courseID, StudentID: studentID, OrgID: orgID, CreatedAt: now, UpdatedAt: now}
		record.PublicationChangedAt = &now
		record.PublicationChangedBy = &actorID
		if published {
			record.PublishedAt = &now
			record.PublishedBy = &actorID
		}
		grades = append(grades, record)
	}
	if err := uc.courseGradeRepo.SetPublication(ctx, grades); err != nil {
		return nil, err
	}
	result := &dto.GradebookPublicationResultDTO{Published: published, AffectedCount: len(grades)}
	if published {
		result.PublishedAt = &now
	}
	return result, nil
}

func (uc *GradeUseCase) applyCourseGrade(ctx context.Context, courseID, studentID, orgID uuid.UUID, result *dto.StudentGradeDTO) error {
	if uc.courseGradeRepo == nil {
		return nil
	}
	record, err := uc.courseGradeRepo.Find(ctx, courseID, studentID, orgID)
	if err != nil {
		return err
	}
	if record == nil {
		return nil
	}
	result.OverrideGrade = record.OverrideGrade
	result.OverrideNote = record.OverrideNote
	result.IsPublished = record.PublishedAt != nil
	result.PublishedAt = record.PublishedAt
	if record.OverrideGrade != nil {
		result.FinalGrade = copyFloat(record.OverrideGrade)
		result.Percentage = copyFloat(record.OverrideGrade)
	}
	return nil
}

func (uc *GradeUseCase) enrolledStudentIDs(ctx context.Context, courseID, orgID uuid.UUID) ([]uuid.UUID, error) {
	if uc.groupRepo == nil {
		return nil, apperrors.InternalError("course enrolment dependencies are not configured")
	}
	groups, err := uc.groupRepo.ListGroupsByCourse(ctx, courseID, orgID)
	if err != nil {
		return nil, err
	}
	seen := make(map[uuid.UUID]struct{})
	studentIDs := make([]uuid.UUID, 0)
	for _, group := range groups {
		members, err := uc.groupRepo.ListMembers(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		for _, member := range members {
			if _, ok := seen[member.StudentID]; ok {
				continue
			}
			seen[member.StudentID] = struct{}{}
			studentIDs = append(studentIDs, member.StudentID)
		}
	}
	return studentIDs, nil
}

func (uc *GradeUseCase) requireEnrolledStudent(ctx context.Context, courseID, studentID, orgID uuid.UUID) error {
	studentIDs, err := uc.enrolledStudentIDs(ctx, courseID, orgID)
	if err != nil {
		return err
	}
	for _, id := range studentIDs {
		if id == studentID {
			return nil
		}
	}
	return apperrors.ForbiddenError(fmt.Sprintf("student %s is not enrolled in this course", studentID))
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

func weightedGrade(assignments, midterm, endterm, finalExam *float64) *float64 {
	if assignments == nil || midterm == nil || endterm == nil || finalExam == nil {
		return nil
	}
	value := *assignments*0.25 + *midterm*0.125 + *endterm*0.125 + *finalExam*0.5
	return &value
}

func copyFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func selectEnrolledStudents(enrolled, requested []uuid.UUID) ([]uuid.UUID, error) {
	if len(requested) == 0 {
		return enrolled, nil
	}
	allowed := make(map[uuid.UUID]struct{}, len(enrolled))
	for _, studentID := range enrolled {
		allowed[studentID] = struct{}{}
	}
	seen := make(map[uuid.UUID]struct{}, len(requested))
	result := make([]uuid.UUID, 0, len(requested))
	for _, studentID := range requested {
		if _, ok := allowed[studentID]; !ok {
			return nil, apperrors.ForbiddenError(fmt.Sprintf("student %s is not enrolled in this course", studentID))
		}
		if _, duplicate := seen[studentID]; !duplicate {
			seen[studentID] = struct{}{}
			result = append(result, studentID)
		}
	}
	return result, nil
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
