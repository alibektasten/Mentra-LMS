'use client'

import { useState, useRef } from 'react'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import {
  DndContext, PointerSensor, useSensor, useSensors, closestCenter, DragEndEvent,
} from '@dnd-kit/core'
import { arrayMove } from '@dnd-kit/sortable'
import { GripVertical, ChevronDown, ChevronRight, Plus, Trash2, ClipboardList } from 'lucide-react'
import { toast } from 'sonner'
import { BuilderModule, BuilderLesson } from './index'
import { LessonList } from './LessonList'
import { AssignmentItem } from './AssignmentItem'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useDeleteModule } from '@/lib/queries/modules.queries'
import { useCreateLesson } from '@/lib/queries/lessons.queries'
import { useCreateAssignment } from '@/lib/queries/assignments.queries'
import * as lessonsApi from '@/lib/api/lessons'
import * as assignmentsApi from '@/lib/api/assignments'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { AssignmentDTO } from '@/types/assignment'

interface ModuleItemProps {
  courseId: string
  module: BuilderModule
  onUpdate: (updates: Partial<BuilderModule>) => void
  onLessonsChange: (lessons: BuilderLesson[]) => void
}

export function ModuleItem({ courseId, module, onUpdate, onLessonsChange }: ModuleItemProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [title, setTitle] = useState(module.title)
  const [assignments, setAssignments] = useState<AssignmentDTO[]>([])
  const inputRef = useRef<HTMLInputElement>(null)
  const deleteModule = useDeleteModule(courseId)
  const createLesson = useCreateLesson(module.id)
  const createAssignment = useCreateAssignment(courseId, module.id)

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }))

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: module.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const handleTitleBlur = () => {
    setIsEditing(false)
    if (title !== module.title) {
      onUpdate({ title })
    }
  }

  const handleAddLesson = async () => {
    try {
      const lesson = await createLesson.mutateAsync({
        title: 'New Lesson',
        type: 'text',
        content: '',
        order: module.lessons.length,
      })
      onLessonsChange([...module.lessons, { ...lesson, _dirty: false }])
      onUpdate({ _open: true })
    } catch {
      toast.error('Failed to create lesson')
    }
  }

  const handleAddAssignment = async () => {
    try {
      const assignment = await createAssignment.mutateAsync({
        title: 'New Assignment',
        max_points: 100,
      })
      setAssignments((prev) => [...prev, assignment])
      onUpdate({ _open: true })
    } catch {
      toast.error('Failed to create assignment')
    }
  }

  const handleLessonDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIdx = module.lessons.findIndex((l) => l.id === active.id)
    const newIdx = module.lessons.findIndex((l) => l.id === over.id)
    const reordered = arrayMove(module.lessons, oldIdx, newIdx).map((l, i) => ({
      ...l, order: i, _dirty: true,
    }))
    onLessonsChange(reordered)
  }

  const handleDelete = async () => {
    try {
      await deleteModule.mutateAsync(module.id)
      toast.success('Module deleted')
    } catch {
      toast.error('Failed to delete module')
    }
  }

  // Fetch lessons and assignments when opening
  const handleToggleOpen = async () => {
    const opening = !module._open
    if (opening && module.lessons.length === 0) {
      try {
        const lessons = await lessonsApi.listLessons(module.id)
        onLessonsChange(lessons.map((l) => ({ ...l, _dirty: false })))
      } catch {
        // ignore
      }
    }
    if (opening && assignments.length === 0) {
      try {
        const fetched = await assignmentsApi.listAssignments(courseId, module.id)
        setAssignments(fetched)
      } catch {
        // ignore
      }
    }
    onUpdate({ _open: opening })
  }

  return (
    <div ref={setNodeRef} style={style} className="rounded-lg border bg-white shadow-sm">
      <div className="flex items-center gap-2 px-3 py-2.5">
        <button {...attributes} {...listeners} className="cursor-grab text-ink-subtle hover:text-ink-muted">
          <GripVertical className="h-4 w-4" />
        </button>
        <button
          type="button"
          onClick={handleToggleOpen}
          aria-label={`Toggle module ${module.title}`}
          aria-expanded={Boolean(module._open)}
          className="text-ink-muted hover:text-ink"
        >
          {module._open ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
        </button>
        {isEditing ? (
          <Input
            ref={inputRef}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onBlur={handleTitleBlur}
            onKeyDown={(e) => { if (e.key === 'Enter') inputRef.current?.blur() }}
            className="h-7 flex-1 text-sm"
            autoFocus
          />
        ) : (
          <span
            className="flex-1 text-sm font-medium cursor-pointer hover:text-accent"
            onClick={() => setIsEditing(true)}
          >
            {module.title}
            {module._dirty && <span className="ml-1 text-xs text-accent">•</span>}
          </span>
        )}
        <div className="flex items-center gap-1">
          <Button size="sm" variant="ghost" className="h-7 px-2 text-xs" onClick={handleAddLesson} title="Add lesson">
            <Plus className="h-3.5 w-3.5 mr-1" />Lesson
          </Button>
          <Button size="sm" variant="ghost" className="h-7 px-2 text-xs" onClick={handleAddAssignment} title="Add assignment">
            <ClipboardList className="h-3.5 w-3.5 mr-1" />Task
          </Button>
          <ConfirmDialog
            trigger={
              <Button size="sm" variant="ghost" className="h-7 px-2 text-destructive hover:text-destructive">
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            }
            title="Delete module?"
            description="All lessons in this module will also be deleted."
            confirmLabel="Delete"
            onConfirm={handleDelete}
            destructive
          />
        </div>
      </div>
      {module._open && (
        <div className="border-t px-3 py-2">
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleLessonDragEnd}>
            <LessonList
              moduleId={module.id}
              lessons={module.lessons}
              onLessonsChange={onLessonsChange}
            />
          </DndContext>
          {assignments.length > 0 && (
            <div className="mt-2 space-y-1">
              <p className="text-[10px] font-semibold text-ink-subtle uppercase tracking-wide px-1 pt-1">Assignments</p>
              {assignments.map((a) => (
                <AssignmentItem
                  key={a.id}
                  moduleId={module.id}
                  courseId={courseId}
                  assignment={a}
                  onDelete={(id) => setAssignments((prev) => prev.filter((x) => x.id !== id))}
                  onUpdate={(updated) => setAssignments((prev) => prev.map((x) => x.id === updated.id ? updated : x))}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
