'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { ChevronLeft, CheckCircle, BookOpen, Video, HelpCircle, FileIcon, Link2, Layers, Paperclip, Download } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useModules } from '@/lib/queries/modules.queries'
import { useMarkComplete, useProgress } from '@/lib/queries/progress.queries'
import * as lessonsApi from '@/lib/api/lessons'
import { LessonViewer } from '@/components/courses/LessonViewer'
import { FlashcardDeck } from '@/components/ai/FlashcardDeck'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { LessonDTO, LessonType } from '@/types/lesson'
import { cn } from '@/lib/utils/cn'
import { useGenerateFlashcards } from '@/lib/queries/ai.queries'
import { FlashcardDTO } from '@/lib/api/ai'
import { useAttachments } from '@/lib/queries/attachments.queries'
import { openPrivateFile } from '@/lib/api/private-files'
import { getApiErrorMessage } from '@/lib/api/error-message'

const lessonTypeBadge: Record<LessonType, string> = {
  video: 'text-sky-600 bg-sky-50',
  text: 'text-violet-600 bg-violet-50',
  quiz: 'text-amber-600 bg-amber-50',
  pdf: 'text-rose-600 bg-rose-50',
  link: 'text-teal-600 bg-teal-50',
}

const LessonTypeIcon: Record<LessonType, React.ElementType> = {
  video: Video,
  text: BookOpen,
  quiz: HelpCircle,
  pdf: FileIcon,
  link: Link2,
}

function useLessonFromCourse(courseId: string, lessonId: string) {
  const { data: modulesData, isLoading: modulesLoading } = useModules(courseId)
  const modules = Array.isArray(modulesData) ? modulesData : []

  return useQuery({
    queryKey: ['lesson-resolved', courseId, lessonId],
    queryFn: async (): Promise<LessonDTO> => {
      const results = await Promise.allSettled(
        modules.map((m) => lessonsApi.getLesson(m.id, lessonId))
      )
      const found = results.find((r) => r.status === 'fulfilled')
      if (found && found.status === 'fulfilled') return found.value
      throw new Error('Lesson not found')
    },
    enabled: !modulesLoading && modules.length > 0 && !!lessonId,
  })
}

