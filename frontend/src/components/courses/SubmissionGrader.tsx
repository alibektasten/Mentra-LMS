'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { useListSubmissions, useGradeSubmission } from '@/lib/queries/assignments.queries'
import { SubmissionDTO } from '@/types/assignment'
import { openPrivateFile } from '@/lib/api/private-files'
import { validateIntegerScore } from '@/lib/validators/score'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils/cn'
import { CheckCircle2, Clock, Download, ExternalLink, FileText, Link2, Paperclip } from 'lucide-react'

interface SubmissionGraderProps {
  assignmentId: string
  maxPoints: number
  dueDate?: string | null
}

export function SubmissionGrader({ assignmentId, maxPoints, dueDate }: SubmissionGraderProps) {
  const { data: submissions, isLoading, isError, refetch } = useListSubmissions(assignmentId)
  const gradeSubmission = useGradeSubmission(assignmentId)

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map((i) => <Skeleton key={i} className="h-28 rounded-lg" />)}
      </div>
    )
  }

  if (isError) {
    return (
      <div role="alert" className="flex flex-col items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
        <p>Could not load submissions. Try again before grading.</p>
        <Button type="button" variant="outline" size="sm" onClick={() => void refetch()}>Try again</Button>
      </div>
    )
  }

  if (!submissions || submissions.length === 0) {
    return (
      <div className="py-12 text-center text-sm text-[#9b9b9b]">
        No submissions yet.
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {submissions.map((submission) => (
        <SubmissionRow
          key={submission.id}
          submission={submission}
          maxPoints={maxPoints}
          dueDate={dueDate}
          onGrade={(score, feedback) =>
            gradeSubmission.mutateAsync({ submissionID: submission.id, score, feedback })
          }
        />
      ))}
    </div>
  )
}

interface SubmissionRowProps {
  submission: SubmissionDTO
  maxPoints: number
  dueDate?: string | null
  onGrade: (score: number, feedback: string) => Promise<SubmissionDTO>
}

