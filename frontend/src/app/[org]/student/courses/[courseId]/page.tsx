'use client'

import { useParams } from 'next/navigation'
import Link from 'next/link'
import {
  ChevronLeft, BookOpen, Video, HelpCircle, ClipboardList,
  CheckCircle2, Circle, ChevronDown, ChevronRight, FileIcon, Link2, Megaphone, Users, Calendar, Clock, MapPin, GraduationCap,
} from 'lucide-react'
import { useState } from 'react'
import { useCourse } from '@/lib/queries/courses.queries'
import { useModules } from '@/lib/queries/modules.queries'
import { useLessons } from '@/lib/queries/lessons.queries'
import { useAssignments } from '@/lib/queries/assignments.queries'
import { useProgress } from '@/lib/queries/progress.queries'
import { useAnnouncements } from '@/lib/queries/announcements.queries'
import { useMyGroup, useGroupSchedules } from '@/lib/queries/groups.queries'
import { useExams, useMyAttempts } from '@/lib/queries/exams.queries'
import { ExamListItemDTO } from '@/types/exam'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { MyGrades } from '@/components/courses/MyGrades'
import { UpcomingDeadlines } from '@/components/courses/UpcomingDeadlines'
import { cn } from '@/lib/utils/cn'
import { useI18n } from '@/lib/i18n'
import { completedLessonIDs, lessonCompletion } from '@/lib/progress/course-lesson-progress'

const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const activityColor: Record<string, string> = {
  video: 'bg-sky-400',
  text: 'bg-violet-400',
  quiz: 'bg-amber-400',
  assignment: 'bg-orange-400',
  pdf: 'bg-rose-400',
  link: 'bg-teal-400',
}

const ActivityIcon: Record<string, React.ElementType> = {
  video: Video,
  text: BookOpen,
  quiz: HelpCircle,
  assignment: ClipboardList,
  pdf: FileIcon,
  link: Link2,
}

