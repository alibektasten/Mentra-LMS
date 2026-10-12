'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { Plus, GraduationCap, Clock, Star, Pencil, Trash2, Users, CheckCircle2, Circle } from 'lucide-react'
import { toast } from 'sonner'
import { useExams, useDeleteExam } from '@/lib/queries/exams.queries'
import { ExamListItemDTO } from '@/types/exam'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { ExamFormDialog } from './ExamFormDialog'
import { useI18n } from '@/lib/i18n'

interface ExamBuilderProps {
  courseId: string
}

export function ExamBuilder({ courseId }: ExamBuilderProps) {
  const { t } = useI18n()
  const { org } = useParams<{ org: string }>()
  const { data: exams, isLoading } = useExams(courseId)
  const deleteExam = useDeleteExam(courseId)

  const [formOpen, setFormOpen] = useState(false)
  const [editExam, setEditExam] = useState<ExamListItemDTO | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<ExamListItemDTO | null>(null)

  const handleEdit = (exam: ExamListItemDTO) => {
    setEditExam(exam)
    setFormOpen(true)
  }

  const handleCreate = () => {
    setEditExam(null)
    setFormOpen(true)
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      await deleteExam.mutateAsync(deleteTarget.id)
      toast.success(t('examDeleted'))
    } catch {
      toast.error(t('failedToDeleteExam'))
    } finally {
      setDeleteTarget(null)
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[1, 2].map((i) => <Skeleton key={i} className="h-20 rounded-xl" />)}
      </div>
    )
  }

  const examList = exams ?? []
  const requiredAssessments = [
    { type: 'midterm', label: 'Midterm', weight: '12.5%' },
    { type: 'endterm', label: 'Endterm', weight: '12.5%' },
    { type: 'final', label: 'Final exam', weight: '50%' },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="h-8 w-8 rounded-lg bg-[#f0eeeb] flex items-center justify-center">
            <GraduationCap className="h-4 w-4 text-[#6b6b6b]" />
          </div>
          <div>
            <p className="text-sm font-semibold text-[#1a1a1a]">Exams</p>
            <p className="text-xs text-[#9b9b9b]">Course-level timed exams</p>
          </div>
        </div>
        <Button
          size="sm"
          className="bg-[#059669] hover:bg-[#047857] text-white gap-1.5"
          onClick={handleCreate}
        >
          <Plus className="h-4 w-4" />
          {t('newExam')}
        </Button>
      </div>

      <div className="rounded-xl border border-[#e4e2de] bg-[#fafaf8] p-4">
        <p className="text-sm font-semibold text-[#1a1a1a]">Required assessment plan</p>
        <p className="mt-1 text-xs text-[#77766f]">Assignments contribute 25% of the final grade. Add each required exam to complete the course schedule.</p>
        <div className="mt-3 grid gap-2 sm:grid-cols-3">
          {requiredAssessments.map((item) => {
            const exists = examList.some((exam) => exam.assessment_type === item.type)
            return <div key={item.type} className="flex items-center gap-2 rounded-lg border border-[#e7e6e0] bg-white px-3 py-2 text-xs"><span className={exists ? 'text-emerald-600' : 'text-[#aaa9a2]'}>{exists ? <CheckCircle2 className="h-4 w-4" /> : <Circle className="h-4 w-4" />}</span><span className="font-semibold text-[#45453f]">{item.label}</span><span className="ml-auto text-[#888780]">{item.weight}</span></div>
          })}
        </div>
      </div>

      {examList.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-14 rounded-xl border border-dashed border-[#e4e2de]">
          <GraduationCap className="h-8 w-8 text-[#d4d2ce] mb-2" />
          <p className="text-sm text-[#9b9b9b]">{t('noExamsToCreate')}</p>
        </div>
      ) : (
        <div className="space-y-2">
          {examList.map((exam) => {
            const dueDate = exam.due_date ? new Date(exam.due_date) : null
            return (
              <div
                key={exam.id}
                className="rounded-xl border border-[#e4e2de] bg-white p-4 flex items-center gap-3 shadow-sm"
              >
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-[#1a1a1a] truncate">{exam.title}</p>
                  <div className="mt-1 flex flex-wrap items-center gap-3 text-xs text-[#9b9b9b]">
                    <span className="flex items-center gap-1">
					  {exam.assessment_type !== 'other' && <span className="px-1.5 py-0.5 rounded bg-emerald-50 text-emerald-700 border border-emerald-200 capitalize">{exam.assessment_type}</span>}
                      <Clock className="h-3 w-3" />
                      {exam.duration_minutes} min
                    </span>
                    <span className="flex items-center gap-1">
                      <Star className="h-3 w-3" />
                      {exam.total_points} pts
                    </span>
                    {dueDate && (
                      <span>Due {dueDate.toLocaleDateString()}</span>
                    )}
                    <span className="flex items-center gap-1">
                      {exam.mcq_enabled && <span className="px-1.5 py-0.5 rounded bg-amber-50 text-amber-700 border border-amber-200">MCQ</span>}
                      {exam.file_enabled && <span className="px-1.5 py-0.5 rounded bg-sky-50 text-sky-700 border border-sky-200">File</span>}
                    </span>
                  </div>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <Link href={`/${org}/teacher/courses/${courseId}/exams/${exam.id}/attempts`}>
                    <Button size="sm" variant="ghost" className="h-8 w-8 p-0">
                      <Users className="h-3.5 w-3.5" />
                    </Button>
                  </Link>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-8 w-8 p-0"
                    onClick={() => handleEdit(exam)}
                  >
                    <Pencil className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                    onClick={() => setDeleteTarget(exam)}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      <ExamFormDialog
        courseId={courseId}
        open={formOpen}
        onOpenChange={(v) => { setFormOpen(v); if (!v) setEditExam(null) }}
        editExam={editExam}
      />

      {/* Confirm Delete Dialog */}
      <Dialog open={!!deleteTarget} onOpenChange={(v) => { if (!v) setDeleteTarget(null) }}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete Exam</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-[#6b6b6b]">
            Are you sure you want to delete <strong>{deleteTarget?.title}</strong>? This cannot be undone.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>Cancel</Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleteExam.isPending}
            >
              {deleteExam.isPending ? 'Deleting...' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
