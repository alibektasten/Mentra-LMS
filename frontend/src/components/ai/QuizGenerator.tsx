'use client'

import { useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { ClipboardCheck, Loader2, WandSparkles } from 'lucide-react'
import { toast } from 'sonner'
import { QuizDTO } from '@/types/quiz'
import { useGenerateQuiz } from '@/lib/queries/ai.queries'
import { useCourses } from '@/lib/queries/courses.queries'
import { useModules } from '@/lib/queries/modules.queries'
import { useLessons } from '@/lib/queries/lessons.queries'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { GeneratedQuizDisplay } from './GeneratedQuizDisplay'
import { getApiErrorMessage } from '@/lib/api/error-message'
import { useI18n } from '@/lib/i18n'

export function QuizGenerator() {
  const { t } = useI18n()
  const [selectedCourse, setSelectedCourse] = useState('')
  const [selectedModule, setSelectedModule] = useState('')
  const [selectedLesson, setSelectedLesson] = useState('')
  const [numQuestions, setNumQuestions] = useState(3)
  const [result, setResult] = useState<QuizDTO | null>(null)
  const [quizVersion, setQuizVersion] = useState(0)

  const { data: courses } = useCourses({ page: 1, page_size: 100 })
  const { data: modules = [] } = useModules(selectedCourse)
  const { data: lessons = [] } = useLessons(selectedModule)
  const generateQuiz = useGenerateQuiz()

  const handleGenerate = async () => {
    if (!selectedLesson) {
      toast.error(t('selectLessonFirst'))
      return
    }
    try {
      const quiz = await generateQuiz.mutateAsync({ lesson_id: selectedLesson, num_questions: numQuestions })
      setResult(quiz)
      setQuizVersion((version) => version + 1)
      toast.success(t('quizGenerated'))
    } catch (error) {
      toast.error(getApiErrorMessage(error, 'Не удалось создать квиз. Попробуйте ещё раз.'))
    }
  }

  return (
    <div className="space-y-6">
      <div className="ai-generator-surface ai-generator-surface--violet rounded-2xl border border-[#e5e2f1] bg-[linear-gradient(135deg,_rgba(245,243,255,0.9),_rgba(255,255,255,0.94)_52%,_rgba(250,245,255,0.85))] p-4 sm:p-5">
        <div className="mb-5 flex items-start gap-3">
          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-violet-600 text-white shadow-[0_8px_18px_rgba(109,40,217,0.18)]">
            <ClipboardCheck className="h-4 w-4" />
          </span>
          <div>
            <p className="text-sm font-semibold tracking-[-0.02em] text-[#282139]">{t('createLessonCheckIn')}</p>
            <p className="mt-0.5 text-xs leading-relaxed text-[#7d748e]">Generate a formative quiz grounded in your lesson, ready to review or build on.</p>
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div className="space-y-1.5">
            <Label className="text-[#665d77]">{t('course')}</Label>
          <Select value={selectedCourse} onValueChange={(v) => { setSelectedCourse(v); setSelectedModule(''); setSelectedLesson('') }}>
              <SelectTrigger className="h-10 rounded-xl border-[#e3dff0] bg-white/90 shadow-none"><SelectValue placeholder={t('selectCourse')} /></SelectTrigger>
            <SelectContent>
              {(courses?.data ?? []).map((c) => (
                <SelectItem key={c.id} value={c.id}>{c.title}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-[#665d77]">{t('module')}</Label>
          <Select value={selectedModule} onValueChange={(v) => { setSelectedModule(v); setSelectedLesson('') }} disabled={!selectedCourse}>
              <SelectTrigger className="h-10 rounded-xl border-[#e3dff0] bg-white/90 shadow-none"><SelectValue placeholder={t('selectModule')} /></SelectTrigger>
            <SelectContent>
              {modules.map((m) => (
                <SelectItem key={m.id} value={m.id}>{m.title}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-[#665d77]">{t('lesson')}</Label>
          <Select value={selectedLesson} onValueChange={setSelectedLesson} disabled={!selectedModule}>
              <SelectTrigger className="h-10 rounded-xl border-[#e3dff0] bg-white/90 shadow-none"><SelectValue placeholder={t('selectLesson')} /></SelectTrigger>
            <SelectContent>
              {lessons.map((l) => (
                <SelectItem key={l.id} value={l.id}>{l.title}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-[#665d77]">{t('questions')}</Label>
          <Input
            type="number"
            min={1}
            max={10}
            value={numQuestions}
            onChange={(e) => setNumQuestions(Number(e.target.value))}
              className="h-10 rounded-xl border-[#e3dff0] bg-white/90 shadow-none"
            />
          </div>
        </div>

        <div className="mt-5 flex flex-col items-start justify-between gap-3 border-t border-[#e5e1f0] pt-4 sm:flex-row sm:items-center">
          <p className="text-xs text-[#7c738d]">AI will prepare {numQuestions || 0} questions, answers, and explanations.</p>
          <Button onClick={handleGenerate} disabled={!selectedLesson || generateQuiz.isPending} className="w-full rounded-xl bg-violet-600 shadow-[0_4px_12px_rgba(109,40,217,0.18)] hover:bg-violet-700 hover:shadow-[0_8px_20px_rgba(109,40,217,0.22)] sm:w-auto">
            {generateQuiz.isPending ? (
              <><Loader2 className="h-4 w-4 animate-spin" /> {t('draftingQuiz')}</>
            ) : (
              <><WandSparkles className="h-4 w-4" /> {t('generateQuiz')}</>
            )}
          </Button>
        </div>
      </div>
      <AnimatePresence mode="wait">
        {result && (
          <motion.div
            key={quizVersion}
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
          >
            <GeneratedQuizDisplay quiz={result} />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
