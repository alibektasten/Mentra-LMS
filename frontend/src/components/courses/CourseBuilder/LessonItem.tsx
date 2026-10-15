'use client'

import { useRef, useState } from 'react'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical, Video, FileText, HelpCircle, FileIcon, Link2, Trash2, Pencil, Upload, X } from 'lucide-react'
import { toast } from 'sonner'
import { LessonType } from '@/types/lesson'
import { BuilderLesson } from './index'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { useDeleteLesson, useUpdateLesson } from '@/lib/queries/lessons.queries'
import { useAttachments, useCreateAttachment, useDeleteAttachment } from '@/lib/queries/attachments.queries'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { uploadFile } from '@/lib/api/upload'
import { QuizBuilder } from './QuizBuilder'

const LessonTypeIcon: Record<LessonType, React.ElementType> = {
  video: Video,
  text: FileText,
  quiz: HelpCircle,
  pdf: FileIcon,
  link: Link2,
}

interface LessonItemProps {
  moduleId: string
  lesson: BuilderLesson
  onUpdate: (updates: Partial<BuilderLesson>) => void
  onDelete: () => void
}

export function LessonItem({ moduleId, lesson, onUpdate, onDelete }: LessonItemProps) {
  const deleteLesson = useDeleteLesson(moduleId)
  const updateLesson = useUpdateLesson(moduleId)
  const Icon = LessonTypeIcon[lesson.type] ?? FileText

  const [editOpen, setEditOpen] = useState(false)
  const [editTitle, setEditTitle] = useState(lesson.title)
  const [editType, setEditType] = useState<LessonType>(lesson.type)
  const [editContent, setEditContent] = useState(lesson.content)
  const [editVideoUrl, setEditVideoUrl] = useState(lesson.video_url ?? '')
  const [editLinkUrl, setEditLinkUrl] = useState(lesson.link_url ?? '')
  const [editFileUrl, setEditFileUrl] = useState(lesson.file_url ?? '')
  const [editFileUploadToken, setEditFileUploadToken] = useState('')
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const pdfInputRef = useRef<HTMLInputElement>(null)
  const attachInputRef = useRef<HTMLInputElement>(null)
  const fieldID = `lesson-${lesson.id}`

  const { data: attachments = [] } = useAttachments('lesson', lesson.id)
  const createAttachment = useCreateAttachment()
  const deleteAttachment = useDeleteAttachment('lesson', lesson.id)

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: lesson.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const handleDelete = async () => {
    try {
      await deleteLesson.mutateAsync(lesson.id)
      onDelete()
    } catch {
      toast.error('Failed to delete lesson')
    }
  }

  const handleEdit = () => {
    setEditTitle(lesson.title)
    setEditType(lesson.type)
    setEditContent(lesson.content)
    setEditVideoUrl(lesson.video_url ?? '')
    setEditLinkUrl(lesson.link_url ?? '')
    setEditFileUrl(lesson.file_url ?? '')
    setEditFileUploadToken('')
    setFormError(null)
    setEditOpen(true)
  }

  const handleSave = async () => {
    if (saving) return
    setFormError(null)
    setSaving(true)
    try {
      const updated = await updateLesson.mutateAsync({
        id: lesson.id,
        input: {
          title: editTitle,
          type: editType,
          content: editContent,
          video_url: editVideoUrl,
          link_url: editLinkUrl,
          file_url: editFileUrl,
          upload_token: editFileUploadToken,
        },
      })
      onUpdate({ title: editTitle, type: editType, content: editContent, video_url: editVideoUrl, link_url: editLinkUrl, file_url: updated.file_url, _dirty: false })
      setEditFileUploadToken('')
      setEditOpen(false)
      toast.success('Lesson updated')
    } catch {
      toast.error('Failed to update lesson')
      setFormError('Could not save this lesson. Your changes are still here; try again.')
    } finally {
      setSaving(false)
    }
  }

  const handlePdfUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const result = await uploadFile(file)
      setEditFileUrl(result.path)
      setEditFileUploadToken(result.upload_token)
      toast.success('PDF uploaded')
    } catch {
      toast.error('Failed to upload PDF')
      setFormError('Could not upload the PDF. You can choose it and try again.')
    } finally {
      setUploading(false)
      if (pdfInputRef.current) pdfInputRef.current.value = ''
    }
  }

  const handleAttachUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const result = await uploadFile(file)
      await createAttachment.mutateAsync({
        ref_type: 'lesson',
        ref_id: lesson.id,
        stored_path: result.path,
        upload_token: result.upload_token,
        original_name: result.name,
        mime_type: file.type,
        size_bytes: file.size,
      })
      toast.success('File attached')
    } catch {
      toast.error('Failed to upload file')
      setFormError('Could not attach the file. You can choose it and try again.')
    } finally {
      setUploading(false)
      if (attachInputRef.current) attachInputRef.current.value = ''
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

  return (
    <>
      <div ref={setNodeRef} style={style} className="group flex items-center gap-2 rounded-md py-1.5 px-2 hover:bg-muted/50">
        <button {...attributes} {...listeners} aria-label={`Reorder ${lesson.title}`} className="cursor-grab text-ink-subtle hover:text-ink-muted">
          <GripVertical className="h-3.5 w-3.5" />
        </button>
        <Icon className="h-3.5 w-3.5 text-ink-muted shrink-0" />
        <span className="flex-1 text-xs text-ink">
          {lesson.title}
          {lesson._dirty && <span className="ml-1 text-accent">•</span>}
        </span>
        <span className="text-xs text-ink-subtle capitalize">{lesson.type}</span>
        <div className="flex items-center gap-1 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100">
          <Button size="sm" variant="ghost" className="h-6 w-6 p-0" onClick={handleEdit} aria-label={`Edit ${lesson.title}`}>
            <Pencil className="h-3 w-3" />
          </Button>
          <ConfirmDialog
            trigger={
              <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-destructive hover:text-destructive" aria-label={`Delete ${lesson.title}`}>
                <Trash2 className="h-3 w-3" />
              </Button>
            }
            title="Delete lesson?"
            description={`"${lesson.title}" will be permanently deleted.`}
            confirmLabel="Delete"
            onConfirm={handleDelete}
            destructive
          />
        </div>
      </div>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Lesson</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            {formError && <div role="alert" className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{formError}</div>}
            <div className="space-y-1.5">
              <Label htmlFor={`${fieldID}-title`}>Title</Label>
              <Input id={`${fieldID}-title`} value={editTitle} onChange={(e) => setEditTitle(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor={`${fieldID}-type`}>Type</Label>
              <Select value={editType} onValueChange={(v) => setEditType(v as LessonType)}>
                <SelectTrigger id={`${fieldID}-type`}><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="text">Text</SelectItem>
                  <SelectItem value="video">Video</SelectItem>
                  <SelectItem value="quiz">Quiz</SelectItem>
                  <SelectItem value="pdf">PDF</SelectItem>
                  <SelectItem value="link">Link</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {editType === 'video' && (
              <div className="space-y-1.5">
                <Label htmlFor={`${fieldID}-video-url`}>Video URL</Label>
                <Input id={`${fieldID}-video-url`} value={editVideoUrl} onChange={(e) => setEditVideoUrl(e.target.value)} placeholder="https://..." />
              </div>
            )}
            {editType === 'link' && (
              <div className="space-y-1.5">
                <Label htmlFor={`${fieldID}-link-url`}>Link URL</Label>
                <Input id={`${fieldID}-link-url`} value={editLinkUrl} onChange={(e) => setEditLinkUrl(e.target.value)} placeholder="https://..." />
              </div>
            )}
            {editType === 'pdf' && (
              <div className="space-y-1.5">
                <Label htmlFor={`${fieldID}-pdf-url`}>PDF File</Label>
                <div className="flex items-center gap-2">
                  <Input
                    id={`${fieldID}-pdf-url`}
                    value={editFileUrl}
                    onChange={(e) => setEditFileUrl(e.target.value)}
                    placeholder="/uploads/file.pdf"
                    className="flex-1"
                  />
                  <Button
                    size="sm"
                    variant="outline"
                    className="gap-1.5 shrink-0"
                    onClick={() => pdfInputRef.current?.click()}
                    disabled={uploading}
                    type="button"
                  >
                    <Upload className="h-3.5 w-3.5" />
                    {uploading ? 'Uploading…' : 'Upload'}
                  </Button>
                  <input
                    ref={pdfInputRef}
                    id={`${fieldID}-pdf-file`}
                    type="file"
                    accept=".pdf,application/pdf"
                    className="hidden"
                    onChange={handlePdfUpload}
                  />
                </div>
              </div>
            )}
            {editType === 'quiz' ? (
              <QuizBuilder lessonId={lesson.id} />
            ) : (editType === 'text' || editType === 'link') && (
              <div className="space-y-1.5">
                <Label htmlFor={`${fieldID}-content`}>Content</Label>
                <Textarea
                  id={`${fieldID}-content`}
                  value={editContent}
                  onChange={(e) => setEditContent(e.target.value)}
                  rows={editType === 'link' ? 3 : 10}
                  placeholder={editType === 'link' ? 'Optional description…' : 'Lesson content…'}
                />
              </div>
            )}

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor={`${fieldID}-attachment-file`}>Supplementary Files</Label>
                <Button
                  size="sm"
                  variant="outline"
                  className="gap-1.5 h-7 text-xs"
                  onClick={() => attachInputRef.current?.click()}
                  disabled={uploading}
                  type="button"
                >
                  <Upload className="h-3 w-3" />
                  {uploading ? 'Uploading…' : 'Attach File'}
                </Button>
                <input
                  ref={attachInputRef}
                  id={`${fieldID}-attachment-file`}
                  type="file"
                  className="hidden"
                  onChange={handleAttachUpload}
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
