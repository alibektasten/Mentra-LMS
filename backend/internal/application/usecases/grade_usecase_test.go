package usecases

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGradeUseCase_GetMyGrades_Success_CalculatesPercentage(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	attemptRepo := new(mocks.MockQuizAttemptRepository)
	quizRepo := new(mocks.MockQuizRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewGradeUseCase(assignmentRepo, attemptRepo, quizRepo, memberRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	a1ID := uuid.New()
	a2ID := uuid.New()

	assignments := []*entities.Assignment{
		{ID: a1ID, OrgID: orgID, CourseID: courseID, Title: "HW1", MaxPoints: 100},
		{ID: a2ID, OrgID: orgID, CourseID: courseID, Title: "HW2", MaxPoints: 50},
	}
	sub1 := &entities.AssignmentSubmission{ID: uuid.New(), AssignmentID: a1ID, StudentID: studentID, Score: ptrInt(80)}
	sub2 := &entities.AssignmentSubmission{ID: uuid.New(), AssignmentID: a2ID, StudentID: studentID, Score: ptrInt(40)}

	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return(assignments, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, a1ID, studentID).Return(sub1, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, a2ID, studentID).Return(sub2, nil)

	result, err := uc.GetMyGrades(context.Background(), courseID, studentID, orgID)
	require.NoError(t, err)
	assert.Equal(t, studentID.String(), result.StudentID)
	assert.Equal(t, 120, result.TotalEarned)
	assert.Equal(t, 150, result.TotalPossible)
	assert.Nil(t, result.Percentage, "an assignment-only result is not a final academic grade")
	assert.False(t, result.IsFinalGradeReady)
}

func TestGradeUseCase_GetMyGrades_NoSubmissions_ZeroPercentage(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	attemptRepo := new(mocks.MockQuizAttemptRepository)
	quizRepo := new(mocks.MockQuizRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewGradeUseCase(assignmentRepo, attemptRepo, quizRepo, memberRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	a1ID := uuid.New()

	assignments := []*entities.Assignment{
		{ID: a1ID, OrgID: orgID, CourseID: courseID, Title: "HW1", MaxPoints: 100},
	}

	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return(assignments, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, a1ID, studentID).Return(nil, apperrors.NotFoundError("submission", ""))

	result, err := uc.GetMyGrades(context.Background(), courseID, studentID, orgID)
	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalEarned)
	assert.Equal(t, 100, result.TotalPossible)
	assert.Nil(t, result.Percentage)
}

func TestGradeUseCase_GetPublishedMyGrades_HidesUnpublishedFinalResult(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	assignmentRepo.On("FindByCourse", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.Assignment{}, nil)
	uc := NewGradeUseCase(assignmentRepo, new(mocks.MockQuizAttemptRepository), new(mocks.MockQuizRepository), new(mocks.MockMembershipRepository))

	result, err := uc.GetPublishedMyGrades(context.Background(), uuid.New(), uuid.New(), uuid.New())

	require.NoError(t, err)
	assert.False(t, result.IsFinalGradeReady)
	assert.Nil(t, result.Percentage)
	assert.Nil(t, result.FinalGrade)
	assert.Nil(t, result.OverrideGrade)
	assert.Empty(t, result.OverrideNote)
}

func TestGradeUseCase_GetPublishedMyGrades_OmitsUnpublishedFinalValuesFromJSON(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	examRepo := new(mocks.MockExamRepository)
	examAttemptRepo := new(mocks.MockExamAttemptRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()

	assignmentID := uuid.New()
	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Assignment{{ID: assignmentID, CourseID: courseID, OrgID: orgID, Title: "Assignment", MaxPoints: 100}}, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, assignmentID, studentID).Return(&entities.AssignmentSubmission{Score: ptrInt(100)}, nil)

	exams := make([]*entities.Exam, 0, 3)
	for _, assessmentType := range []string{"midterm", "endterm", "final"} {
		exam := &entities.Exam{ID: uuid.New(), CourseID: courseID, OrgID: orgID, AssessmentType: assessmentType, Title: assessmentType, MCQPoints: 100}
		exams = append(exams, exam)
		examAttemptRepo.On("FindByExamAndStudent", mock.Anything, exam.ID, studentID).Return([]*entities.ExamAttempt{{ExamID: exam.ID, StudentID: studentID, Status: "submitted", TotalScore: ptrInt(100)}}, nil)
	}
	examRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return(exams, nil)
	gradeRepo.On("Find", mock.Anything, courseID, studentID, orgID).Return(nil, nil)

	uc := NewGradeUseCase(assignmentRepo, new(mocks.MockQuizAttemptRepository), new(mocks.MockQuizRepository), new(mocks.MockMembershipRepository), examRepo, examAttemptRepo, nil, nil, gradeRepo)
	result, err := uc.GetPublishedMyGrades(context.Background(), courseID, studentID, orgID)
	require.NoError(t, err)
	assert.True(t, result.IsFinalGradeReady, "the result is calculated even though it has not been published")
	assert.Nil(t, result.FinalGrade)
	assert.Nil(t, result.Percentage)

	payload, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(payload), `"final_grade":`)
	assert.NotContains(t, string(payload), `"percentage":`)
	assert.NotContains(t, string(payload), `"override_grade":`)
}

func TestWeightedGrade_RequiresEveryComponentAndKeepsZeroScores(t *testing.T) {
	full := weightedGrade(floatPtr(100), floatPtr(80), floatPtr(60), floatPtr(40))
	require.NotNil(t, full)
	assert.InDelta(t, 62.5, *full, 0.001)
	assert.Nil(t, weightedGrade(floatPtr(100), nil, floatPtr(60), floatPtr(40)))

	zero := weightedGrade(floatPtr(0), floatPtr(0), floatPtr(0), floatPtr(0))
	require.NotNil(t, zero)
	assert.Equal(t, 0.0, *zero)
}

func TestGradeUseCase_SetGradebookPublication_ReturnsRepositoryFailureWithoutSuccessResult(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	groupRepo := new(mocks.MockGroupRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(assignmentRepo, new(mocks.MockQuizAttemptRepository), new(mocks.MockQuizRepository), new(mocks.MockMembershipRepository), nil, nil, groupRepo, nil, gradeRepo)
	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	groupID := uuid.New()
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{{ID: groupID}}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: studentID}}, nil)
	gradeRepo.On("SetPublication", mock.Anything, mock.MatchedBy(func(grades []*entities.CourseGrade) bool {
		return len(grades) == 1 && grades[0].StudentID == studentID && grades[0].PublishedAt != nil
	})).Return(apperrors.InternalError("database unavailable"))

	result, err := uc.SetGradebookPublication(context.Background(), courseID, orgID, uuid.New(), true, nil)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGradeUseCase_SetGradebookPublication_RejectsStudentOutsideCourse(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(new(mocks.MockAssignmentRepository), new(mocks.MockQuizAttemptRepository), new(mocks.MockQuizRepository), new(mocks.MockMembershipRepository), nil, nil, groupRepo, nil, gradeRepo)
	orgID, courseID := uuid.New(), uuid.New()
	groupID, enrolledID := uuid.New(), uuid.New()
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{{ID: groupID}}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: enrolledID}}, nil)

	result, err := uc.SetGradebookPublication(context.Background(), courseID, orgID, uuid.New(), true, []uuid.UUID{uuid.New()})
	require.Error(t, err)
	assert.Nil(t, result)
	gradeRepo.AssertNotCalled(t, "SetPublication", mock.Anything, mock.Anything)
}

