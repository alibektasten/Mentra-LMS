import { LessonDTO } from '@/types/lesson'
import { VideoViewer } from './VideoViewer'
import { TextViewer } from './TextViewer'
import { QuizViewer } from './QuizViewer'
import { PDFViewer } from './PDFViewer'
import { LinkViewer } from './LinkViewer'

interface LessonViewerProps {
  lesson: LessonDTO
  onQuizComplete?: (score: number) => void
}

export function LessonViewer({ lesson, onQuizComplete }: LessonViewerProps) {
  switch (lesson.type) {
    case 'video':
      return <VideoViewer videoUrl={lesson.video_url ?? ''} title={lesson.title} />
    case 'quiz': {
      let questions = []
      try {
        questions = JSON.parse(lesson.content)
      } catch {
        questions = []
      }
      return <QuizViewer questions={questions} onComplete={onQuizComplete} />
    }
    case 'pdf':
      return <PDFViewer lessonId={lesson.id} fileUrl={lesson.file_url ?? ''} title={lesson.title} />
    case 'link':
      return <LinkViewer linkUrl={lesson.link_url ?? ''} title={lesson.title} content={lesson.content} />
    case 'text':
    default:
      return <TextViewer content={lesson.content} />
  }
}