function ModuleSection({
  courseId, moduleId, moduleTitle, moduleIndex, org, completedLessonIds,
}: {
  courseId: string
  moduleId: string
  moduleTitle: string
  moduleIndex: number
  org: string
  completedLessonIds: Set<string>
}) {
  const [open, setOpen] = useState(true)
  const { data: lessonsData, isLoading: loadingLessons } = useLessons(moduleId)
  const { data: assignmentsData, isLoading: loadingAssignments } = useAssignments(courseId, moduleId)

  const lessons = Array.isArray(lessonsData) ? lessonsData : []
  const assignments = Array.isArray(assignmentsData) ? assignmentsData : []

  const sortedLessons = [...lessons].sort((a, b) => a.order - b.order)
  const isLoading = loadingLessons || loadingAssignments

  // Assignments are listed in this module, but lesson completion only tracks lessons.
  const { completed, total } = lessonCompletion(sortedLessons.map((lesson) => lesson.id), completedLessonIds)

  return (
    <div className="rounded-xl border border-[#e4e2de] bg-white overflow-hidden shadow-sm">
      {/* Section header */}
      <button
        onClick={() => setOpen((o) => !o)}
        className="w-full flex items-center gap-3 px-5 py-4 hover:bg-[#f7f6f3] transition-colors text-left"
      >
        <span className="flex h-6 w-6 items-center justify-center rounded-full bg-[#1a1a1a] text-[10px] font-bold text-white shrink-0">
          {moduleIndex}
        </span>
        <span className="flex-1 text-sm font-semibold text-[#1a1a1a]">{moduleTitle}</span>
        {total > 0 && (
          <span className="text-xs text-[#9b9b9b] mr-2">
            {completed}/{total}
          </span>
        )}
        {open ? <ChevronDown className="h-4 w-4 text-[#9b9b9b]" /> : <ChevronRight className="h-4 w-4 text-[#9b9b9b]" />}
      </button>

      {open && (
        <div className="border-t border-[#e4e2de]">
          {isLoading ? (
            <div className="p-4 space-y-2">
              {[1, 2].map((i) => <Skeleton key={i} className="h-10 rounded" />)}
            </div>
          ) : sortedLessons.length === 0 && assignments.length === 0 ? (
            <p className="px-5 py-4 text-xs text-[#9b9b9b]">No activities yet.</p>
          ) : (
            <div className="divide-y divide-[#f0eeeb]">
              {sortedLessons.map((lesson) => {
                const done = completedLessonIds.has(lesson.id)
                const Icon = ActivityIcon[lesson.type] ?? BookOpen
                const color = activityColor[lesson.type] ?? 'bg-gray-300'
                return (
                  <Link
                    key={lesson.id}
                    href={`/${org}/student/courses/${courseId}/lessons/${lesson.id}`}
                    className="group flex items-center gap-0 hover:bg-[#f7f6f3] transition-colors"
                  >
                    <div className={cn('w-1 self-stretch shrink-0', color)} />
                    <div className="flex items-center gap-3 px-4 py-3 flex-1 min-w-0">
                      {done
                        ? <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0" />
                        : <Circle className="h-4 w-4 text-[#d4d2ce] shrink-0" />
                      }
                      <Icon className="h-3.5 w-3.5 text-[#9b9b9b] shrink-0" />
                      <span className={cn(
                        'flex-1 text-sm truncate',
                        done ? 'text-[#9b9b9b] line-through' : 'text-[#1a1a1a] group-hover:text-[#059669]',
                      )}>
                        {lesson.title}
                      </span>
                      <span className="text-[10px] font-medium px-1.5 py-0.5 rounded capitalize text-[#9b9b9b] bg-[#f0eeeb] shrink-0">
                        {lesson.type}
                      </span>
                    </div>
                  </Link>
                )
              })}
              {assignments.map((assignment) => {
                const dueDate = assignment.due_date ? new Date(assignment.due_date) : null
                const overdue = dueDate && new Date() > dueDate
                return (
                  <Link
                    key={assignment.id}
                    href={`/${org}/student/courses/${courseId}/assignments/${assignment.id}`}
                    className="group flex items-center gap-0 hover:bg-[#f7f6f3] transition-colors"
                  >
                    <div className="w-1 self-stretch shrink-0 bg-orange-400" />
                    <div className="flex items-center gap-3 px-4 py-3 flex-1 min-w-0">
                      <Circle className="h-4 w-4 text-[#d4d2ce] shrink-0" />
                      <ClipboardList className="h-3.5 w-3.5 text-[#9b9b9b] shrink-0" />
                      <span className="flex-1 text-sm truncate text-[#1a1a1a] group-hover:text-[#059669]">
                        {assignment.title}
                      </span>
                      <span className="text-xs text-[#9b9b9b] shrink-0">{assignment.max_points}pts</span>
                      {dueDate && (
                        <span className={cn(
                          'text-[10px] px-1.5 py-0.5 rounded shrink-0',
                          overdue ? 'text-red-600 bg-red-50' : 'text-[#9b9b9b] bg-[#f0eeeb]',
                        )}>
                          {overdue ? 'Overdue' : `Due ${dueDate.toLocaleDateString()}`}
                        </span>
                      )}
                    </div>
                  </Link>
                )
              })}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function ExamCard({ exam, org, courseId }: { exam: ExamListItemDTO; org: string; courseId: string }) {
  const { data: attempts } = useMyAttempts(exam.id)
  const attemptList = attempts ?? []
  const lastAttempt = [...attemptList].reverse().find((a) => a.status !== 'expired')
  const dueDate = exam.due_date ? new Date(exam.due_date) : null

  return (
    <Link
      href={`/${org}/student/courses/${courseId}/exams/${exam.id}`}
      className="group flex items-center gap-0 rounded-xl border border-[#e4e2de] bg-white hover:bg-[#f7f6f3] transition-colors shadow-sm overflow-hidden"
    >
      <div className="w-1 self-stretch bg-purple-400 shrink-0" />
      <div className="flex items-center gap-3 px-4 py-3 flex-1 min-w-0">
        <GraduationCap className="h-4 w-4 text-[#9b9b9b] shrink-0" />
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-[#1a1a1a] truncate group-hover:text-[#059669]">{exam.title}</p>
          <p className="text-xs text-[#9b9b9b]"><span className="capitalize font-semibold text-[#6b6b6b]">{exam.assessment_type === 'other' ? 'Exam' : exam.assessment_type}</span> · {exam.duration_minutes} min · {exam.total_points} pts</p>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {dueDate && (
            <span className="text-[10px] px-1.5 py-0.5 rounded bg-[#f0eeeb] text-[#9b9b9b]">
              Due {dueDate.toLocaleDateString()}
            </span>
          )}
          {lastAttempt && (
            <span className={cn(
              'text-[10px] font-semibold uppercase tracking-wide px-1.5 py-0.5 rounded border',
              lastAttempt.status === 'submitted' ? 'bg-sky-50 text-sky-700 border-sky-200' :
              'bg-amber-50 text-amber-700 border-amber-200'
            )}>
              {lastAttempt.status === 'submitted' && lastAttempt.total_score != null
                ? `${lastAttempt.total_score} pts`
                : lastAttempt.status}
            </span>
          )}
        </div>
      </div>
    </Link>
  )
}

function ExamListSection({ courseId, org }: { courseId: string; org: string }) {
  const { t } = useI18n()
  const { data: exams, isLoading } = useExams(courseId)
  const examList = exams ?? []

  if (isLoading) return <div className="space-y-2">{[1, 2].map((i) => <Skeleton key={i} className="h-14 rounded-xl" />)}</div>

  if (examList.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-14 rounded-xl border border-dashed border-[#e4e2de]">
        <GraduationCap className="h-7 w-7 text-[#d4d2ce] mb-2" />
        <p className="text-sm text-[#9b9b9b]">{t('noExamsYet')}</p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {examList.filter((exam) => exam.assessment_type === 'other' || !!exam.due_date).map((exam) => (
        <ExamCard key={exam.id} exam={exam} org={org} courseId={courseId} />
      ))}
    </div>
  )
}

function MyGroupSection({ courseId }: { courseId: string }) {
  const { t } = useI18n()
  const { data: group, isLoading, isError } = useMyGroup(courseId)
  const { data: schedulesData } = useGroupSchedules(courseId, group?.id ?? '', )

  if (isLoading) return <Skeleton className="h-20 rounded-xl" />
  if (isError || !group) return null

  const schedules = Array.isArray(schedulesData) ? schedulesData : []

  return (
    <div className="rounded-xl border border-[#e4e2de] bg-white shadow-sm overflow-hidden">
      <div className="flex items-center gap-3 px-5 py-4 bg-[#f7f6f3] border-b border-[#e4e2de]">
        <Users className="h-4 w-4 text-[#6b6b6b]" />
        <span className="text-sm font-semibold text-[#1a1a1a]">{t('myGroup')} — {group.name}</span>
      </div>
      {schedules.length > 0 && (
        <div className="px-5 py-4 space-y-2">
          <p className="text-xs font-semibold text-[#9b9b9b] uppercase tracking-wide mb-2">{t('schedule')}</p>
          {schedules.map((s) => (
            <div key={s.id} className="flex items-center gap-4 text-sm">
              <span className="w-8 text-xs font-medium text-[#6b6b6b]">{DAYS[s.day_of_week]}</span>
              <div className="flex items-center gap-1 text-xs text-[#6b6b6b]">
                <Clock className="h-3 w-3" />
                {s.start_time} – {s.end_time}
              </div>
              {s.location && (
                <div className="flex items-center gap-1 text-xs text-[#9b9b9b]">
                  <MapPin className="h-3 w-3" />
                  {s.location}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default function StudentCourseDetailPage() {
  const { t } = useI18n()
  const { org, courseId } = useParams<{ org: string; courseId: string }>()
  const { data: course, isLoading: courseLoading, isError: courseError } = useCourse(courseId)
  const { data: modulesData, isLoading: modulesLoading } = useModules(courseId)
  const { data: progressData, isLoading: progressLoading, isError: progressError } = useProgress(courseId)
  const { data: announcementsData } = useAnnouncements(courseId)

  const modules = Array.isArray(modulesData) ? modulesData : []
  const sortedModules = [...modules].sort((a, b) => a.order - b.order)
  const progressList = Array.isArray(progressData) ? progressData : []
  const completedLessonIds = completedLessonIDs(progressList)
  const announcements = Array.isArray(announcementsData?.data) ? announcementsData.data : []

  if (courseLoading) {
    return (
      <div className="max-w-3xl space-y-4 py-6">
        <Skeleton className="h-7 w-56" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-40 rounded-xl" />
      </div>
    )
  }

  if (courseError || !course) {
    return (
      <div className="max-w-3xl flex flex-col items-center justify-center py-24 text-center">
        <BookOpen className="h-8 w-8 text-[#c9c9c9] mb-3" />
        <p className="text-sm font-semibold text-[#1a1a1a] mb-1">Course unavailable</p>
        <p className="text-xs text-[#9b9b9b] mb-4">This course may no longer be available to you.</p>
        <Link
          href={`/${org}/student/courses`}
          className="inline-flex items-center gap-1.5 text-xs text-[#059669] hover:underline"
        >
          <ChevronLeft className="h-3.5 w-3.5" /> Back to my courses
        </Link>
      </div>
    )
  }

  return (
    <div className="max-w-3xl py-2">
      <Link
        href={`/${org}/student/courses`}
        className="inline-flex items-center gap-1 text-xs text-[#9b9b9b] hover:text-[#1a1a1a] transition-colors mb-5"
      >
        <ChevronLeft className="h-3.5 w-3.5" />
        My Courses
      </Link>

      {/* Course header */}
      <div className="mb-6">
        <h1 className="text-[1.6rem] font-bold tracking-tight text-[#1a1a1a] leading-tight">{course.title}</h1>
        {course.description && (
          <p className="mt-1.5 text-sm text-[#6b6b6b] leading-relaxed">{course.description}</p>
        )}
        {/* Progress indicator */}
        {!progressLoading && !progressError && (
          <div className="mt-3 inline-flex items-center gap-1.5 text-xs text-emerald-700 bg-emerald-50 border border-emerald-200 px-2.5 py-1 rounded-full">
            <CheckCircle2 className="h-3 w-3" />
            {completedLessonIds.size} lesson{completedLessonIds.size !== 1 ? 's' : ''} completed
          </div>
        )}
        {progressError && <p role="alert" className="mt-3 text-xs text-red-700">Progress unavailable. Try refreshing this course.</p>}
      </div>

      <Tabs defaultValue="content">
        <TabsList className="mb-5 bg-[#f0eeeb] border border-[#e4e2de]">
          <TabsTrigger value="content">{t('content')}</TabsTrigger>
          <TabsTrigger value="announcements">
            {t('announcements')}
            {announcements.length > 0 && (
              <span className="ml-1.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-amber-500 text-[9px] font-bold text-white px-1">
                {announcements.length}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="exams">{t('exams')}</TabsTrigger>
          <TabsTrigger value="group">{t('groups')}</TabsTrigger>
          <TabsTrigger value="grades">{t('grades')}</TabsTrigger>
          <TabsTrigger value="deadlines">{t('deadlines')}</TabsTrigger>
        </TabsList>

        <TabsContent value="content">
          {modulesLoading ? (
            <div className="space-y-3">
              {[1, 2].map((i) => <Skeleton key={i} className="h-32 rounded-xl" />)}
            </div>
          ) : sortedModules.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center rounded-xl border border-dashed border-[#e4e2de]">
              <BookOpen className="h-8 w-8 text-[#d4d2ce] mb-2" />
              <p className="text-sm text-[#9b9b9b]">{t('noContentYet')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {sortedModules.map((module, idx) => (
                <ModuleSection
                  key={module.id}
                  courseId={courseId}
                  moduleId={module.id}
                  moduleTitle={module.title}
                  moduleIndex={idx + 1}
                  org={org}
                  completedLessonIds={completedLessonIds}
                />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="announcements">
          {announcements.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center rounded-xl border border-dashed border-[#e4e2de]">
              <Megaphone className="h-7 w-7 text-[#d4d2ce] mb-2" />
              <p className="text-sm text-[#9b9b9b]">{t('noAnnouncementsYet')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {announcements.map((a) => (
                <div key={a.id} className="rounded-xl border border-[#e4e2de] bg-white p-4 shadow-sm">
                  <div className="flex items-start gap-3">
                    <div className="h-8 w-8 rounded-lg bg-amber-50 border border-amber-200 flex items-center justify-center shrink-0 mt-0.5">
                      <Megaphone className="h-4 w-4 text-amber-600" />
                    </div>
                    <div className="min-w-0">
                      <p className="text-sm font-semibold text-[#1a1a1a]">{a.title}</p>
                      <p className="mt-1 text-sm text-[#6b6b6b] leading-relaxed whitespace-pre-wrap">{a.content}</p>
                      <p className="mt-2 text-xs text-[#9b9b9b]">
                        {new Date(a.created_at).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="exams">
          <ExamListSection courseId={courseId} org={org} />
        </TabsContent>

        <TabsContent value="group">
          <MyGroupSection courseId={courseId} />
        </TabsContent>

        <TabsContent value="grades">
          <MyGrades courseId={courseId} />
        </TabsContent>

        <TabsContent value="deadlines">
          <UpcomingDeadlines courseId={courseId} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
