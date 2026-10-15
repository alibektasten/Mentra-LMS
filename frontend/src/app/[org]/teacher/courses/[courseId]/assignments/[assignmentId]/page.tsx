'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import { ChevronLeft, ClipboardCheck } from 'lucide-react'
import { useAssignment } from '@/lib/queries/assignments.queries'
import { SubmissionGrader } from '@/components/courses/SubmissionGrader'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

export default function TeacherAssignmentSubmissionsPage() {
  const { org, courseId, assignmentId } = useParams<{ org: string; courseId: string; assignmentId: string }>()
  const { data: assignment, isLoading, isError, refetch } = useAssignment(assignmentId)
  const courseHref = `/${org}/teacher/courses/${courseId}`

  if (isLoading) {
    return (
      <div className="max-w-4xl space-y-4 py-2">
        <Skeleton className="h-4 w-28" />
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-32 rounded-xl" />
      </div>
    )
  }

  if (isError) {
    return (
      <div className="max-w-3xl space-y-4 py-2">
        <BackToCourse href={courseHref} />
        <div role="alert" className="flex flex-col items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          <p>Could not load this assignment. Check your access and try again.</p>
          <Button type="button" variant="outline" size="sm" onClick={() => void refetch()}>Try again</Button>
        </div>
      </div>
    )
  }

  // Keep the route truthful even if a copied URL combines a valid assignment
  // with another course id. The API remains the authorization boundary.
  if (!assignment || assignment.course_id !== courseId) {
    return (
      <div className="max-w-3xl space-y-4 py-2">
        <BackToCourse href={courseHref} />
        <div role="alert" className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
          This assignment does not belong to the selected course.
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl py-2">
      <BackToCourse href={courseHref} />
      <div className="mb-6 flex items-start gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-700">
          <ClipboardCheck className="h-4 w-4" />
        </div>
        <div className="min-w-0">
          <h1 className="truncate text-xl font-semibold text-[#1a1a1a]">Review submissions: {assignment.title}</h1>
          <p className="mt-1 text-sm text-[#6b6b6b]">Maximum score: {assignment.max_points}</p>
        </div>
      </div>
      <SubmissionGrader
        assignmentId={assignment.id}
        maxPoints={assignment.max_points}
        dueDate={assignment.due_date}
      />
    </div>
  )
}

function BackToCourse({ href }: { href: string }) {
  return (
    <Link
      href={href}
      className="mb-5 inline-flex items-center gap-1 text-xs text-[#9b9b9b] transition-colors hover:text-[#1a1a1a]"
    >
      <ChevronLeft className="h-3.5 w-3.5" />
      Back to Course
    </Link>
  )
}
