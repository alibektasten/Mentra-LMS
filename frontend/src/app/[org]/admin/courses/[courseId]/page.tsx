'use client'

import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { ChevronLeft, BookOpen, LayoutGrid, BarChart3, Users, Globe, FileEdit, Megaphone, GraduationCap } from 'lucide-react'
import { useCourse, useUpdateCourse } from '@/lib/queries/courses.queries'
import { courseSchema, CourseFormData } from '@/lib/validators/course.schema'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { CourseBuilder } from '@/components/courses/CourseBuilder'
import { EnrollmentManager } from '@/components/courses/EnrollmentManager'
import { GradeBook } from '@/components/courses/GradeBook'
import { AnnouncementList } from '@/components/courses/AnnouncementList'
import { AnnouncementForm } from '@/components/courses/AnnouncementForm'
import { cn } from '@/lib/utils/cn'
import { CourseTeachersSection } from './CourseTeachersSection'
import { CourseGroupsSection } from './CourseGroupsSection'

export default function AdminCourseDetailPage() {
  const { org, courseId } = useParams<{ org: string; courseId: string }>()
  const { data: course, isLoading } = useCourse(courseId)
  const updateCourse = useUpdateCourse()

  const { register, handleSubmit, setValue, watch, reset, formState: { errors, isSubmitting, isDirty } } = useForm<CourseFormData>({
    resolver: zodResolver(courseSchema),
    values: course ? { title: course.title, description: course.description, status: course.status } : undefined,
  })

  const onSubmit = async (data: CourseFormData) => {
    try {
      await updateCourse.mutateAsync({ id: courseId, input: data })
      toast.success('Course updated')
      reset(data)
    } catch {
      toast.error('Failed to update course')
    }
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl space-y-4 py-6">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-8 w-72" />
        <Skeleton className="h-4 w-full max-w-lg" />
        <Skeleton className="h-10 w-80 mt-6" />
        <Skeleton className="h-64 rounded-xl" />
      </div>
    )
  }

  const status = course?.status ?? 'draft'

  return (
    <div className="max-w-4xl py-2">
      {/* Back link */}
      <Link
        href={`/${org}/admin/courses`}
        className="inline-flex items-center gap-1 text-xs text-[#9b9b9b] hover:text-[#1a1a1a] transition-colors mb-5"
      >
        <ChevronLeft className="h-3.5 w-3.5" />
        All Courses
      </Link>

      {/* Course header */}
      <div className="mb-7">
        <div className="flex items-center gap-2.5 mb-1.5">
          <h1 className="text-[1.6rem] font-bold tracking-tight text-[#1a1a1a] leading-tight truncate">
            {course?.title}
          </h1>
          <span className={cn(
            'shrink-0 inline-flex items-center gap-1 text-[10px] font-semibold uppercase tracking-widest px-2 py-1 rounded-full border',
            status === 'published'
              ? 'text-emerald-700 bg-emerald-50 border-emerald-200'
              : status === 'archived'
                ? 'text-slate-700 bg-slate-100 border-slate-200'
                : 'text-amber-700 bg-amber-50 border-amber-200',
          )}>
            {status === 'published' ? <Globe className="h-2.5 w-2.5" /> : <FileEdit className="h-2.5 w-2.5" />}
            {status}
          </span>
        </div>
        {course?.description && (
          <p className="text-sm text-[#6b6b6b] leading-relaxed max-w-2xl">{course.description}</p>
        )}
      </div>

      {/* Tabs */}
      <Tabs defaultValue="details">
        <TabsList className="mb-6 bg-[#f0eeeb] border border-[#e4e2de] h-auto p-1">
          <TabsTrigger
            value="details"
            className="flex items-center gap-1.5 data-[state=active]:bg-white data-[state=active]:shadow-sm"
          >
            <FileEdit className="h-3.5 w-3.5" />
            Details
          </TabsTrigger>
          <TabsTrigger
            value="content"
            className="flex items-center gap-1.5 data-[state=active]:bg-white data-[state=active]:shadow-sm"
          >
            <LayoutGrid className="h-3.5 w-3.5" />
            Content
          </TabsTrigger>
          <TabsTrigger
            value="announcements"
            className="flex items-center gap-1.5 data-[state=active]:bg-white data-[state=active]:shadow-sm"
          >
            <Megaphone className="h-3.5 w-3.5" />
            Announcements
          </TabsTrigger>
          <TabsTrigger
            value="groups"
            className="flex items-center gap-1.5 data-[state=active]:bg-white data-[state=active]:shadow-sm"
          >
            <Users className="h-3.5 w-3.5" />
            Groups
          </TabsTrigger>
          <TabsTrigger
            value="teachers"
            className="flex items-center gap-1.5 data-[state=active]:bg-white data-[state=active]:shadow-sm"
          >
            <GraduationCap className="h-3.5 w-3.5" />
            Teachers
          </TabsTrigger>
          <TabsTrigger
            value="students"
            className="flex items-center gap-1.5 data-[state=active]:bg-white data-[state=active]:shadow-sm"
          >
            <Users className="h-3.5 w-3.5" />
            Students
          </TabsTrigger>
          <TabsTrigger
            value="grades"
            className="flex items-center gap-1.5 data-[state=active]:bg-white data-[state=active]:shadow-sm"
          >
            <BarChart3 className="h-3.5 w-3.5" />
            Grades
          </TabsTrigger>
        </TabsList>

        {/* Details tab */}
        <TabsContent value="details">
          <div className="rounded-xl border border-[#e4e2de] bg-white p-6 max-w-2xl shadow-sm">
            <div className="flex items-center gap-2 mb-6">
              <div className="h-8 w-8 rounded-lg bg-[#f0eeeb] flex items-center justify-center">
                <BookOpen className="h-4 w-4 text-[#6b6b6b]" />
              </div>
              <div>
                <p className="text-sm font-semibold text-[#1a1a1a]">Course Settings</p>
                <p className="text-xs text-[#9b9b9b]">Edit title, description, and publication status</p>
              </div>
            </div>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              <div className="space-y-1.5">
                <Label htmlFor="title" className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">Title</Label>
                <Input
                  id="title"
                  {...register('title')}
                  className="border-[#e4e2de] focus-visible:ring-[#059669] focus-visible:ring-1"
                />
                {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="description" className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">Description</Label>
                <Textarea
                  id="description"
                  rows={4}
                  {...register('description')}
                  className="border-[#e4e2de] focus-visible:ring-[#059669] focus-visible:ring-1 resize-none"
                />
                {errors.description && <p className="text-xs text-destructive">{errors.description.message}</p>}
              </div>
              <div className="space-y-1.5">
                <Label className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">Status</Label>
                <Select value={watch('status')} onValueChange={(v) => setValue('status', v as 'draft' | 'published' | 'archived')}>
                  <SelectTrigger className="w-44 border-[#e4e2de]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="draft">
                      <span className="flex items-center gap-2">
                        <span className="h-1.5 w-1.5 rounded-full bg-amber-400 inline-block" />
                        Draft
                      </span>
                    </SelectItem>
                    <SelectItem value="published">
                      <span className="flex items-center gap-2">
                        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 inline-block" />
                        Published
                      </span>
                    </SelectItem>
                    <SelectItem value="archived">
                      <span className="flex items-center gap-2">
                        <span className="h-1.5 w-1.5 rounded-full bg-slate-400 inline-block" />
                        Archived
                      </span>
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="pt-1 border-t border-[#f0eeeb]">
                <Button
                  type="submit"
                  disabled={isSubmitting || !isDirty}
                  className="bg-[#059669] hover:bg-[#047857] text-white"
                >
                  {isSubmitting ? 'Saving…' : 'Save Changes'}
                </Button>
              </div>
            </form>
          </div>
        </TabsContent>

        {/* Content tab */}
        <TabsContent value="content">
          <CourseBuilder courseId={courseId} />
        </TabsContent>

        {/* Announcements tab */}
        <TabsContent value="announcements">
          <div className="rounded-xl border border-[#e4e2de] bg-white p-6 shadow-sm">
            <div className="flex items-center justify-between mb-5">
              <div className="flex items-center gap-2">
                <div className="h-8 w-8 rounded-lg bg-[#f0eeeb] flex items-center justify-center">
                  <Megaphone className="h-4 w-4 text-[#6b6b6b]" />
                </div>
                <div>
                  <p className="text-sm font-semibold text-[#1a1a1a]">Announcements</p>
                  <p className="text-xs text-[#9b9b9b]">Post updates for enrolled students</p>
                </div>
              </div>
              <AnnouncementForm courseId={courseId} />
            </div>
            <AnnouncementList courseId={courseId} canDelete />
          </div>
        </TabsContent>

        {/* Groups tab */}
        <TabsContent value="groups">
          <CourseGroupsSection courseId={courseId} />
        </TabsContent>

        {/* Teachers tab */}
        <TabsContent value="teachers">
          <CourseTeachersSection courseId={courseId} />
        </TabsContent>

        {/* Students tab */}
        <TabsContent value="students">
          <div className="rounded-xl border border-[#e4e2de] bg-white p-6 shadow-sm">
            <div className="flex items-center gap-2 mb-5">
              <div className="h-8 w-8 rounded-lg bg-[#f0eeeb] flex items-center justify-center">
                <Users className="h-4 w-4 text-[#6b6b6b]" />
              </div>
              <div>
                <p className="text-sm font-semibold text-[#1a1a1a]">Enrolled Students</p>
                <p className="text-xs text-[#9b9b9b]">Manage enrollment for this course</p>
              </div>
            </div>
            <EnrollmentManager courseId={courseId} />
          </div>
        </TabsContent>

        {/* Grades tab */}
        <TabsContent value="grades">
          <div className="rounded-xl border border-[#e4e2de] bg-white p-6 shadow-sm">
            <div className="flex items-center gap-2 mb-5">
              <div className="h-8 w-8 rounded-lg bg-[#f0eeeb] flex items-center justify-center">
                <BarChart3 className="h-4 w-4 text-[#6b6b6b]" />
              </div>
              <div>
                <p className="text-sm font-semibold text-[#1a1a1a]">Gradebook</p>
                <p className="text-xs text-[#9b9b9b]">Assignments and quizzes by module</p>
              </div>
            </div>
            <GradeBook courseId={courseId} />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
