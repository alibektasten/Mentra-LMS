'use client'

import { useState, useEffect, useRef } from 'react'
import {
  DndContext, DragEndEvent, PointerSensor, useSensor, useSensors, closestCenter,
} from '@dnd-kit/core'
import { arrayMove } from '@dnd-kit/sortable'
import { toast } from 'sonner'
import { Plus, Save } from 'lucide-react'
import { ModuleDTO } from '@/types/module'
import { LessonDTO } from '@/types/lesson'
import { useModules, useCreateModule, useUpdateModule } from '@/lib/queries/modules.queries'
import { useCreateLesson, useUpdateLesson } from '@/lib/queries/lessons.queries'
import * as modulesApi from '@/lib/api/modules'
import * as lessonsApi from '@/lib/api/lessons'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ModuleList } from './ModuleList'
import { useI18n } from '@/lib/i18n'

export interface BuilderLesson extends LessonDTO {
  _dirty?: boolean
}

export interface BuilderModule extends ModuleDTO {
  lessons: BuilderLesson[]
  _dirty?: boolean
  _open?: boolean
}

interface CourseBuilderProps {
  courseId: string
}

export function CourseBuilder({ courseId }: CourseBuilderProps) {
  const { t } = useI18n()
  const { data: modulesData, isLoading, isError, refetch } = useModules(courseId)
  const createModule = useCreateModule(courseId)
  const [modules, setModules] = useState<BuilderModule[]>([])
  const [isDirty, setIsDirty] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const isDirtyRef = useRef(isDirty)
  const editRevisionRef = useRef(0)
  const activeSaveRef = useRef(0)
  const currentCourseIdRef = useRef(courseId)

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }))

  useEffect(() => {
    isDirtyRef.current = isDirty
  }, [isDirty])

  useEffect(() => {
    // CourseBuilder can stay mounted while its route parameter changes. Reset
    // local state before accepting the next query result so one course's draft
    // is never rendered or saved under another course id.
    currentCourseIdRef.current = courseId
    editRevisionRef.current += 1
    activeSaveRef.current += 1
    isDirtyRef.current = false
    setModules([])
    setIsDirty(false)
    setIsSaving(false)
    setSaveError(null)
  }, [courseId])

  useEffect(() => {
    // A background refetch must never replace the teacher's unsaved ordering
    // or title edits. Fresh server data is applied again after a successful save.
    if (modulesData && modulesData.every((module) => module.course_id === courseId) && !isDirtyRef.current) {
      const sorted = [...modulesData].sort((a, b) => a.order - b.order)
      setModules(sorted.map((m) => ({ ...m, lessons: [], _open: false })))
    }
  }, [modulesData])

  // Warn on navigate-away if unsaved
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (isDirty) { e.preventDefault(); e.returnValue = '' }
    }
    window.addEventListener('beforeunload', handler)
    return () => window.removeEventListener('beforeunload', handler)
  }, [isDirty])

  const handleAddModule = async () => {
    if (createModule.isPending || isSaving) return
    const requestedCourseId = courseId
    setSaveError(null)
    try {
      const newModule = await createModule.mutateAsync({
        title: 'New Module',
        order: modules.length,
      })
      if (currentCourseIdRef.current !== requestedCourseId) return
      setModules((prev) => [...prev, { ...newModule, lessons: [], _open: true }])
    } catch {
      if (currentCourseIdRef.current !== requestedCourseId) return
      toast.error(t('moduleCreateFailedRetry'))
      setSaveError(t('moduleCreateFailedRetry'))
    }
  }

  const handleModuleDragEnd = (event: DragEndEvent) => {
    if (isSaving) return
    const { active, over } = event
    if (!over || active.id === over.id) return
    editRevisionRef.current += 1
    isDirtyRef.current = true
    setIsDirty(true)
    setModules((prev) => {
      const oldIdx = prev.findIndex((m) => m.id === active.id)
      const newIdx = prev.findIndex((m) => m.id === over.id)
      const reordered = arrayMove(prev, oldIdx, newIdx).map((m, i) => ({
        ...m, order: i, _dirty: m.order !== i || m._dirty,
      }))
      return reordered
    })
  }

  const handleModuleUpdate = (moduleId: string, updates: Partial<BuilderModule>) => {
    editRevisionRef.current += 1
    isDirtyRef.current = true
    setModules((prev) => prev.map((m) => m.id === moduleId ? { ...m, ...updates, _dirty: true } : m))
    setIsDirty(true)
  }

  const handleLessonsChange = (moduleId: string, lessons: BuilderLesson[]) => {
    editRevisionRef.current += 1
    isDirtyRef.current = true
    setModules((prev) => prev.map((m) => m.id === moduleId ? { ...m, lessons } : m))
    setIsDirty(true)
  }

  const handleSave = async () => {
    if (isSaving) return
    const savedCourseId = courseId
    const savedRevision = editRevisionRef.current
    const saveToken = activeSaveRef.current + 1
    activeSaveRef.current = saveToken
    const modulesToSave = modules
    setSaveError(null)
    setIsSaving(true)
    try {
      for (const module of modulesToSave) {
        if (module._dirty) {
          await modulesApi.updateModule(courseId, module.id, { title: module.title, order: module.order })
        }
        for (const lesson of module.lessons) {
          if (lesson._dirty) {
            await lessonsApi.updateLesson(module.id, lesson.id, {
              title: lesson.title, type: lesson.type, content: lesson.content,
              video_url: lesson.video_url, order: lesson.order,
            })
          }
        }
      }
      if (currentCourseIdRef.current !== savedCourseId) return
      if (editRevisionRef.current === savedRevision) {
        setModules((prev) => prev.map((m) => ({ ...m, _dirty: false, lessons: m.lessons.map((l) => ({ ...l, _dirty: false })) })))
        isDirtyRef.current = false
        setIsDirty(false)
      }
      toast.success(t('changesSaved'))
    } catch {
      if (currentCourseIdRef.current === savedCourseId) {
        toast.error(t('failedToSaveChanges'))
        setSaveError(t('courseContentSaveFailedRetry'))
      }
    } finally {
      if (activeSaveRef.current === saveToken) setIsSaving(false)
    }
  }

  if (isLoading) {
    return <div className="space-y-3">{Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-14" />)}</div>
  }

  if (isError && !modulesData) {
    return (
      <div role="alert" className="flex flex-col items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
        <p>{t('courseContentLoadFailed')}</p>
        <Button type="button" variant="outline" size="sm" onClick={() => void refetch()}>{t('tryAgain')}</Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h3 className="text-sm font-semibold text-ink">{t('courseContent')}</h3>
        <div className="flex flex-wrap gap-2">
          {isDirty && (
            <Button size="sm" onClick={handleSave} disabled={isSaving}>
              <Save className="h-4 w-4 mr-2" />
              {isSaving ? t('saving') : t('saveChanges')}
            </Button>
          )}
          <Button size="sm" variant="outline" onClick={handleAddModule} disabled={createModule.isPending || isSaving}>
            <Plus className="h-4 w-4 mr-2" /> {createModule.isPending ? t('creating') : t('addModule')}
          </Button>
        </div>
      </div>
      {saveError && (
        <div role="alert" className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {saveError}
        </div>
      )}
      <div
        aria-busy={isSaving || createModule.isPending}
        inert={isSaving}
        className={isSaving ? 'pointer-events-none select-none opacity-60' : undefined}
      >
        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleModuleDragEnd}>
        <ModuleList
          courseId={courseId}
          modules={modules}
          onModuleUpdate={handleModuleUpdate}
          onLessonsChange={handleLessonsChange}
        />
        </DndContext>
      </div>
    </div>
  )
}
