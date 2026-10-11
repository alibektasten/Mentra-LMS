'use client'

import { useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'
import { useAssignment, useMySubmission, useSubmitAssignment, useDeleteMySubmission } from '@/lib/queries/assignments.queries'
import { useAttachments } from '@/lib/queries/attachments.queries'
import { UPLOAD_BASE_URL } from '@/lib/api/client'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils/cn'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@/components/ui/alert-dialog'
import { CheckCircle2, Clock, AlertTriangle, Pencil, Trash2, Download, Link2, FileText, Paperclip } from 'lucide-react'
import { AssignmentFeedbackPanel } from '@/components/ai/AssignmentFeedbackPanel'

interface AssignmentViewerProps {
  assignmentId: string
}

export function AssignmentViewer({ assignmentId }: AssignmentViewerProps) {
  const { data: assignment, isLoading: loadingAssignment } = useAssignment(assignmentId)
  const { data: submission, isLoading: loadingSubmission } = useMySubmission(assignmentId)
  const { data: attachments = [] } = useAttachments('assignment', assignmentId)
  const submitAssignment = useSubmitAssignment(assignmentId)
  const deleteMySubmission = useDeleteMySubmission(assignmentId)

  const [editing, setEditing] = useState(false)
  const [textContent, setTextContent] = useState('')
  const [linkUrl, setLinkUrl] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  // Pre-fill form with existing submission when entering edit mode
  useEffect(() => {
    if (editing && submission) {
      setTextContent(submission.text_content ?? '')
      setLinkUrl(submission.link_url ?? '')
      setFile(null)
    }
  }, [editing, submission])

  if (loadingAssignment) {
    return <div className="space-y-3"><Skeleton className="h-6 w-48" /><Skeleton className="h-24" /></div>
  }

  if (!assignment) return null

  const now = new Date()
  const dueDate = assignment.due_date ? new Date(assignment.due_date) : null
  const isPastDue = dueDate ? now > dueDate : false
  const isLocked = isPastDue && !assignment.allow_late_submission
  const daysUntilDue = dueDate ? Math.ceil((dueDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)) : null

  const isGraded = submission && submission.score !== null

  const handleSubmit = async () => {
    if (isLocked) return
    const formData = new FormData()
    formData.append('text_content', textContent)
    formData.append('link_url', linkUrl)
    if (file) formData.append('file', file)

    try {
      await submitAssignment.mutateAsync(formData)
      toast.success(submission ? 'Submission updated!' : 'Assignment submitted!')
      setEditing(false)
    } catch {
      toast.error('Failed to submit assignment')
    }
  }

  const handleDelete = async () => {
    try {
      await deleteMySubmission.mutateAsync()
      toast.success('Submission removed')
      setEditing(false)
      setTextContent('')
      setLinkUrl('')
      setFile(null)
    } catch {
      toast.error('Failed to remove submission')
    }
  }

  const showForm = !submission || editing

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-lg font-semibold">{assignment.title}</h2>
        <div className="flex items-center gap-3 mt-1">
          <span className="text-sm text-ink-muted">{assignment.max_points} points</span>
          {dueDate && (
            <span className={cn(
              'flex items-center gap-1 text-xs',
              isLocked ? 'text-red-600' : daysUntilDue !== null && daysUntilDue <= 2 ? 'text-orange-600' : 'text-ink-muted'
            )}>
              <Clock className="h-3.5 w-3.5" />
              Due {dueDate.toLocaleDateString()}
              {isPastDue && ' (past due)'}
            </span>
          )}
        </div>
        {assignment.description && (
          <div className="mt-3 rounded-md bg-[#f8f7f5] border border-[#e4e2de] px-4 py-3">
            <p className="text-sm leading-relaxed whitespace-pre-wrap text-[#1a1a1a]">{assignment.description}</p>
          </div>
        )}
        {assignment.rubric && (
          <div className="mt-3 rounded-md border border-sky-200 bg-sky-50 px-4 py-3 dark:border-sky-900 dark:bg-sky-950/30">
            <p className="text-xs font-semibold uppercase tracking-wide text-sky-700 dark:text-sky-300">Assessment criteria</p>
            <p className="mt-1 whitespace-pre-wrap text-sm leading-relaxed text-sky-950 dark:text-sky-100">{assignment.rubric}</p>
          </div>
        )}
        {attachments.length > 0 && (
          <div className="mt-3 space-y-1.5">
            <p className="flex items-center gap-1.5 text-xs font-semibold text-[#9b9b9b] uppercase tracking-wide">
              <Paperclip className="h-3 w-3" />
              Attachments
            </p>
            <div className="space-y-1">
              {attachments.map((a) => (
                <a
                  key={a.id}
                  href={`${UPLOAD_BASE_URL}/${a.stored_path}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  download
                  className="flex items-center gap-2 rounded-md border border-[#e4e2de] bg-white px-3 py-2 text-sm text-[#059669] hover:bg-[#f0eeeb] transition-colors"
                >
                  <Download className="h-3.5 w-3.5 shrink-0" />
                  {a.original_name}
                </a>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Deadline warning */}
      {daysUntilDue !== null && daysUntilDue <= 2 && !isPastDue && !submission && (
        <div className="flex items-center gap-2 rounded-lg border border-orange-200 bg-orange-50 px-4 py-3 text-sm text-orange-700">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          Due in {daysUntilDue === 0 ? 'less than a day' : `${daysUntilDue} day${daysUntilDue !== 1 ? 's' : ''}`}
        </div>
      )}

      {/* Locked state */}
      {isLocked && !submission && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          The deadline has passed and late submissions are not allowed.
        </div>
      )}

      {/* Existing submission — view mode */}
      {submission && !loadingSubmission && !editing && (
        <div className={cn(
          'rounded-lg border p-4 space-y-3',
          isGraded ? 'border-green-200 bg-green-50' : 'border-amber-200 bg-amber-50'
        )}>
          {/* Grade / status header */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CheckCircle2 className={cn('h-4 w-4 shrink-0', isGraded ? 'text-green-600' : 'text-amber-600')} />
              <span className="text-sm font-medium">
                {isGraded
                  ? `Graded: ${submission.score} / ${assignment.max_points} pts`
                  : 'Submitted — awaiting grade'}
              </span>
            </div>
            {!isGraded && !isLocked && (
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0 text-[#6b6b6b] hover:text-[#1a1a1a]"
                  onClick={() => setEditing(true)}
                >
                  <Pencil className="h-3.5 w-3.5" />
                </Button>
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 w-7 p-0 text-[#6b6b6b] hover:text-red-600"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Remove submission?</AlertDialogTitle>
                      <AlertDialogDescription>
                        This will permanently delete your submission. You can re-submit later if the deadline allows.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={handleDelete}
                        className="bg-red-600 hover:bg-red-700 text-white"
                      >
                        Remove
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            )}
          </div>

          {/* Feedback */}
          {submission.feedback && (
            <p className="text-sm text-ink-muted pl-6 border-l-2 border-green-200">{submission.feedback}</p>
          )}

          {/* Submission content */}
          <div className="space-y-2 pt-1">
            {submission.text_content && (
              <div className="rounded-md bg-white/60 p-3">
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide mb-1">
                  <FileText className="h-3 w-3" />
                  Your Response
                </div>
                <p className="text-sm whitespace-pre-wrap leading-relaxed">{submission.text_content}</p>
              </div>
            )}
            {submission.link_url && (
              <div className="rounded-md bg-white/60 p-3">
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide mb-1">
                  <Link2 className="h-3 w-3" />
                  Submitted Link
                </div>
                <a
                  href={submission.link_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-[#059669] hover:underline break-all"
                >
                  {submission.link_url}
                </a>
              </div>
            )}
            {submission.file_path && (
              <div className="rounded-md bg-white/60 p-3">
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-[#9b9b9b] uppercase tracking-wide mb-1">
                  <Download className="h-3 w-3" />
                  Attached File
                </div>
                <a
                  href={`${UPLOAD_BASE_URL}/${submission.file_path}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  download
                  className="text-sm text-[#059669] hover:underline flex items-center gap-1"
                >
                  {submission.file_path.split('/').pop()}
                </a>
              </div>
            )}
          </div>
        </div>
      )}

      {/* AI feedback — shown when submitted with text content, before grading */}
      {submission && !editing && (
        <AssignmentFeedbackPanel
          submissionId={submission.id}
          hasTextContent={!!submission.text_content}
        />
      )}

      {/* Submission form — new or edit */}
      {showForm && !isLocked && (
        <div className="space-y-4 rounded-lg border border-[#e4e2de] p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium">
              {editing ? 'Edit Submission' : 'Submit Your Work'}
            </h3>
            {editing && (
              <Button variant="ghost" size="sm" className="text-xs h-7" onClick={() => setEditing(false)}>
                Cancel
              </Button>
            )}
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="text_content">Written Response</Label>
            <Textarea
              id="text_content"
              rows={4}
              placeholder="Type your response here..."
              value={textContent}
              onChange={(e) => setTextContent(e.target.value)}
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="link_url">Link URL (optional)</Label>
            <Input
              id="link_url"
              type="url"
              placeholder="https://..."
              value={linkUrl}
              onChange={(e) => setLinkUrl(e.target.value)}
            />
          </div>
          <div className="space-y-1.5">
            <Label>File Upload (optional)</Label>
            <input
              ref={fileRef}
              type="file"
              className="hidden"
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            />
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={() => fileRef.current?.click()} type="button">
                Choose File
              </Button>
              {file ? (
                <span className="text-xs text-ink-muted">{file.name}</span>
              ) : editing && submission?.file_path ? (
                <span className="text-xs text-ink-muted">
                  Current: {submission.file_path.split('/').pop()} (leave empty to keep)
                </span>
              ) : null}
            </div>
          </div>
          <Button
            onClick={handleSubmit}
            disabled={submitAssignment.isPending || (!textContent && !linkUrl && !file)}
            className="bg-[#059669] hover:bg-[#047857] text-white"
          >
            {submitAssignment.isPending
              ? 'Saving…'
              : editing ? 'Save Changes' : 'Submit'}
          </Button>
        </div>
      )}
    </div>
  )
}
