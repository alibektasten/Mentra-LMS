'use client'

import { useState } from 'react'
import { useAssignments } from '@/lib/queries/assignments.queries'
import { useModules } from '@/lib/queries/modules.queries'
import { useListSubmissions } from '@/lib/queries/assignments.queries'
import { AssignmentDTO } from '@/types/assignment'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { SubmissionGrader } from './SubmissionGrader'
import { ClipboardList, Download } from 'lucide-react'

interface GradeBookProps {
  courseId: string
}

export function GradeBook({ courseId }: GradeBookProps) {
  const { data: modulesData, isLoading } = useModules(courseId)
  const modules = Array.isArray(modulesData) ? modulesData : []

  if (isLoading) {
    return <Skeleton className="h-40 rounded-lg" />
  }

  if (modules.length === 0) {
    return <p className="text-sm text-ink-muted py-4">No modules with gradeable items yet.</p>
  }

  return (
    <div className="space-y-4">
      {modules.map((module) => (
        <ModuleGrades key={module.id} courseId={courseId} moduleId={module.id} moduleTitle={module.title} />
      ))}
    </div>
  )
}

function ModuleGrades({ courseId, moduleId, moduleTitle }: { courseId: string; moduleId: string; moduleTitle: string }) {
  const { data, isLoading } = useAssignments(courseId, moduleId)
  const assignments = data ?? []

  if (isLoading) return <Skeleton className="h-20 rounded-lg" />
  if (assignments.length === 0) return null

  return (
    <div>
      <h3 className="text-xs font-semibold text-[#9b9b9b] uppercase tracking-wide mb-2">{moduleTitle}</h3>
      <div className="rounded-lg border border-[#e4e2de] divide-y divide-[#f0eeeb]">
        {assignments.map((assignment) => (
          <AssignmentRow key={assignment.id} assignment={assignment} />
        ))}
      </div>
    </div>
  )
}

function AssignmentRow({ assignment }: { assignment: AssignmentDTO }) {
  const [open, setOpen] = useState(false)
  const { data: submissions } = useListSubmissions(assignment.id)

  const total = submissions?.length ?? 0
  const graded = submissions?.filter((s) => s.score !== null).length ?? 0
  const pending = total - graded

  const downloadCSV = () => {
    if (!submissions?.length) return
    const escape = (value: string | number | null) => `"${String(value ?? '').replaceAll('"', '""')}"`
    const rows = [
      ['Student', 'Student ID', 'Assignment', 'Score', 'Max points', 'Status', 'Submitted at', 'Graded at'],
      ...submissions.map((submission) => [
        submission.student_name || 'Student',
        submission.student_id,
        assignment.title,
        submission.score,
        assignment.max_points,
        submission.score === null ? 'Pending review' : 'Graded',
        new Date(submission.submitted_at).toLocaleString(),
        submission.graded_at ? new Date(submission.graded_at).toLocaleString() : '',
      ]),
    ]
    const content = `\uFEFF${rows.map((row) => row.map(escape).join(',')).join('\n')}`
    const url = URL.createObjectURL(new Blob([content], { type: 'text/csv;charset=utf-8' }))
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = `${assignment.title.replaceAll(/[^a-z0-9]+/gi, '-').replace(/(^-|-$)/g, '') || 'assignment'}-grades.csv`
    anchor.click()
    URL.revokeObjectURL(url)
  }

  return (
    <>
      <div className="flex items-center justify-between px-4 py-3">
        <div className="min-w-0">
          <p className="text-sm font-medium text-[#1a1a1a] truncate">{assignment.title}</p>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="text-xs text-[#9b9b9b]">{assignment.max_points} pts</span>
            {total > 0 && (
              <>
                <span className="text-[#d4d0cb]">·</span>
                <span className="text-xs text-[#9b9b9b]">{total} submitted</span>
                {pending > 0 && (
                  <>
                    <span className="text-[#d4d0cb]">·</span>
                    <span className="text-xs font-medium text-amber-600">{pending} pending</span>
                  </>
                )}
                {pending === 0 && graded > 0 && (
                  <>
                    <span className="text-[#d4d0cb]">·</span>
                    <span className="text-xs font-medium text-emerald-600">All graded</span>
                  </>
                )}
              </>
            )}
          </div>
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          {total > 0 && <Button variant="ghost" size="sm" className="h-7 w-7 p-0 text-[#6b6b6b]" onClick={downloadCSV} aria-label={`Export ${assignment.title} grades`} title="Export CSV"><Download className="h-3.5 w-3.5" /></Button>}
          <Button
            variant="outline"
            size="sm"
            className="h-7 text-xs gap-1.5 border-[#e4e2de]"
            onClick={() => setOpen(true)}
          >
            <ClipboardList className="h-3.5 w-3.5" />
            {total > 0 ? 'Grade' : 'Submissions'}
          </Button>
        </div>
      </div>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="text-base">{assignment.title}</DialogTitle>
            <p className="text-xs text-[#9b9b9b]">
              {total} submission{total !== 1 ? 's' : ''} · {assignment.max_points} pts max
            </p>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto pr-1 -mr-1">
            <SubmissionGrader assignmentId={assignment.id} maxPoints={assignment.max_points} dueDate={assignment.due_date} />
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