export default function LessonPage() {
  const { org, courseId, lessonId } = useParams<{ org: string; courseId: string; lessonId: string }>()
  const router = useRouter()

  const { data: lesson, isLoading } = useLessonFromCourse(courseId, lessonId)
  const { data: attachments = [] } = useAttachments('lesson', lessonId)
  const markComplete = useMarkComplete()
  const { data: progressList } = useProgress(courseId)
  const isCompleted = Array.isArray(progressList)
    ? progressList.some((p) => p.lesson_id === lessonId && !!p.completed_at)
    : false

  const generateFlashcards = useGenerateFlashcards()
  const [flashcards, setFlashcards] = useState<FlashcardDTO[] | null>(null)

  const handleFlashcards = async () => {
    try {
      const cards = await generateFlashcards.mutateAsync({ lessonId, numCards: 8 })
      setFlashcards(cards)
    } catch (error) {
      toast.error(getApiErrorMessage(error, 'Не удалось создать карточки. Попробуйте ещё раз.'))
    }
  }

  const handleMarkComplete = async (score?: number) => {
    try {
      await markComplete.mutateAsync({ lessonId, score })
      toast.success('Lesson marked as complete!')
    } catch {
      toast.error('Could not save progress.')
    }
  }

  const handleQuizComplete = async (score: number) => {
    await handleMarkComplete(score)
  }

  if (isLoading) {
    return (
      <div className="max-w-3xl space-y-4">
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-64 w-full rounded-lg" />
      </div>
    )
  }

  if (!lesson) {
    return (
      <div className="max-w-3xl flex flex-col items-center justify-center py-24 text-center">
        <BookOpen className="h-8 w-8 text-[#c9c9c9] mb-3" />
        <p className="text-sm font-semibold text-[#1a1a1a] mb-1">Lesson not found</p>
        <p className="text-xs text-[#9b9b9b] mb-4">This lesson may have been removed or you don't have access.</p>
        <Link
          href={`/${org}/student/courses/${courseId}`}
          className="inline-flex items-center gap-1.5 text-xs text-[#059669] hover:underline"
        >
          <ChevronLeft className="h-3.5 w-3.5" /> Back to course
        </Link>
      </div>
    )
  }

  const Icon = LessonTypeIcon[lesson.type] ?? BookOpen
  const badge = lessonTypeBadge[lesson.type] ?? ''

  return (
    <div className="max-w-3xl">
      {/* Breadcrumb */}
      <Link
        href={`/${org}/student/courses/${courseId}`}
        className="inline-flex items-center gap-1 text-xs text-[#9b9b9b] hover:text-[#1a1a1a] transition-colors mb-6"
      >
        <ChevronLeft className="h-3.5 w-3.5" />
        Back to course
      </Link>

      {/* Lesson header */}
      <div className="mb-6">
        <div className="flex items-center gap-2 mb-2">
          <Icon className="h-4 w-4 text-[#9b9b9b]" />
          <span className={cn('text-[11px] font-medium px-2 py-0.5 rounded capitalize', badge)}>
            {lesson.type}
          </span>
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-[#1a1a1a]">{lesson.title}</h1>
      </div>

      {/* Content */}
      <div className="border border-[#e8e8e6] rounded-lg overflow-hidden">
        <div className="p-6">
          <LessonViewer lesson={lesson} onQuizComplete={handleQuizComplete} />
        </div>
      </div>

      {attachments.length > 0 && (
        <section className="mt-6 rounded-lg border border-[#e8e8e6] overflow-hidden">
          <div className="flex items-center gap-2 border-b border-[#e8e8e6] bg-[#fbfbfa] px-5 py-3">
            <Paperclip className="h-3.5 w-3.5 text-[#6b6b6b]" />
            <p className="text-xs font-semibold uppercase tracking-widest text-[#6b6b6b]">Lesson resources</p>
          </div>
          <div className="divide-y divide-[#e8e8e6]">
            {attachments.map((attachment) => (
              <a
                key={attachment.id}
                href="#"
                onClick={(event) => { event.preventDefault(); void openPrivateFile('attachments', attachment.id, attachment.original_name) }}
                className="flex items-center gap-3 px-5 py-3 text-sm text-[#059669] hover:bg-[#f7f7f5] transition-colors"
              >
                <Download className="h-3.5 w-3.5 shrink-0" />
                <span className="truncate">{attachment.original_name}</span>
              </a>
            ))}
          </div>
        </section>
      )}

      {/* Actions */}
      {lesson.type !== 'quiz' && (
        <div className="flex gap-2 mt-6 flex-wrap">
          <Button
            onClick={() => handleMarkComplete()}
            disabled={markComplete.isPending || isCompleted}
            size="sm"
            className={cn(
              'border-0 h-8 px-4 text-xs',
              isCompleted
                ? 'bg-[#059669] hover:bg-[#059669] text-white cursor-default'
                : 'bg-[#1a1a1a] hover:bg-[#2a2a2a] text-white'
            )}
          >
            <CheckCircle className="h-3.5 w-3.5 mr-1.5" />
            {markComplete.isPending ? 'Saving…' : isCompleted ? 'Completed' : 'Mark complete'}
          </Button>
          {lesson.type === 'text' && (
            <Button
              variant="ghost"
              size="sm"
              className="h-8 px-4 text-xs text-[#6b6b6b] hover:text-[#1a1a1a] hover:bg-[#f0efed]"
              onClick={handleFlashcards}
              disabled={generateFlashcards.isPending}
            >
              <Layers className="h-3.5 w-3.5 mr-1.5" />
              {generateFlashcards.isPending ? 'Generating…' : flashcards ? 'Regenerate flashcards' : 'Study with flashcards'}
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-4 text-xs text-[#6b6b6b] hover:text-[#1a1a1a] hover:bg-[#f0efed]"
            onClick={() => router.push(`/${org}/student/courses/${courseId}`)}
          >
            Back to course
          </Button>
        </div>
      )}

      {/* Flashcard deck */}
      {flashcards && flashcards.length > 0 && (
        <div className="mt-6 border border-[#e8e8e6] rounded-lg p-5">
          <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest mb-4">Flashcards</p>
          <FlashcardDeck cards={flashcards} />
        </div>
      )}
    </div>
  )
}
