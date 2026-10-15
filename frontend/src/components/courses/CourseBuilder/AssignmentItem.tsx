'use client'

import { useRef, useState } from 'react'
import Link from 'next/link'
import { useParams, usePathname } from 'next/navigation'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical, Trash2, ClipboardList, Pencil, Upload, X, Clock, ClipboardCheck } from 'lucide-react'
import { toast } from 'sonner'
import { AssignmentDTO } from '@/types/assignment'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { useUpdateAssignment } from '@/lib/queries/assignments.queries'
import { useAttachments, useCreateAttachment, useDeleteAttachment } from '@/lib/queries/attachments.queries'
import { uploadFile } from '@/lib/api/upload'

interface AssignmentItemProps {
  moduleId: string
  courseId: string
  assignment: AssignmentDTO
  onDelete: (id: string) => void
  onUpdate: (updated: AssignmentDTO) => void
}

export function AssignmentItem({ assignment, moduleId, courseId, onDelete, onUpdate }: AssignmentItemProps) {
  const { org } = useParams<{ org: string }>()
  const pathname = usePathname()
  const isTeacherCourse = pathname.includes(`/${org}/teacher/courses/`)
  const [isDeleting, setIsDeleting] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editTitle, setEditTitle] = useState(assignment.title)
  const [editDescription, setEditDescription] = useState(assignment.description ?? '')
  const [editRubric, setEditRubric] = useState(assignment.rubric ?? '')
  const [editMaxPoints, setEditMaxPoints] = useState(assignment.max_points)
  const [editDueDate, setEditDueDate] = useState(
    assignment.due_date ? new Date(assignment.due_date).toISOString().slice(0, 16) : ''
  )
  const [editAllowLate, setEditAllowLate] = useState(assignment.allow_late_submission)
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const fieldID = `assignment-${assignment.id}`

  const updateAssignment = useUpdateAssignment(courseId, moduleId)
  const { data: attachments = [] } = useAttachments('assignment', assignment.id)
  const createAttachment = useCreateAttachment()
  const deleteAttachment = useDeleteAttachment('assignment', assignment.id)

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: assignment.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      onDelete(assignment.id)
      toast.success('Assignment deleted')
    } catch {
      toast.error('Failed to delete assignment')
      setIsDeleting(false)
    }
  }

  const handleEdit = () => {
    setEditTitle(assignment.title)
    setEditDescription(assignment.description ?? '')
    setEditRubric(assignment.rubric ?? '')
    setEditMaxPoints(assignment.max_points)
    setEditDueDate(
      assignment.due_date ? new Date(assignment.due_date).toISOString().slice(0, 16) : ''
    )
    setEditAllowLate(assignment.allow_late_submission)
    setFormError(null)
    setEditOpen(true)
  }

  const handleSave = async () => {
    if (saving) return
    setFormError(null)
    setSaving(true)
    try {
      const updated = await updateAssignment.mutateAsync({
        id: assignment.id,
        input: {
          title: editTitle,
          description: editDescription,
          rubric: editRubric,
          max_points: editMaxPoints,
          due_date: editDueDate ? new Date(editDueDate).toISOString() : undefined,
          allow_late_submission: editAllowLate,
        },
      })
      onUpdate(updated)
      setEditOpen(false)
      toast.success('Assignment updated')
    } catch {
      toast.error('Failed to update assignment')
      setFormError('Could not save this assignment. Your changes are still here; try again.')
    } finally {
      setSaving(false)
    }
  }

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const result = await uploadFile(file)
      await createAttachment.mutateAsync({
        ref_type: 'assignment',
        ref_id: assignment.id,
        stored_path: result.path,
        upload_token: result.upload_token,
        original_name: result.name,
        mime_type: file.type,
        size_bytes: file.size,
      })
      toast.success('File attached')
    } catch {
      toast.error('Failed to upload file')
      setFormError('Could not upload the file. You can choose it and try again.')
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  const handleDeleteAttachment = async (id: string) => {
    try {
      await deleteAttachment.mutateAsync(id)
      toast.success('Attachment removed')
    } catch {
      toast.error('Failed to remove attachment')
    }
  }

  const handleToggleLate = async () => {
    try {
      const updated = await updateAssignment.mutateAsync({
        id: assignment.id,
        input: { allow_late_submission: !assignment.allow_late_submission },
      })
      onUpdate(updated)
      toast.success(updated.allow_late_submission ? 'Late submissions allowed' : 'Late submissions disabled')
    } catch {
      toast.error('Failed to update assignment')
    }
  }

  const dueLabel = assignment.due_date
    ? new Date(assignment.due_date).toLocaleDateString()
    : null

  return (
    <>
      <div ref={setNodeRef} style={style} className="flex flex-wrap items-center gap-2 rounded-md border border-dashed bg-amber-50/50 px-3 py-2">
        <button {...attributes} {...listeners} aria-label={`Reorder ${assignment.title}`} className="cursor-grab text-ink-subtle hover:text-ink-muted">
          <GripVertical className="h-4 w-4" />
        </button>
        <ClipboardList className="h-3.5 w-3.5 shrink-0 text-amber-600" />
        <span className="min-w-0 flex-1 text-sm truncate">{assignment.title}</span>
        <span className="text-xs text-ink-muted">{assignment.max_points} pts</span>
        {dueLabel && (
          <span className="text-xs text-ink-subtle">Due {dueLabel}</span>
        )}
        <button
          onClick={handleToggleLate}
          title={assignment.allow_late_submission ? 'Late submissions allowed — click to disable' : 'Late submissions disabled — click to allow'}
          className={assignment.allow_late_submission
            ? 'flex items-center gap-1 text-[10px] font-medium px-1.5 py-0.5 rounded text-emerald-700 bg-emerald-50 hover:bg-emerald-100 transition-colors'
            : 'flex items-center gap-1 text-[10px] font-medium px-1.5 py-0.5 rounded text-ink-subtle bg-transparent hover:bg-surface-100 transition-colors'
          }
          disabled={updateAssignment.isPending}
        >
          <Clock className="h-3 w-3" />
          {assignment.allow_late_submission ? 'Late OK' : 'No late'}
        </button>
        <Button
          size="sm"
          variant="ghost"
          className="h-7 px-2"
          onClick={handleEdit}
          aria-label={`Edit ${assignment.title}`}
        >
          <Pencil className="h-3.5 w-3.5" />
        </Button>
        {isTeacherCourse && (
          <Button asChild size="sm" variant="ghost" className="h-7 px-2" title="Review submissions">
            <Link
              href={`/${org}/teacher/courses/${courseId}/assignments/${assignment.id}`}
              aria-label={`Review submissions for ${assignment.title}`}
            >
              <ClipboardCheck className="h-3.5 w-3.5" />
            </Link>
          </Button>
        )}
        <ConfirmDialog
          trigger={
            <Button
              size="sm"
              variant="ghost"
              className="h-7 px-2 text-destructive hover:text-destructive"
              disabled={isDeleting}
              aria-label={`Delete ${assignment.title}`}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          }
          title="Delete assignment?"
          description="All submissions for this assignment will also be deleted."
          confirmLabel="Delete"
          onConfirm={handleDelete}
          destructive
        />
      </div>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Assignment</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            {formError && <div role="alert" className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{formError}</div>}
            <div className="space-y-1.5">
              <Label htmlFor={`${fieldID}-title`}>Title</Label>
              <Input id={`${fieldID}-title`} value={editTitle} onChange={(e) => setEditTitle(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor={`${fieldID}-description`}>Statement / Description</Label>
              <Textarea
                id={`${fieldID}-description`}
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
                rows={5}
                placeholder="Assignment instructions…"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor={`${fieldID}-rubric`}>Rubric / assessment criteria</Label>
              <Textarea
                id={`${fieldID}-rubric`}
                value={editRubric}
                onChange={(e) => setEditRubric(e.target.value)}
                rows={4}
                placeholder="For example: research quality — 40 points; argumentation — 35 points; clarity and references — 25 points."
              />
              <p className="text-xs text-ink-muted">Students see these criteria before submitting; graders can use them while giving feedback.</p>
            </div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="space-y-1.5">
                <Label htmlFor={`${fieldID}-max-points`}>Max Points</Label>
                <Input
                  id={`${fieldID}-max-points`}
                  type="number"
                  min={0}
                  value={editMaxPoints}
                  onChange={(e) => setEditMaxPoints(Number(e.target.value))}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor={`${fieldID}-due-date`}>Due Date</Label>
                <Input
                  id={`${fieldID}-due-date`}
                  type="datetime-local"
                  value={editDueDate}
                  onChange={(e) => setEditDueDate(e.target.value)}
                />
              </div>
            </div>
            <div className="flex items-center gap-2">
              <input
                id={`${fieldID}-allow-late`}
                type="checkbox"
                checked={editAllowLate}
                onChange={(e) => setEditAllowLate(e.target.checked)}
                className="h-4 w-4 accent-accent-600"
              />
              <Label htmlFor={`${fieldID}-allow-late`} className="font-normal cursor-pointer">
                Allow late submission
              </Label>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor={`${fieldID}-file`}>Attachments</Label>
                <Button
                  size="sm"
                  variant="outline"
                  className="gap-1.5 h-7 text-xs"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={uploading}
                  type="button"
                >
                  <Upload className="h-3 w-3" />
                  {uploading ? 'Uploading…' : 'Upload File'}
                </Button>
                <input
                  ref={fileInputRef}
                  id={`${fieldID}-file`}
                  type="file"
                  className="hidden"
                  onChange={handleFileUpload}
                />
              </div>
              {attachments.length === 0 ? (
                <p className="text-xs text-ink-muted">No attachments yet.</p>
              ) : (
                <ul className="space-y-1">
                  {attachments.map((a) => (
                    <li key={a.id} className="flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm">
                      <span className="flex-1 truncate">{a.original_name}</span>
                      <button
                        onClick={() => handleDeleteAttachment(a.id)}
                        className="text-ink-muted hover:text-destructive"
                        title="Remove"
                        aria-label={`Remove ${a.original_name}`}
                      >
                        <X className="h-3.5 w-3.5" />
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Cancel</Button>
            <Button onClick={handleSave} disabled={saving || !editTitle.trim()} aria-busy={saving}>
              {saving ? 'Saving…' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