func TestGradeUseCase_SetGradebookPublication_EmptySelectionMeansAllEnrolledStudents(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(new(mocks.MockAssignmentRepository), new(mocks.MockQuizAttemptRepository), new(mocks.MockQuizRepository), new(mocks.MockMembershipRepository), nil, nil, groupRepo, nil, gradeRepo)

	orgID, courseID, groupID, firstStudentID, secondStudentID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{{ID: groupID}}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: firstStudentID}, {StudentID: secondStudentID}}, nil)
	gradeRepo.On("SetPublication", mock.Anything, mock.MatchedBy(func(grades []*entities.CourseGrade) bool {
		return len(grades) == 2 && grades[0].StudentID == firstStudentID && grades[1].StudentID == secondStudentID
	})).Return(nil).Once()

	result, err := uc.SetGradebookPublication(context.Background(), courseID, orgID, uuid.New(), true, []uuid.UUID{})
	require.NoError(t, err)
	assert.Equal(t, 2, result.AffectedCount)
	gradeRepo.AssertExpectations(t)
}

func TestGradeUseCase_SetGradebookPublication_TargetsOnlySelectedEnrolledStudents(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(new(mocks.MockAssignmentRepository), new(mocks.MockQuizAttemptRepository), new(mocks.MockQuizRepository), new(mocks.MockMembershipRepository), nil, nil, groupRepo, nil, gradeRepo)

	orgID, courseID, groupID, selectedID, unselectedID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{{ID: groupID}}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: selectedID}, {StudentID: unselectedID}}, nil)
	gradeRepo.On("SetPublication", mock.Anything, mock.MatchedBy(func(grades []*entities.CourseGrade) bool {
		return len(grades) == 1 && grades[0].StudentID == selectedID
	})).Return(nil).Once()

	result, err := uc.SetGradebookPublication(context.Background(), courseID, orgID, uuid.New(), true, []uuid.UUID{selectedID, selectedID})
	require.NoError(t, err)
	assert.Equal(t, 1, result.AffectedCount)
	gradeRepo.AssertExpectations(t)
}

