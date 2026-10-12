'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { useListSubmissions, useGradeSubmission } from '@/lib/queries/assignments.queries'
import { SubmissionDTO } from '@/types/assignment'
import { UPLOAD_BASE_URL } from '@/lib/api/client'
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
  const { data: submissions, isLoading } = useListSubmissions(assignmentId)
  const gradeSubmission = useGradeSubmission(assignmentId)

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map((i) => <Skeleton key={i} className="h-28 rounded-lg" />)}
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

  const handleSave = async () => {
    const parsed = parseInt(score, 10)
    if (isNaN(parsed) || parsed < 0 || parsed > maxPoints) {
      toast.error(`Score must be between 0 and ${maxPoints}`)
      return
    }
    setSaving(true)
    try {
      await onGrade(parsed, feedback)
      toast.success('Grade saved')
    } catch {
      toast.error('Failed to save grade')
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
      <div className="flex items-center justify-between px-4 py-3">
        <div className="flex items-center gap-3">
          <div className="h-7 w-7 rounded-full bg-[#f0eeeb] flex items-center justify-center text-xs font-semibold text-[#6b6b6b]">
            {initials || 'S'}
          </div>
          <div>
            <p className="text-sm font-semibold text-[#1a1a1a]">{studentName}</p>
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

        <div className="flex items-center gap-2">
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
          >
            {expanded ? 'Collapse' : 'Review'}
          </Button>
        </div>
      </div>

      {/* Expanded content */}
      {expanded && (
        <div className="border-t border-[#f0eeeb] px-4 py-4 space-y-4">
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
                  href={`${UPLOAD_BASE_URL}/${submission.file_path}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  download
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
            <div className="flex items-end gap-3">
              <div className="space-y-1">
                <label className="text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide">
                  Score (max {maxPoints})
                </label>
                <Input
                  type="number"
                  min={0}
                  max={maxPoints}
                  value={score}
                  onChange={(e) => setScore(e.target.value)}
                  className="w-24 h-8 text-sm border-[#e4e2de] focus-visible:ring-[#059669] focus-visible:ring-1"
                  placeholder="0"
                />
              </div>
              <span className="text-sm text-[#9b9b9b] mb-1.5">/ {maxPoints}</span>
            </div>
            <div className="space-y-1">
              <label className="text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide">
                Feedback (optional)
              </label>
              <Textarea
                rows={3}
                value={feedback}
                onChange={(e) => setFeedback(e.target.value)}
                placeholder="Add feedback for the student..."
                className="resize-none text-sm border-[#e4e2de] focus-visible:ring-[#059669] focus-visible:ring-1"
              />
            </div>
            <Button
              size="sm"
              disabled={saving || score.trim() === '' || !hasChanged}
              onClick={handleSave}
              className="bg-[#059669] hover:bg-[#047857] text-white"
            >
              {saving ? 'Saving…' : isGraded ? 'Update Grade' : 'Save Grade'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
