package usecases

import (
	"context"
	"sort"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/google/uuid"
)

type TeacherUseCase struct {
	courseRepo     repositories.CourseRepository
	groupRepo      repositories.GroupRepository
	ctRepo         repositories.CourseTeacherRepository
	userRepo       repositories.UserRepository
	assignmentRepo repositories.AssignmentRepository
}

func NewTeacherUseCase(courseRepo repositories.CourseRepository, groupRepo repositories.GroupRepository, ctRepo repositories.CourseTeacherRepository, userRepo repositories.UserRepository, assignmentRepo repositories.AssignmentRepository) *TeacherUseCase {
	return &TeacherUseCase{courseRepo: courseRepo, groupRepo: groupRepo, ctRepo: ctRepo, userRepo: userRepo, assignmentRepo: assignmentRepo}
}

func (uc *TeacherUseCase) assignedCourses(ctx context.Context, teacherID, orgID uuid.UUID) ([]entities.Course, error) {
	ids, err := uc.ctRepo.FindCourseIDsByTeacher(ctx, teacherID, orgID)
	if err != nil || len(ids) == 0 {
		return []entities.Course{}, err
	}
	courses, _, err := uc.courseRepo.FindByIDs(ctx, ids, orgID, 1, len(ids))
	return courses, err
}

func (uc *TeacherUseCase) GetOverview(ctx context.Context, teacherID, orgID uuid.UUID) (*dto.TeacherOverviewDTO, error) {
	courses, err := uc.assignedCourses(ctx, teacherID, orgID)
	if err != nil {
		return nil, err
	}
	overview := &dto.TeacherOverviewDTO{TotalCourses: len(courses)}
	studentIDs := map[uuid.UUID]struct{}{}
	for _, course := range courses {
		if course.Status == entities.StatusPublished {
			overview.PublishedCourses++
		} else {
			overview.DraftCourses++
		}
		groups, err := uc.groupRepo.ListGroupsByCourse(ctx, course.ID, orgID)
		if err != nil {
			return nil, err
		}
		overview.TotalGroups += len(groups)
		for _, group := range groups {
			members, err := uc.groupRepo.ListMembers(ctx, group.ID)
			if err != nil {
				return nil, err
			}
			for _, member := range members {
				studentIDs[member.StudentID] = struct{}{}
			}
		}
	}
	overview.TotalStudents = len(studentIDs)
	return overview, nil
}

func (uc *TeacherUseCase) ListStudents(ctx context.Context, teacherID, orgID uuid.UUID) ([]dto.TeacherStudentDTO, error) {
	courses, err := uc.assignedCourses(ctx, teacherID, orgID)
	if err != nil {
		return nil, err
	}
	type studentInfo struct{ courses, groups map[string]struct{} }
	students := map[uuid.UUID]*studentInfo{}
	for _, course := range courses {
		groups, err := uc.groupRepo.ListGroupsByCourse(ctx, course.ID, orgID)
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			members, err := uc.groupRepo.ListMembers(ctx, group.ID)
			if err != nil {
				return nil, err
			}
			for _, member := range members {
				info := students[member.StudentID]
				if info == nil {
					info = &studentInfo{courses: map[string]struct{}{}, groups: map[string]struct{}{}}
					students[member.StudentID] = info
				}
				info.courses[course.Title] = struct{}{}
				info.groups[group.Name] = struct{}{}
			}
		}
	}
	if len(students) == 0 {
		return []dto.TeacherStudentDTO{}, nil
	}
	ids := make([]uuid.UUID, 0, len(students))
	for id := range students {
		ids = append(ids, id)
	}
	users, err := uc.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make([]dto.TeacherStudentDTO, 0, len(users))
	for _, user := range users {
		info := students[user.ID]
		courses, groups := mapKeys(info.courses), mapKeys(info.groups)
		result = append(result, dto.TeacherStudentDTO{UserID: user.ID.String(), Name: user.Name, Email: user.Email, Courses: courses, Groups: groups})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// ListReviewQueue returns only ungraded submissions from the teacher's own courses.
func (uc *TeacherUseCase) ListReviewQueue(ctx context.Context, teacherID, orgID uuid.UUID) ([]dto.TeacherReviewItemDTO, error) {
	courses, err := uc.assignedCourses(ctx, teacherID, orgID)
	if err != nil {
		return nil, err
	}

	type pendingSubmission struct {
		submission entities.AssignmentSubmission
		assignment entities.Assignment
		course     entities.Course
	}
	pending := make([]pendingSubmission, 0)
	studentIDs := map[uuid.UUID]struct{}{}
	for _, course := range courses {
		assignments, err := uc.assignmentRepo.FindByCourse(ctx, course.ID, orgID)
		if err != nil {
			return nil, err
		}
		for _, assignment := range assignments {
			submissions, err := uc.assignmentRepo.FindSubmissionsByAssignment(ctx, assignment.ID)
			if err != nil {
				return nil, err
			}
			for _, submission := range submissions {
				if submission.Score != nil {
					continue
				}
				pending = append(pending, pendingSubmission{submission: *submission, assignment: *assignment, course: course})
				studentIDs[submission.StudentID] = struct{}{}
			}
		}
	}
	if len(pending) == 0 {
		return []dto.TeacherReviewItemDTO{}, nil
	}

	ids := make([]uuid.UUID, 0, len(studentIDs))
	for id := range studentIDs {
		ids = append(ids, id)
	}
	users, err := uc.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	names := make(map[uuid.UUID]string, len(users))
	for _, user := range users {
		names[user.ID] = user.Name
	}

	result := make([]dto.TeacherReviewItemDTO, 0, len(pending))
	for _, item := range pending {
		result = append(result, dto.TeacherReviewItemDTO{
			SubmissionID: item.submission.ID.String(), AssignmentID: item.assignment.ID.String(),
			CourseID: item.course.ID.String(), CourseTitle: item.course.Title,
			AssignmentTitle: item.assignment.Title, StudentName: names[item.submission.StudentID],
			SubmittedAt: item.submission.SubmittedAt, MaxPoints: item.assignment.MaxPoints,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SubmittedAt.Before(result[j].SubmittedAt) })
	return result, nil
}

func mapKeys(items map[string]struct{}) []string {
	result := make([]string, 0, len(items))
	for item := range items {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}
