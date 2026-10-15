'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ChevronLeft, GraduationCap } from 'lucide-react'
import { toast } from 'sonner'
import { useExam } from '@/lib/queries/exams.queries'
import { useListAttempts, useGradeExamFile, useGrantExtraAttempt } from '@/lib/queries/exams.queries'
import { ExamAttemptDTO } from '@/types/exam'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { cn } from '@/lib/utils/cn'
import { openPrivateFile } from '@/lib/api/private-files'

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    in_progress: 'bg-amber-50 text-amber-700 border-amber-200',
    submitted: 'bg-sky-50 text-sky-700 border-sky-200',
    expired: 'bg-red-50 text-red-600 border-red-200',
  }
  return (
    <span className={cn('text-[10px] font-semibold uppercase tracking-wide px-1.5 py-0.5 rounded border', styles[status] ?? 'bg-[#f0eeeb] text-[#6b6b6b] border-[#e4e2de]')}>
      {status.replace('_', ' ')}
    </span>
  )
}

export default function ExamAttemptsPage() {
  const { org, courseId, examId } = useParams<{ org: string; courseId: string; examId: string }>()
  const { data: exam, isLoading: examLoading } = useExam(examId)
  const { data: attempts, isLoading: attemptsLoading } = useListAttempts(examId)
  const gradeFile = useGradeExamFile(examId)
  const grantExtra = useGrantExtraAttempt(examId)

  const [gradeTarget, setGradeTarget] = useState<ExamAttemptDTO | null>(null)
  const [gradeScore, setGradeScore] = useState('')
  const [gradeFeedback, setGradeFeedback] = useState('')

  const [grantTarget, setGrantTarget] = useState<ExamAttemptDTO | null>(null)
  const [grantCount, setGrantCount] = useState('1')

  const handleGrade = async () => {
    if (!gradeTarget) return
    const score = parseInt(gradeScore, 10)
    if (isNaN(score) || score < 0) { toast.error('Enter a valid score'); return }
    try {
      await gradeFile.mutateAsync({ attemptID: gradeTarget.id, data: { score, feedback: gradeFeedback } })
      toast.success('File section graded')
      setGradeTarget(null)
    } catch {
      toast.error('Failed to grade')
    }
  }

  const handleGrantExtra = async () => {
    if (!grantTarget) return
    const count = parseInt(grantCount, 10)
    if (isNaN(count) || count < 1) { toast.error('Enter a valid count'); return }
    try {
      await grantExtra.mutateAsync({ student_id: grantTarget.student_id, extra_count: count })
      toast.success('Extra attempt granted')
      setGrantTarget(null)
    } catch {
      toast.error('Failed to grant attempt')
    }
  }

  const isLoading = examLoading || attemptsLoading
  const attemptList = attempts ?? []

  return (
    <div className="max-w-4xl py-2">
      <Link
        href={`/${org}/teacher/courses/${courseId}`}
        className="inline-flex items-center gap-1 text-xs text-[#9b9b9b] hover:text-[#1a1a1a] transition-colors mb-5"
      >
        <ChevronLeft className="h-3.5 w-3.5" />
        Back to Course
      </Link>

      <div className="flex items-center gap-2.5 mb-6">
        <div className="h-9 w-9 rounded-lg bg-[#f0eeeb] flex items-center justify-center">
          <GraduationCap className="h-5 w-5 text-[#6b6b6b]" />
        </div>
        <div>
          {isLoading ? (
            <Skeleton className="h-5 w-48" />
          ) : (
            <h1 className="text-lg font-bold text-[#1a1a1a]">{exam?.title} — Attempts</h1>
          )}
          <p className="text-xs text-[#9b9b9b]">Student submissions and grading</p>
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-14 rounded-xl" />)}
        </div>
      ) : attemptList.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 rounded-xl border border-dashed border-[#e4e2de]">
          <GraduationCap className="h-8 w-8 text-[#d4d2ce] mb-2" />
          <p className="text-sm text-[#9b9b9b]">No attempts yet.</p>
        </div>
      ) : (
        <div className="rounded-xl border border-[#e4e2de] bg-white shadow-sm overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-[#f7f6f3] border-b border-[#e4e2de]">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">Student</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">Status</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">Submitted</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">MCQ</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">File</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide">Total</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#f0eeeb]">
              {attemptList.map((attempt) => (
                <tr key={attempt.id} className="hover:bg-[#fafaf9]">
                  <td className="px-4 py-3 text-xs font-mono text-[#6b6b6b] truncate max-w-[140px]">
                    {attempt.student_id.slice(0, 8)}...
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={attempt.status} />
                  </td>
                  <td className="px-4 py-3 text-xs text-[#9b9b9b]">
                    {attempt.submitted_at
                      ? new Date(attempt.submitted_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
                      : '—'}
                  </td>
                  <td className="px-4 py-3 text-right text-xs">
                    {attempt.mcq_score != null ? `${attempt.mcq_score}/${attempt.mcq_max_score}` : '—'}
                  </td>
                  <td className="px-4 py-3 text-right text-xs">
                    {attempt.file_score != null ? `${attempt.file_score}/${attempt.file_points}` : attempt.file_path ? 'Pending' : '—'}
                  </td>
                  <td className="px-4 py-3 text-right text-xs font-semibold">
                    {attempt.total_score != null ? attempt.total_score : '—'}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 justify-end">
                      {attempt.file_path && <Button size="sm" variant="outline" onClick={() => openPrivateFile('exam-attempts', attempt.id, 'exam-submission').catch(() => toast.error('Не удалось скачать файл'))}>Download File</Button>}
                      {attempt.status === 'submitted' && attempt.file_path && attempt.file_score == null && (
                        <Button
                          size="sm"
                          variant="outline"
                          className="h-7 text-xs"
                          onClick={() => { setGradeTarget(attempt); setGradeScore(''); setGradeFeedback('') }}
                        >
                          Grade File
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-7 text-xs text-[#9b9b9b]"
                        onClick={() => { setGrantTarget(attempt); setGrantCount('1') }}
                      >
                        +Attempt
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Grade File Dialog */}
      <Dialog open={!!gradeTarget} onOpenChange={(v) => { if (!v) setGradeTarget(null) }}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Grade File Submission</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <Label>Score (max {gradeTarget?.file_points})</Label>
              <Input
                type="number"
                min={0}
                max={gradeTarget?.file_points}
                value={gradeScore}
                onChange={(e) => setGradeScore(e.target.value)}
                className="border-[#e4e2de]"
              />
            </div>
            <div className="space-y-1.5">
              <Label>Feedback (optional)</Label>
              <Textarea
                value={gradeFeedback}
                onChange={(e) => setGradeFeedback(e.target.value)}
                rows={3}
                className="border-[#e4e2de] resize-none"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGradeTarget(null)}>Cancel</Button>
            <Button
              onClick={handleGrade}
              disabled={gradeFile.isPending}
              className="bg-[#059669] hover:bg-[#047857] text-white"
            >
              {gradeFile.isPending ? 'Saving...' : 'Save Grade'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Grant Extra Attempt Dialog */}
      <Dialog open={!!grantTarget} onOpenChange={(v) => { if (!v) setGrantTarget(null) }}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Grant Extra Attempts</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <p className="text-sm text-[#6b6b6b]">How many extra attempts for this student?</p>
            <div className="space-y-1.5">
              <Label>Extra Attempts</Label>
              <Input
                type="number"
                min={1}
                value={grantCount}
                onChange={(e) => setGrantCount(e.target.value)}
                className="border-[#e4e2de] w-24"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGrantTarget(null)}>Cancel</Button>
            <Button
              onClick={handleGrantExtra}
              disabled={grantExtra.isPending}
              className="bg-[#059669] hover:bg-[#047857] text-white"
            >
              {grantExtra.isPending ? 'Granting...' : 'Grant'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
