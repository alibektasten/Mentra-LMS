'use client'

import { useState, useEffect, useCallback } from 'react'
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
  const { data: modulesData, isLoading } = useModules(courseId)
  const createModule = useCreateModule(courseId)
  const [modules, setModules] = useState<BuilderModule[]>([])
  const [isDirty, setIsDirty] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }))

  useEffect(() => {
    if (modulesData) {
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
    try {
      const newModule = await createModule.mutateAsync({
        title: 'New Module',
        order: modules.length,
      })
      setModules((prev) => [...prev, { ...newModule, lessons: [], _open: true }])
    } catch {
      toast.error('Failed to create module')
    }
  }

  const handleModuleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    setModules((prev) => {
      const oldIdx = prev.findIndex((m) => m.id === active.id)
      const newIdx = prev.findIndex((m) => m.id === over.id)
      const reordered = arrayMove(prev, oldIdx, newIdx).map((m, i) => ({
        ...m, order: i, _dirty: m.order !== i || m._dirty,
      }))
      setIsDirty(true)
      return reordered
    })
  }

  const handleModuleUpdate = (moduleId: string, updates: Partial<BuilderModule>) => {
    setModules((prev) => prev.map((m) => m.id === moduleId ? { ...m, ...updates, _dirty: true } : m))
    setIsDirty(true)
  }

  const handleLessonsChange = (moduleId: string, lessons: BuilderLesson[]) => {
    setModules((prev) => prev.map((m) => m.id === moduleId ? { ...m, lessons } : m))
    setIsDirty(true)
  }

  const handleSave = async () => {
    setIsSaving(true)
    try {
      for (const module of modules) {
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
      setModules((prev) => prev.map((m) => ({ ...m, _dirty: false, lessons: m.lessons.map((l) => ({ ...l, _dirty: false })) })))
      setIsDirty(false)
      toast.success(t('changesSaved'))
    } catch {
      toast.error(t('failedToSaveChanges'))
    } finally {
      setIsSaving(false)
    }
  }

  if (isLoading) {
    return <div className="space-y-3">{Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-14" />)}</div>
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-ink">{t('courseContent')}</h3>
        <div className="flex gap-2">
          {isDirty && (
            <Button size="sm" onClick={handleSave} disabled={isSaving}>
              <Save className="h-4 w-4 mr-2" />
              {isSaving ? t('saving') : t('saveChanges')}
            </Button>
          )}
          <Button size="sm" variant="outline" onClick={handleAddModule}>
            <Plus className="h-4 w-4 mr-2" /> {t('addModule')}
          </Button>
        </div>
      </div>
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleModuleDragEnd}>
        <ModuleList
          courseId={courseId}
          modules={modules}
          onModuleUpdate={handleModuleUpdate}
          onLessonsChange={handleLessonsChange}
        />
      </DndContext>
    </div>
  )
}
