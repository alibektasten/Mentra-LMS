// Additional coverage tests for simple CRUD methods
package usecases

import (
	"context"
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

// ---- Assignment CRUD ----

func TestAssignmentUseCase_GetByModule(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	moduleID := uuid.New()
	assignments := []*entities.Assignment{
		{ID: uuid.New(), ModuleID: moduleID, OrgID: orgID, Title: "A1", MaxPoints: 50},
	}
	assignmentRepo.On("FindByModule", mock.Anything, moduleID, orgID).Return(assignments, nil)

	result, err := uc.GetByModule(context.Background(), moduleID, orgID)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "A1", result[0].Title)
}

func TestAssignmentUseCase_GetByID(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	assignmentID := uuid.New()
	a := &entities.Assignment{ID: assignmentID, OrgID: orgID, Title: "HW1", MaxPoints: 100}
	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(a, nil)

	result, err := uc.GetByID(context.Background(), assignmentID, orgID)
	require.NoError(t, err)
	assert.Equal(t, "HW1", result.Title)
}

func TestAssignmentUseCase_Update(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	assignmentID := uuid.New()
	a := &entities.Assignment{
		ID: assignmentID, OrgID: orgID, Title: "Old", Description: "Old desc",
		MaxPoints: 50, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(a, nil)
	assignmentRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Assignment")).Return(nil)

	result, err := uc.Update(context.Background(), assignmentID, orgID, dto.UpdateAssignmentRequest{
		Title: ptrString("New Title"), MaxPoints: ptrInt(100),
	})
	require.NoError(t, err)
	assert.Equal(t, "New Title", result.Title)
	assert.Equal(t, 100, result.MaxPoints)
}

func TestAssignmentUseCase_Delete(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	assignmentID := uuid.New()
	assignmentRepo.On("Delete", mock.Anything, assignmentID, orgID).Return(nil)

	err := uc.Delete(context.Background(), assignmentID, orgID)
	require.NoError(t, err)
}

func TestAssignmentUseCase_GetMySubmission(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	assignmentID := uuid.New()
	studentID := uuid.New()
	sub := &entities.AssignmentSubmission{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		StudentID:    studentID,
		TextContent:  "My answer",
	}
	assignmentRepo.On("FindSubmission", mock.Anything, assignmentID, studentID).Return(sub, nil)

	result, err := uc.GetMySubmission(context.Background(), assignmentID, studentID)
	require.NoError(t, err)
	assert.Equal(t, "My answer", result.TextContent)
}

func TestAssignmentUseCase_ListSubmissions(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	assignmentID := uuid.New()
	subs := []*entities.AssignmentSubmission{
		{ID: uuid.New(), AssignmentID: assignmentID, TextContent: "S1"},
		{ID: uuid.New(), AssignmentID: assignmentID, TextContent: "S2"},
	}
	assignmentRepo.On("FindSubmissionsByAssignment", mock.Anything, assignmentID).Return(subs, nil)

	result, err := uc.ListSubmissions(context.Background(), assignmentID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// ---- Exam CRUD ----

func TestExamUseCase_GetExam(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)

	result, err := uc.GetExam(context.Background(), exam.ID, orgID)
	require.NoError(t, err)
	assert.Equal(t, exam.ID.String(), result.ID)
}

func TestExamUseCase_ListExams(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exams := []*entities.Exam{
		makeExamWithQuestions(orgID, courseID, true, false),
		makeExamWithQuestions(orgID, courseID, false, true),
	}
	examRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return(exams, nil)

	result, err := uc.ListExams(context.Background(), courseID, orgID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestExamUseCase_DeleteExam(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	examID := uuid.New()
	examRepo.On("Delete", mock.Anything, examID, orgID).Return(nil)

	err := uc.DeleteExam(context.Background(), examID, orgID)
	require.NoError(t, err)
}

func TestExamUseCase_MyAttempts(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	examID := uuid.New()
	studentID := uuid.New()
	attempts := []*entities.ExamAttempt{
		{ID: uuid.New(), ExamID: examID, StudentID: studentID, Status: "submitted"},
	}
	attemptRepo.On("FindByExamAndStudent", mock.Anything, examID, studentID).Return(attempts, nil)

	result, err := uc.MyAttempts(context.Background(), examID, studentID)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "submitted", result[0].Status)
}

func TestExamUseCase_ListAttempts(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	examID := uuid.New()
	attempts := []*entities.ExamAttempt{
		{ID: uuid.New(), ExamID: examID, Status: "submitted"},
		{ID: uuid.New(), ExamID: examID, Status: "in_progress"},
	}
	attemptRepo.On("FindByExam", mock.Anything, examID).Return(attempts, nil)

	result, err := uc.ListAttempts(context.Background(), examID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// ---- Group CRUD ----

func TestGroupUseCase_GetGroup(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	groupID := uuid.New()
	group := &entities.Group{ID: groupID, OrgID: orgID, Name: "Group A"}
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(group, nil)

	result, err := uc.GetGroup(context.Background(), groupID, orgID)
	require.NoError(t, err)
	assert.Equal(t, "Group A", result.Name)
}

func TestGroupUseCase_ListGroups(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	groups := []entities.Group{
		{ID: uuid.New(), CourseID: &courseID, OrgID: orgID, Name: "G1"},
		{ID: uuid.New(), CourseID: &courseID, OrgID: orgID, Name: "G2"},
	}
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return(groups, nil)

	result, err := uc.ListGroups(context.Background(), courseID, orgID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGroupUseCase_DeleteGroup(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	groupID := uuid.New()
	groupRepo.On("DeleteGroup", mock.Anything, groupID, orgID).Return(nil)

	err := uc.DeleteGroup(context.Background(), groupID, orgID)
	require.NoError(t, err)
}

func TestGroupUseCase_AddSchedule(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	groupID := uuid.New()
	group := &entities.Group{ID: groupID, OrgID: orgID, Name: "G1"}
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(group, nil)
	groupRepo.On("AddSchedule", mock.Anything, mock.AnythingOfType("*entities.GroupSchedule")).Return(nil)

	result, err := uc.AddSchedule(context.Background(), groupID, orgID, dto.CreateGroupScheduleRequest{
		DayOfWeek: 1, StartTime: "09:00", EndTime: "11:00", Location: "Room 101",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.DayOfWeek)
	assert.Equal(t, "09:00", result.StartTime)
}

func TestGroupUseCase_ListSchedules(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	groupID := uuid.New()
	group := &entities.Group{ID: groupID, OrgID: orgID}
	schedules := []entities.GroupSchedule{
		{ID: uuid.New(), GroupID: groupID, DayOfWeek: 1, StartTime: "09:00", EndTime: "11:00"},
	}
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(group, nil)
	groupRepo.On("ListSchedulesByGroup", mock.Anything, groupID).Return(schedules, nil)

	result, err := uc.ListSchedules(context.Background(), groupID, orgID)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGroupUseCase_RemoveMember(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	groupID := uuid.New()
	studentID := uuid.New()
	group := &entities.Group{ID: groupID, OrgID: orgID}
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(group, nil)
	groupRepo.On("RemoveMember", mock.Anything, groupID, studentID).Return(nil)

	err := uc.RemoveMember(context.Background(), groupID, studentID, orgID)
	require.NoError(t, err)
}

func TestGroupUseCase_ListMembers(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	groupID := uuid.New()
	group := &entities.Group{ID: groupID, OrgID: orgID}
	members := []entities.GroupMember{
		{ID: uuid.New(), GroupID: groupID, StudentID: uuid.New(), OrgID: orgID},
	}
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(group, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return(members, nil)

	result, err := uc.ListMembers(context.Background(), groupID, orgID)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

// ---- Progress ----

func TestProgressUseCase_GetStudentProgress(t *testing.T) {
	progressRepo := new(mocks.MockLessonProgressRepository)
	lessonRepo := new(mocks.MockLessonRepository)
	aiService := new(mocks.MockAIService)
	uc := NewProgressUseCase(progressRepo, lessonRepo, aiService, new(mocks.MockModuleRepository), new(mocks.MockQuizRepository), new(mocks.MockQuizAttemptRepository))

	orgID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	progresses := []entities.LessonProgress{
		{ID: uuid.New(), UserID: userID, LessonID: uuid.New(), OrgID: orgID, CompletedAt: &now},
	}
	progressRepo.On("FindByUser", mock.Anything, userID, orgID).Return(progresses, nil)

	result, err := uc.GetStudentProgress(context.Background(), userID, orgID)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

// ---- Quiz Attempt ----

func TestQuizAttemptUseCase_GetMyAttempt(t *testing.T) {
	attemptRepo := new(mocks.MockQuizAttemptRepository)
	quizRepo := new(mocks.MockQuizRepository)
	uc := NewQuizAttemptUseCase(attemptRepo, quizRepo, new(mocks.MockLessonRepository), new(mocks.MockAIService))

	quizID := uuid.New()
	studentID := uuid.New()
	attempt := &entities.QuizAttempt{
		ID: uuid.New(), QuizID: quizID, StudentID: studentID, Score: 1, MaxScore: 1,
	}
	attemptRepo.On("FindByQuizAndStudent", mock.Anything, quizID, studentID).Return(attempt, nil)

	result, err := uc.GetMyAttempt(context.Background(), quizID, studentID)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Score)
}

// ---- File Attachment ----

func TestFileAttachmentUseCase_ListByRef(t *testing.T) {
	fileRepo := new(mocks.MockFileAttachmentRepository)
	uc := NewFileAttachmentUseCase(fileRepo)

	refID := uuid.New()
	attachments := []*entities.FileAttachment{
		{ID: uuid.New(), RefType: "lesson", RefID: refID, OriginalName: "file.pdf"},
	}
	fileRepo.On("FindByRef", mock.Anything, "lesson", refID).Return(attachments, nil)

	result, err := uc.ListByRef(context.Background(), "lesson", refID)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "file.pdf", result[0].OriginalName)
}

// ---- Course list/delete ----

func TestCourseUseCase_ListCourses(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, new(mocks.MockGroupRepository), new(mocks.MockCourseTeacherRepository))

	orgID := uuid.New()
	courses := []entities.Course{
		{ID: uuid.New(), OrgID: orgID, Title: "C1", Status: entities.StatusDraft},
		{ID: uuid.New(), OrgID: orgID, Title: "C2", Status: entities.StatusPublished},
	}
	courseRepo.On("FindByOrg", mock.Anything, orgID, 1, 10).Return(courses, int64(2), nil)

	result, total, err := uc.ListCourses(context.Background(), orgID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
}

func TestCourseUseCase_DeleteCourse(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, new(mocks.MockGroupRepository), new(mocks.MockCourseTeacherRepository))

	orgID := uuid.New()
	courseID := uuid.New()
	courseRepo.On("Delete", mock.Anything, courseID, orgID).Return(nil)

	err := uc.DeleteCourse(context.Background(), courseID, orgID)
	require.NoError(t, err)
}

// ---- Group UpdateGroup / GetStudentGroup ----

func TestGroupUseCase_UpdateGroup(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	groupID := uuid.New()
	group := &entities.Group{ID: groupID, OrgID: orgID, Name: "Old Name", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(group, nil)
	groupRepo.On("UpdateGroup", mock.Anything, mock.AnythingOfType("*entities.Group")).Return(nil)

	result, err := uc.UpdateGroup(context.Background(), groupID, orgID, dto.UpdateGroupRequest{Name: "New Name"})
	require.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
}

func TestGroupUseCase_GetStudentGroup(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	group := &entities.Group{ID: uuid.New(), CourseID: &courseID, OrgID: orgID, Name: "G1"}
	groupRepo.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).Return(group, nil)

	result, err := uc.GetStudentGroup(context.Background(), courseID, studentID, orgID)
	require.NoError(t, err)
	assert.Equal(t, "G1", result.Name)
}

// ---- Exam UpdateExam ----

func TestExamUseCase_UpdateExam_Success(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	exam.CreatedAt = time.Now()
	exam.UpdatedAt = time.Now()

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	examRepo.On("UpdateWithAttemptGuard", mock.Anything, mock.AnythingOfType("*entities.Exam"), false).Return(true, nil)

	newTitle := "Updated Exam"
	result, err := uc.UpdateExam(context.Background(), exam.ID, orgID, dto.UpdateExamRequest{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "Updated Exam", result.Title)
}

// ---- GroupSchedule delete ----

func TestGroupUseCase_DeleteSchedule(t *testing.T) {
	groupRepo := new(mocks.MockGroupRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := NewGroupUseCase(groupRepo, courseRepo)

	orgID := uuid.New()
	groupID := uuid.New()
	schedID := uuid.New()
	group := &entities.Group{ID: groupID, OrgID: orgID}
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(group, nil)
	groupRepo.On("DeleteSchedule", mock.Anything, schedID, groupID).Return(nil)

	err := uc.DeleteSchedule(context.Background(), schedID, groupID, orgID)
	require.NoError(t, err)
}

// ---- Exam GradeFileSection FileOnly ----

func TestExamUseCase_GradeFileSection_FileOnly(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, false, true)
	graderID := uuid.New()

	attempt := &entities.ExamAttempt{
		ID:         uuid.New(),
		ExamID:     exam.ID,
		OrgID:      orgID,
		Status:     "submitted",
		FilePath:   "uploads/file.pdf",
		FilePoints: exam.FilePoints,
	}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(nil)

	result, err := uc.GradeFileSection(context.Background(), attempt.ID, graderID, dto.GradeExamFileRequest{Score: 45, Feedback: "Excellent"})
	require.NoError(t, err)
	require.NotNil(t, result.TotalScore)
	assert.Equal(t, 45, *result.TotalScore)
}

// ---- Exam SubmitAttempt_BothSections ----

func TestExamUseCase_SubmitAttempt_BothSections_NoTotalUntilFileGraded(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, true)
	studentID := uuid.New()

	attempt := &entities.ExamAttempt{
		ID:        uuid.New(),
		ExamID:    exam.ID,
		StudentID: studentID,
		OrgID:     orgID,
		Status:    "in_progress",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	correctAnswerID := exam.Questions[0].Answers[0].ID.String()
	questionID := exam.Questions[0].ID.String()

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SubmitIfInProgress", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(true, nil)

	// Submit with file path + MCQ answers — total should NOT be set yet (waiting for file grade)
	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID,
		[]dto.ExamMCQAnswerInput{{QuestionID: questionID, AnswerID: correctAnswerID}},
		"uploads/submission.pdf")
	require.NoError(t, err)
	assert.Equal(t, "submitted", result.Status)
	assert.Nil(t, result.TotalScore) // not yet graded
}

// ---- Exam SubmitAttempt_Partial ----

func TestExamUseCase_SubmitAttempt_Partial(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	// 2-question exam: student answers 1 correctly
	q2ID := uuid.New()
	a3ID := uuid.New()
	a4ID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	exam.Questions = append(exam.Questions, entities.ExamQuestion{
		ID:       q2ID,
		Question: "Second question",
		Position: 2,
		Answers: []entities.ExamAnswer{
			{ID: a3ID, QuestionID: q2ID, Answer: "Right", IsCorrect: true},
			{ID: a4ID, QuestionID: q2ID, Answer: "Wrong", IsCorrect: false},
		},
	})
	studentID := uuid.New()

	attempt := &entities.ExamAttempt{
		ID:        uuid.New(),
		ExamID:    exam.ID,
		StudentID: studentID,
		OrgID:     orgID,
		Status:    "in_progress",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	correctAnswerID := exam.Questions[0].Answers[0].ID.String()
	questionID1 := exam.Questions[0].ID.String()
	// Only answer q1, not q2

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SubmitIfInProgress", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(true, nil)

	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID,
		[]dto.ExamMCQAnswerInput{{QuestionID: questionID1, AnswerID: correctAnswerID}}, "")
	require.NoError(t, err)
	// 1 correct out of 2 questions, MCQPoints=100 => score = (1 * 100) / 2 = 50
	assert.Equal(t, 50, *result.MCQScore)
}

// ---- Exam UpdateExam DisableMCQ removes questions ----

func TestExamUseCase_UpdateExam_DisableMCQ_RemovesQuestions(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, true) // both enabled

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	examRepo.On("UpdateWithAttemptGuard", mock.Anything, mock.AnythingOfType("*entities.Exam"), true).Return(true, nil)

	mcqDisabled := false
	result, err := uc.UpdateExam(context.Background(), exam.ID, orgID, dto.UpdateExamRequest{MCQEnabled: &mcqDisabled})
	require.NoError(t, err)
	assert.False(t, result.MCQEnabled)
	assert.Empty(t, result.Questions)
}

// ---- Exam CreateExam with BothSections ----

func TestExamUseCase_CreateExam_Success_BothSections(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	course := &entities.Course{ID: courseID, OrgID: orgID}

	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(course, nil)
	examRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Exam")).Return(nil)

	resp, err := uc.CreateExam(context.Background(), courseID, orgID, makeCreateExamReq(true, true))
	require.NoError(t, err)
	assert.True(t, resp.MCQEnabled)
	assert.True(t, resp.FileEnabled)
	assert.Equal(t, 150, resp.TotalPoints) // 100 + 50
}

// ---- Exam UpdateExam BothSections disabled => validation ----

func TestExamUseCase_UpdateExam_BothSectionsDisabled_ReturnsValidation(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false) // only MCQ enabled

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("HasAttempts", mock.Anything, exam.ID).Return(false, nil)

	// Disable MCQ (only enabled section) without enabling file
	mcqDisabled := false
	_, err := uc.UpdateExam(context.Background(), exam.ID, orgID, dto.UpdateExamRequest{MCQEnabled: &mcqDisabled})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.HTTPStatus)
}