func TestGradeUseCase_UpdateCourseGrade_ClearOverrideWritesAuditEvent(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	groupRepo := new(mocks.MockGroupRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(assignmentRepo, new(mocks.MockQuizAttemptRepository), new(mocks.MockQuizRepository), new(mocks.MockMembershipRepository), nil, nil, groupRepo, nil, gradeRepo)
	orgID, courseID, studentID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	groupID := uuid.New()
	previous := floatPtr(72.5)
	record := &entities.CourseGrade{ID: uuid.New(), CourseID: courseID, StudentID: studentID, OrgID: orgID, OverrideGrade: previous, OverrideNote: "Initial correction"}
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{{ID: groupID}}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: studentID}}, nil)
	gradeRepo.On("UpdateOverride", mock.Anything, courseID, studentID, orgID, actorID, (*float64)(nil), (*string)(nil), true, mock.Anything).
		Return(&entities.CourseGrade{ID: record.ID, CourseID: courseID, StudentID: studentID, OrgID: orgID}, nil).Once()
	gradeRepo.On("Find", mock.Anything, courseID, studentID, orgID).Return(&entities.CourseGrade{ID: record.ID, CourseID: courseID, StudentID: studentID, OrgID: orgID}, nil).Once()
	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Assignment{}, nil)

	_, err := uc.UpdateCourseGrade(context.Background(), courseID, studentID, orgID, actorID, dto.UpdateCourseGradeRequest{ClearOverride: true})
	require.NoError(t, err)
	gradeRepo.AssertExpectations(t)
}

