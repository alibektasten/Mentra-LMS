'use client'

import { useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { FileText, Loader2, WandSparkles } from 'lucide-react'
import { toast } from 'sonner'
import { useSummarizeLesson } from '@/lib/queries/ai.queries'
import { useCourses } from '@/lib/queries/courses.queries'
import { useModules } from '@/lib/queries/modules.queries'
import { useLessons } from '@/lib/queries/lessons.queries'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { getApiErrorMessage } from '@/lib/api/error-message'
import { useI18n } from '@/lib/i18n'

export function LessonSummarizer() {
  const { t } = useI18n()
  const [selectedCourse, setSelectedCourse] = useState('')
  const [selectedModule, setSelectedModule] = useState('')
  const [selectedLesson, setSelectedLesson] = useState('')
  const [summary, setSummary] = useState('')

  const { data: courses } = useCourses({ page: 1, page_size: 100 })
  const { data: modules = [] } = useModules(selectedCourse)
  const { data: lessons = [] } = useLessons(selectedModule)
  const summarize = useSummarizeLesson()

  const handleSummarize = async () => {
    if (!selectedLesson) { toast.error(t('selectLessonFirst')); return }
    try {
      const result = await summarize.mutateAsync(selectedLesson)
      setSummary(result.summary)
      toast.success(t('summaryGenerated'))
    } catch (error) {
      toast.error(getApiErrorMessage(error, 'Не удалось создать конспект. Попробуйте ещё раз.'))
    }
  }

  return (
    <div className="space-y-6">
      <div className="ai-generator-surface ai-generator-surface--sky rounded-2xl border border-[#dfe9e7] bg-[linear-gradient(135deg,_rgba(240,249,255,0.78),_rgba(255,255,255,0.94)_52%,_rgba(240,253,250,0.78))] p-4 sm:p-5">
        <div className="mb-5 flex items-start gap-3">
          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-sky-600 text-white shadow-[0_8px_18px_rgba(2,132,199,0.18)]">
            <FileText className="h-4 w-4" />
          </span>
          <div>
            <p className="text-sm font-semibold tracking-[-0.02em] text-[#173040]">{t('lessonBrief')}</p>
            <p className="mt-0.5 text-xs leading-relaxed text-[#6f7e87]">Get a concise, structured review of the ideas your learners should keep in mind.</p>
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div className="space-y-1.5">
            <Label className="text-[#526873]">{t('course')}</Label>
          <Select value={selectedCourse} onValueChange={(v) => { setSelectedCourse(v); setSelectedModule(''); setSelectedLesson('') }}>
              <SelectTrigger className="h-10 rounded-xl border-[#dce8eb] bg-white/90 shadow-none"><SelectValue placeholder={t('selectCourse')} /></SelectTrigger>
            <SelectContent>
              {(courses?.data ?? []).map((c) => (
                <SelectItem key={c.id} value={c.id}>{c.title}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-[#526873]">{t('module')}</Label>
          <Select value={selectedModule} onValueChange={(v) => { setSelectedModule(v); setSelectedLesson('') }} disabled={!selectedCourse}>
              <SelectTrigger className="h-10 rounded-xl border-[#dce8eb] bg-white/90 shadow-none"><SelectValue placeholder={t('selectModule')} /></SelectTrigger>
            <SelectContent>
              {modules.map((m) => <SelectItem key={m.id} value={m.id}>{m.title}</SelectItem>)}
            </SelectContent>
          </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-[#526873]">{t('lesson')}</Label>
          <Select value={selectedLesson} onValueChange={setSelectedLesson} disabled={!selectedModule}>
              <SelectTrigger className="h-10 rounded-xl border-[#dce8eb] bg-white/90 shadow-none"><SelectValue placeholder={t('selectLesson')} /></SelectTrigger>
            <SelectContent>
              {lessons.map((l) => <SelectItem key={l.id} value={l.id}>{l.title}</SelectItem>)}
            </SelectContent>
          </Select>
          </div>
        </div>
        <div className="mt-5 flex flex-col items-start justify-between gap-3 border-t border-[#dce8eb] pt-4 sm:flex-row sm:items-center">
          <p className="text-xs text-[#71818a]">Your summary will be grounded in the selected lesson’s content.</p>
          <Button onClick={handleSummarize} disabled={!selectedLesson || summarize.isPending} className="w-full rounded-xl bg-sky-600 shadow-[0_4px_12px_rgba(2,132,199,0.18)] hover:bg-sky-700 hover:shadow-[0_8px_20px_rgba(2,132,199,0.22)] sm:w-auto">
            {summarize.isPending ? (
              <><Loader2 className="h-4 w-4 animate-spin" /> {t('writingSummary')}</>
            ) : (
              <><WandSparkles className="h-4 w-4" /> {t('generateSummary')}</>
            )}
          </Button>
        </div>
      </div>
      <AnimatePresence mode="wait">
        {summary && (
          <motion.article
            key={summary}
            initial={{ opacity: 0, y: 14, filter: 'blur(4px)' }}
            animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.38, ease: [0.22, 1, 0.36, 1] }}
            className="overflow-hidden rounded-2xl border border-[#d9e8ee] bg-white shadow-[0_12px_28px_rgba(21,80,105,0.07)]"
          >
            <div className="flex items-center gap-2 border-b border-[#e7eff2] bg-[#f8fcfd] px-5 py-3">
              <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sky-100 text-sky-700"><FileText className="h-3.5 w-3.5" /></span>
              <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[#59717d]">{t('aiLessonBrief')}</p>
            </div>
            <div className="p-5 sm:p-6">
              <p className="whitespace-pre-wrap text-sm leading-7 text-[#263b45]">{summary}</p>
            </div>
          </motion.article>
        )}
      </AnimatePresence>
    </div>
  )
}
