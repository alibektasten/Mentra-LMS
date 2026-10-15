package usecases

import (
	"context"
	"sort"
	"time"

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
	examRepo       repositories.ExamRepository
}

func NewTeacherUseCase(courseRepo repositories.CourseRepository, groupRepo repositories.GroupRepository, ctRepo repositories.CourseTeacherRepository, userRepo repositories.UserRepository, assignmentRepo repositories.AssignmentRepository, examRepo repositories.ExamRepository) *TeacherUseCase {
	return &TeacherUseCase{courseRepo: courseRepo, groupRepo: groupRepo, ctRepo: ctRepo, userRepo: userRepo, assignmentRepo: assignmentRepo, examRepo: examRepo}
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

// ListCalendar returns dated coursework for courses assigned to the teacher.
// Weekly group schedules are expanded only inside the requested date range, so
// the API never reveals events from courses the teacher cannot access.
func (uc *TeacherUseCase) ListCalendar(ctx context.Context, teacherID, orgID uuid.UUID, from, to time.Time) ([]dto.TeacherCalendarEventDTO, error) {
	courses, err := uc.assignedCourses(ctx, teacherID, orgID)
	if err != nil {
		return nil, err
	}

	from = calendarDay(from)
	to = calendarDay(to)
	events := make([]dto.TeacherCalendarEventDTO, 0)
	for _, course := range courses {
		assignments, err := uc.assignmentRepo.FindByCourse(ctx, course.ID, orgID)
		if err != nil {
			return nil, err
		}
		for _, assignment := range assignments {
			if assignment.DueDate == nil || !isWithinCalendarRange(*assignment.DueDate, from, to) {
				continue
			}
			events = append(events, dto.TeacherCalendarEventDTO{
				ID:          "assignment:" + assignment.ID.String(),
				ItemID:      assignment.ID.String(),
				ItemType:    "assignment",
				Title:       assignment.Title,
				CourseID:    course.ID.String(),
				CourseTitle: course.Title,
				StartAt:     *assignment.DueDate,
			})
		}

		exams, err := uc.examRepo.FindByCourse(ctx, course.ID, orgID)
		if err != nil {
			return nil, err
		}
		for _, exam := range exams {
			if exam.DueDate == nil || !isWithinCalendarRange(*exam.DueDate, from, to) {
				continue
			}
			events = append(events, dto.TeacherCalendarEventDTO{
				ID:             "exam:" + exam.ID.String(),
				ItemID:         exam.ID.String(),
				ItemType:       "exam",
				Title:          exam.Title,
				CourseID:       course.ID.String(),
				CourseTitle:    course.Title,
				AssessmentType: exam.AssessmentType,
				StartAt:        *exam.DueDate,
			})
		}

		groups, err := uc.groupRepo.ListGroupsByCourse(ctx, course.ID, orgID)
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			schedules, err := uc.groupRepo.ListSchedulesByGroup(ctx, group.ID)
			if err != nil {
				return nil, err
			}
			for _, schedule := range schedules {
				events = append(events, expandGroupSchedule(schedule, group, course, from, to)...)
			}
		}
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartAt.Equal(events[j].StartAt) {
			if events[i].ItemType == events[j].ItemType {
				return events[i].Title < events[j].Title
			}
			return events[i].ItemType < events[j].ItemType
		}
		return events[i].StartAt.Before(events[j].StartAt)
	})
	return events, nil
}

func calendarDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func isWithinCalendarRange(value, from, to time.Time) bool {
	day := calendarDay(value)
	return !day.Before(from) && !day.After(to)
}

func expandGroupSchedule(schedule entities.GroupSchedule, group entities.Group, course entities.Course, from, to time.Time) []dto.TeacherCalendarEventDTO {
	startClock, startErr := time.Parse("15:04", schedule.StartTime)
	endClock, endErr := time.Parse("15:04", schedule.EndTime)
	if startErr != nil || endErr != nil || schedule.DayOfWeek < 0 || schedule.DayOfWeek > 6 {
		// Older or manually inserted malformed schedule rows must not make the
		// whole calendar unavailable.
		return nil
	}

	groupID, groupName := group.ID.String(), group.Name
	events := make([]dto.TeacherCalendarEventDTO, 0)
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		if int(day.Weekday()) != schedule.DayOfWeek {
			continue
		}
		startAt := time.Date(day.Year(), day.Month(), day.Day(), startClock.Hour(), startClock.Minute(), 0, 0, time.UTC)
		endAt := time.Date(day.Year(), day.Month(), day.Day(), endClock.Hour(), endClock.Minute(), 0, 0, time.UTC)
		if !endAt.After(startAt) {
			// Support sessions that finish after midnight.
			endAt = endAt.AddDate(0, 0, 1)
		}
		events = append(events, dto.TeacherCalendarEventDTO{
			ID:          "class:" + schedule.ID.String() + ":" + day.Format("2006-01-02"),
			ItemID:      schedule.ID.String(),
			ItemType:    "class",
			Title:       group.Name,
			CourseID:    course.ID.String(),
			CourseTitle: course.Title,
			GroupID:     &groupID,
			GroupName:   &groupName,
			Location:    schedule.Location,
			StartAt:     startAt,
			EndAt:       &endAt,
		})
	}
	return events
}

func mapKeys(items map[string]struct{}) []string {
	result := make([]string, 0, len(items))
	for item := range items {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}