func TestGradeUseCase_ExamPercentage_UsesBestCompletedAttemptAndKeepsZero(t *testing.T) {
	attemptRepo := new(mocks.MockExamAttemptRepository)
	examID, studentID := uuid.New(), uuid.New()
	best := 80
	lower := 45
	attemptRepo.On("FindByExamAndStudent", mock.Anything, examID, studentID).Return([]*entities.ExamAttempt{
		{Status: "submitted", TotalScore: &lower},
		{Status: "submitted", TotalScore: nil}, // file section is not graded yet
		{Status: "submitted", TotalScore: &best},
	}, nil)
	uc := &GradeUseCase{examAttemptRepo: attemptRepo}
	percent, score, _, _, err := uc.examPercentage(context.Background(), examID, studentID, 100)
	require.NoError(t, err)
	require.NotNil(t, percent)
	require.NotNil(t, score)
	assert.Equal(t, 80, *score)
	assert.Equal(t, 80.0, *percent)

	zeroExamID := uuid.New()
	zero := 0
	attemptRepo.On("FindByExamAndStudent", mock.Anything, zeroExamID, studentID).Return([]*entities.ExamAttempt{{Status: "submitted", TotalScore: &zero}}, nil)
	percent, score, _, _, err = uc.examPercentage(context.Background(), zeroExamID, studentID, 100)
	require.NoError(t, err)
	require.NotNil(t, percent)
	require.NotNil(t, score)
	assert.Equal(t, 0.0, *percent)
}

func TestGradeUseCase_GetUpcomingDeadlines_MarkSubmittedStatus(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	attemptRepo := new(mocks.MockQuizAttemptRepository)
	quizRepo := new(mocks.MockQuizRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewGradeUseCase(assignmentRepo, attemptRepo, quizRepo, memberRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	a1ID := uuid.New()
	a2ID := uuid.New()

	futureDate := time.Now().Add(48 * time.Hour)
	assignments := []*entities.Assignment{
		{ID: a1ID, OrgID: orgID, CourseID: courseID, Title: "HW1", DueDate: &futureDate},
		{ID: a2ID, OrgID: orgID, CourseID: courseID, Title: "HW2", DueDate: &futureDate},
	}
	sub := &entities.AssignmentSubmission{ID: uuid.New(), AssignmentID: a1ID, StudentID: studentID}

	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return(assignments, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, a1ID, studentID).Return(sub, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, a2ID, studentID).Return(nil, apperrors.NotFoundError("submission", ""))

	deadlines, err := uc.GetUpcomingDeadlines(context.Background(), courseID, studentID, orgID)
	require.NoError(t, err)
	require.Len(t, deadlines, 2)

	// Find which is which
	var hw1, hw2 *struct {
		Submitted bool
	}
	for i := range deadlines {
		if deadlines[i].Title == "HW1" {
			hw1 = &struct{ Submitted bool }{deadlines[i].Submitted}
		} else {
			hw2 = &struct{ Submitted bool }{deadlines[i].Submitted}
		}
	}
	require.NotNil(t, hw1)
	require.NotNil(t, hw2)
	assert.True(t, hw1.Submitted)
	assert.False(t, hw2.Submitted)
}

func TestGradeUseCase_GetGradebook_ReturnsOnlyCourseGroupStudents(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	attemptRepo := new(mocks.MockQuizAttemptRepository)
	quizRepo := new(mocks.MockQuizRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	groupRepo := new(mocks.MockGroupRepository)
	userRepo := new(mocks.MockUserRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(assignmentRepo, attemptRepo, quizRepo, memberRepo, nil, nil, groupRepo, userRepo, gradeRepo)

	orgID, courseID := uuid.New(), uuid.New()
	groupID := uuid.New()
	studentA, studentB := uuid.New(), uuid.New()
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{{ID: groupID, OrgID: orgID, CourseID: &courseID}}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: studentB}, {StudentID: studentA}}, nil)
	userRepo.On("FindByIDs", mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2
	})).Return([]entities.User{{ID: studentA, Name: "Aida", Email: "aida@example.com"}, {ID: studentB, Name: "Baur", Email: "baur@example.com"}}, nil)
	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Assignment{}, nil)
	gradeRepo.On("Find", mock.Anything, courseID, studentA, orgID).Return(nil, nil)
	gradeRepo.On("Find", mock.Anything, courseID, studentB, orgID).Return(nil, nil)

	result, err := uc.GetGradebook(context.Background(), courseID, orgID)
	require.NoError(t, err)
	require.Len(t, result.Students, 2)
	assert.Equal(t, "Aida", result.Students[0].StudentName)
	assert.Equal(t, "Baur", result.Students[1].StudentName)
	assert.Equal(t, 0, result.PublishedCount)
}

