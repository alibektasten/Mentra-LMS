'use client'

import { Users, BookOpen, UsersRound } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { EmptyState } from '@/components/shared/EmptyState'
import { Skeleton } from '@/components/ui/skeleton'
import { useTeacherStudents } from '@/lib/queries/teacher.queries'

export default function TeacherStudentsPage() {
  const { data: students, isLoading } = useTeacherStudents()
  if (isLoading) return <div className="space-y-4"><Skeleton className="h-8 w-40" />{[1, 2, 3].map((i) => <Skeleton key={i} className="h-20 w-full" />)}</div>
  return (
    <div className="space-y-6">
      <PageHeader title="Students" description="Students enrolled in your assigned course groups" />
      {!students?.length ? <EmptyState icon={Users} title="No students yet" description="Add students to a group in one of your courses to see them here." /> : <div className="divide-y rounded-xl border border-[#e4e2de] bg-white">
        {students.map((student) => <div key={student.user_id} className="p-4 sm:flex sm:items-center sm:justify-between gap-4">
          <div className="min-w-0"><p className="font-medium text-[#1a1a1a]">{student.name}</p><p className="text-sm text-[#6b6b6b] truncate">{student.email}</p></div>
          <div className="mt-3 sm:mt-0 flex flex-wrap gap-2 text-xs"><span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-1 text-emerald-700"><BookOpen className="h-3 w-3" />{student.courses.join(', ')}</span><span className="inline-flex items-center gap-1 rounded-full bg-stone-100 px-2.5 py-1 text-stone-600"><UsersRound className="h-3 w-3" />{student.groups.join(', ')}</span></div>
        </div>)}
      </div>}
    </div>
  )
}