function SubmissionRow({ submission, maxPoints, dueDate, onGrade }: SubmissionRowProps) {
  const isGraded = submission.score !== null
  const isLate = !!(dueDate && new Date(submission.submitted_at) > new Date(dueDate))
  const [score, setScore] = useState(submission.score?.toString() ?? '')
  const [feedback, setFeedback] = useState(submission.feedback ?? '')
  const [saving, setSaving] = useState(false)
  const [expanded, setExpanded] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)
  const scoreID = `submission-score-${submission.id}`
  const feedbackID = `submission-feedback-${submission.id}`
  const detailID = `submission-detail-${submission.id}`

  const handleSave = async () => {
    if (saving) return
    const validation = validateIntegerScore(score, maxPoints)
    if (!validation.valid) {
      setValidationError(validation.message)
      return
    }

    setValidationError(null)
    setSaveError(null)
    setSaving(true)
    try {
      await onGrade(validation.score, feedback)
      toast.success('Grade saved')
    } catch {
      toast.error('Failed to save grade')
      setSaveError('Could not save this grade. The score and feedback are still here; try again.')
    } finally {
      setSaving(false)
    }
  }

  const submittedDate = new Date(submission.submitted_at).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
  })

  const hasChanged =
    score !== (submission.score?.toString() ?? '') ||
    feedback !== (submission.feedback ?? '')
  const studentName = submission.student_name || 'Student'
  const initials = studentName.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0]).join('').toUpperCase()

  return (
    <div className={cn(
      'rounded-lg border bg-white transition-shadow',
      isGraded ? 'border-[#e4e2de]' : 'border-amber-200',
    )}>
      {/* Row header */}
      <div className="flex flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 items-center gap-3">
          <div className="h-7 w-7 rounded-full bg-[#f0eeeb] flex items-center justify-center text-xs font-semibold text-[#6b6b6b]">
            {initials || 'S'}
          </div>
          <div>
            <p className="truncate text-sm font-semibold text-[#1a1a1a]">{studentName}</p>
            <div className="flex items-center gap-1.5 text-[11px] text-[#9b9b9b]">
              <Clock className="h-3 w-3" />
              {submittedDate}
              {isLate && (
                <span className="flex items-center gap-0.5 text-[10px] font-semibold text-red-600 bg-red-50 border border-red-200 rounded px-1 py-0.5 leading-none">
                  Late
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2 sm:justify-end">
          {isGraded ? (
            <span className="flex items-center gap-1 text-xs font-medium text-emerald-700 bg-emerald-50 border border-emerald-200 rounded-full px-2 py-0.5">
              <CheckCircle2 className="h-3 w-3" />
              {submission.score} / {maxPoints}
            </span>
          ) : (
            <span className="text-xs font-medium text-amber-700 bg-amber-50 border border-amber-200 rounded-full px-2 py-0.5">
              Needs grading
            </span>
          )}
          <Button
            variant="ghost"
            size="sm"
            className="text-xs h-7 px-2 text-[#6b6b6b]"
            onClick={() => setExpanded((v) => !v)}
            aria-expanded={expanded}
            aria-controls={detailID}
          >
            {expanded ? 'Collapse' : 'Review'}
          </Button>
        </div>
      </div>

      {/* Expanded content */}
      {expanded && (
        <div id={detailID} className="border-t border-[#f0eeeb] px-4 py-4 space-y-4">
          {/* Submission content */}
          <div className="space-y-2">
            {submission.text_content && (
              <div className="rounded-md bg-[#f8f7f5] p-3">
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide mb-1.5">
                  <FileText className="h-3 w-3" />
                  Written Response
                </div>
                <p className="text-sm text-[#1a1a1a] whitespace-pre-wrap leading-relaxed">
                  {submission.text_content}
                </p>
              </div>
            )}
            {submission.link_url && (
              <div className="rounded-md bg-[#f8f7f5] p-3">
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide mb-1.5">
                  <Link2 className="h-3 w-3" />
                  Link
                </div>
                <a
                  href={submission.link_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-[#059669] hover:underline flex items-center gap-1"
                >
                  {submission.link_url}
                  <ExternalLink className="h-3 w-3 shrink-0" />
                </a>
              </div>
            )}
            {submission.file_path && (
              <div className="rounded-md bg-[#f8f7f5] p-3">
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide mb-1.5">
                  <Paperclip className="h-3 w-3" />
                  Attached File
                </div>
                <a
                  href="#"
                  onClick={(event) => { event.preventDefault(); void openPrivateFile('submissions', submission.id, submission.file_path.split('/').pop() ?? 'submission').catch(() => toast.error('Could not download file')) }}
                  className="inline-flex items-center gap-1.5 text-sm text-[#059669] hover:underline"
                >
                  <Download className="h-3.5 w-3.5 shrink-0" />
                  {submission.file_path.split('/').pop()}
                </a>
              </div>
            )}
            {!submission.text_content && !submission.link_url && !submission.file_path && (
              <p className="text-sm text-[#9b9b9b] italic">No submission content.</p>
            )}
          </div>

          {/* Grading form */}
          <div className="space-y-3 pt-1 border-t border-[#f0eeeb]">
            <div className="flex flex-wrap items-end gap-3">
              <div className="space-y-1">
                <label htmlFor={scoreID} className="text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide">
                  Score (max {maxPoints})
                </label>
                <Input
                  type="number"
                  id={scoreID}
                  min={0}
                  max={maxPoints}
                  step={1}
                  value={score}
                  onChange={(e) => {
                    setScore(e.target.value)
                    setValidationError(null)
                  }}
                  className="w-24 h-8 text-sm border-[#e4e2de] focus-visible:ring-[#059669] focus-visible:ring-1"
                  placeholder="0"
                  aria-invalid={!!validationError}
                  aria-describedby={validationError ? `${scoreID}-error` : undefined}
                />
              </div>
              <span className="text-sm text-[#9b9b9b] mb-1.5">/ {maxPoints}</span>
            </div>
            {validationError && (
              <div id={`${scoreID}-error`} role="alert" className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
                {validationError}
              </div>
            )}
            <div className="space-y-1">
              <label htmlFor={feedbackID} className="text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide">
                Feedback (optional)
              </label>
              <Textarea
                rows={3}
                id={feedbackID}
                value={feedback}
                onChange={(e) => setFeedback(e.target.value)}
                placeholder="Add feedback for the student..."
                className="resize-none text-sm border-[#e4e2de] focus-visible:ring-[#059669] focus-visible:ring-1"
              />
            </div>
            {saveError && (
              <div role="alert" className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
                <p>{saveError}</p>
                <Button type="button" variant="link" className="mt-1 h-auto p-0 text-red-800" onClick={handleSave} disabled={saving}>Try saving again</Button>
              </div>
            )}
            <Button
              size="sm"
              disabled={saving || !hasChanged}
              onClick={handleSave}
              className="bg-[#059669] hover:bg-[#047857] text-white"
              aria-busy={saving}
            >
              {saving ? 'Saving…' : isGraded ? 'Update Grade' : 'Save Grade'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