func TestGradeUseCase_UpdateCourseGrade_PersistsOverrideForEnrolledStudent(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	attemptRepo := new(mocks.MockQuizAttemptRepository)
	quizRepo := new(mocks.MockQuizRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	groupRepo := new(mocks.MockGroupRepository)
	userRepo := new(mocks.MockUserRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(assignmentRepo, attemptRepo, quizRepo, memberRepo, nil, nil, groupRepo, userRepo, gradeRepo)

	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	groupID := uuid.New()
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{{ID: groupID}}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: studentID}}, nil)
	actorID := uuid.New()
	gradeRepo.On("UpdateOverride", mock.Anything, courseID, studentID, orgID, actorID,
		mock.MatchedBy(func(value *float64) bool { return value != nil && *value == 91.5 }),
		mock.MatchedBy(func(value *string) bool { return value != nil && *value == "Approved correction" }), false, mock.Anything,
	).Return(&entities.CourseGrade{ID: uuid.New(), CourseID: courseID, StudentID: studentID, OrgID: orgID, OverrideGrade: floatPtr(91.5), OverrideNote: "Approved correction"}, nil).Once()
	gradeRepo.On("Find", mock.Anything, courseID, studentID, orgID).Return(&entities.CourseGrade{ID: uuid.New(), CourseID: courseID, StudentID: studentID, OrgID: orgID, OverrideGrade: floatPtr(91.5), OverrideNote: "Approved correction"}, nil).Once()
	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Assignment{}, nil)

	result, err := uc.UpdateCourseGrade(context.Background(), courseID, studentID, orgID, actorID, dto.UpdateCourseGradeRequest{OverrideGrade: floatPtr(91.5), OverrideNote: stringPtr("Approved correction")})
	require.NoError(t, err)
	require.NotNil(t, result.OverrideGrade)
	assert.Equal(t, 91.5, *result.OverrideGrade)
	assert.Equal(t, "Approved correction", result.OverrideNote)
	gradeRepo.AssertExpectations(t)
}

func TestGradeUseCase_UpdateCourseGrade_RejectsUnenrolledStudent(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	attemptRepo := new(mocks.MockQuizAttemptRepository)
	quizRepo := new(mocks.MockQuizRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	groupRepo := new(mocks.MockGroupRepository)
	userRepo := new(mocks.MockUserRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(assignmentRepo, attemptRepo, quizRepo, memberRepo, nil, nil, groupRepo, userRepo, gradeRepo)

	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{}, nil)

	_, err := uc.UpdateCourseGrade(context.Background(), courseID, studentID, orgID, uuid.New(), dto.UpdateCourseGradeRequest{OverrideGrade: floatPtr(80)})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.HTTPStatus)
	gradeRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestGradeUseCase_UpdateCourseGrade_RejectsWhitespaceOnlyReason(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	groupRepo := new(mocks.MockGroupRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)
	uc := NewGradeUseCase(assignmentRepo, new(mocks.MockQuizAttemptRepository), new(mocks.MockQuizRepository), new(mocks.MockMembershipRepository), nil, nil, groupRepo, nil, gradeRepo)

	orgID, courseID, studentID, groupID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{{ID: groupID}}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: studentID}}, nil)
	_, err := uc.UpdateCourseGrade(context.Background(), courseID, studentID, orgID, uuid.New(), dto.UpdateCourseGradeRequest{
		OverrideGrade: floatPtr(80), OverrideNote: stringPtr(" \t "),
	})
	require.Error(t, err)
	gradeRepo.AssertNotCalled(t, "UpdateOverride", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func floatPtr(value float64) *float64 { return &value }

func stringPtr(value string) *string { return &value }
