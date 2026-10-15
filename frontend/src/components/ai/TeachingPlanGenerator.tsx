'use client'

import { useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { ClipboardCheck, Loader2, Sparkles } from 'lucide-react'
import { toast } from 'sonner'
import { generateTeachingPlan } from '@/lib/api/ai'
import { getApiErrorMessage } from '@/lib/api/error-message'
import { useI18n, type Locale } from '@/lib/i18n'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const copy: Record<Locale, Record<string, string>> = {
  en: {
    title: 'Build a ready-to-teach lesson',
    description: 'Give AI a topic, learner level, and duration. It will draft objectives, activities, and an exit ticket.',
    topic: 'Topic', topicPlaceholder: 'e.g. REST API design principles',
    audience: 'Learner level', audiencePlaceholder: 'e.g. Second-year computer science students',
    duration: 'Lesson duration', language: 'Output language',
    generate: 'Generate lesson plan', generating: 'Designing lesson plan…',
    result: 'AI lesson plan', ready: 'Lesson plan ready',
    required: 'Add a topic and learner level first.',
    failed: 'Could not generate the lesson plan. Please try again.',
    minute: 'minutes',
  },
  ru: {
    title: 'Создайте готовый план занятия',
    description: 'Укажите тему, уровень группы и длительность — AI подготовит цели, активности и проверку в конце занятия.',
    topic: 'Тема', topicPlaceholder: 'Например: принципы проектирования REST API',
    audience: 'Уровень группы', audiencePlaceholder: 'Например: студенты 2-го курса computer science',
    duration: 'Длительность занятия', language: 'Язык результата',
    generate: 'Создать план занятия', generating: 'Готовим план занятия…',
    result: 'AI-план занятия', ready: 'План занятия готов',
    required: 'Сначала укажите тему и уровень группы.',
    failed: 'Не удалось создать план занятия. Попробуйте ещё раз.',
    minute: 'минут',
  },
  kk: {
    title: 'Дайын сабақ жоспарын жасаңыз',
    description: 'Тақырыпты, топ деңгейін және ұзақтығын көрсетіңіз — AI мақсаттарды, әрекеттерді және қорытынды тексеруді дайындайды.',
    topic: 'Тақырып', topicPlaceholder: 'Мысалы: REST API жобалау қағидалары',
    audience: 'Топ деңгейі', audiencePlaceholder: 'Мысалы: computer science 2-курс студенттері',
    duration: 'Сабақ ұзақтығы', language: 'Нәтиже тілі',
    generate: 'Сабақ жоспарын жасау', generating: 'Сабақ жоспары дайындалуда…',
    result: 'AI сабақ жоспары', ready: 'Сабақ жоспары дайын',
    required: 'Алдымен тақырып пен топ деңгейін көрсетіңіз.',
    failed: 'Сабақ жоспарын жасау мүмкін болмады. Қайталап көріңіз.',
    minute: 'минут',
  },
}

const outputLanguage: Record<Locale, 'English' | 'Russian' | 'Kazakh'> = {
  en: 'English', ru: 'Russian', kk: 'Kazakh',
}

export function TeachingPlanGenerator() {
  const { locale } = useI18n()
  const text = copy[locale]
  const [topic, setTopic] = useState('')
  const [audience, setAudience] = useState('')
  const [duration, setDuration] = useState('60')
  const [plan, setPlan] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)

  const handleGenerate = async () => {
    if (topic.trim().length < 3 || audience.trim().length < 2) {
      toast.error(text.required)
      return
    }

    setIsGenerating(true)
    try {
      const result = await generateTeachingPlan({
        topic: topic.trim(),
        audience: audience.trim(),
        duration_minutes: Number(duration),
        language: outputLanguage[locale],
      })
      setPlan(result.plan)
      toast.success(text.ready)
    } catch (error) {
      toast.error(getApiErrorMessage(error, text.failed))
    } finally {
      setIsGenerating(false)
    }
  }

  return (
    <div className="space-y-6">
      <section className="ai-generator-surface rounded-2xl border border-violet-100 bg-[linear-gradient(135deg,_rgba(245,243,255,0.96),_rgba(255,255,255,0.98)_58%,_rgba(236,253,245,0.9))] p-4 sm:p-5">
        <div className="mb-5 flex items-start gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-violet-600 text-white shadow-[0_8px_18px_rgba(124,58,237,0.2)]"><ClipboardCheck className="h-5 w-5" /></span>
          <div>
            <h3 className="text-sm font-semibold tracking-[-0.02em] text-violet-950">{text.title}</h3>
            <p className="mt-1 max-w-2xl text-xs leading-relaxed text-violet-950/65">{text.description}</p>
          </div>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="teaching-plan-topic" className="text-violet-950/75">{text.topic}</Label>
            <Input id="teaching-plan-topic" value={topic} onChange={(event) => setTopic(event.target.value)} placeholder={text.topicPlaceholder} className="border-violet-100 bg-white/90" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="teaching-plan-audience" className="text-violet-950/75">{text.audience}</Label>
            <Input id="teaching-plan-audience" value={audience} onChange={(event) => setAudience(event.target.value)} placeholder={text.audiencePlaceholder} className="border-violet-100 bg-white/90" />
          </div>
          <div className="space-y-1.5">
            <Label className="text-violet-950/75">{text.duration}</Label>
            <Select value={duration} onValueChange={setDuration}>
              <SelectTrigger className="border-violet-100 bg-white/90"><SelectValue /></SelectTrigger>
              <SelectContent>{[30, 45, 60, 90, 120].map((minutes) => <SelectItem key={minutes} value={String(minutes)}>{minutes} {text.minute}</SelectItem>)}</SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-violet-950/75">{text.language}</Label>
            <div className="flex h-10 items-center rounded-md border border-violet-100 bg-white/70 px-3 text-sm text-violet-950/75">{outputLanguage[locale]}</div>
          </div>
        </div>

        <div className="mt-5 flex justify-end border-t border-violet-100 pt-4">
          <Button onClick={handleGenerate} disabled={isGenerating} className="w-full rounded-xl bg-violet-600 shadow-[0_4px_12px_rgba(124,58,237,0.2)] hover:bg-violet-700 sm:w-auto">
            {isGenerating ? <><Loader2 className="h-4 w-4 animate-spin" /> {text.generating}</> : <><Sparkles className="h-4 w-4" /> {text.generate}</>}
          </Button>
        </div>
      </section>

      <AnimatePresence mode="wait">
        {plan && (
          <motion.article key={plan} initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -8 }} transition={{ duration: 0.25 }} className="overflow-hidden rounded-2xl border border-violet-100 bg-white shadow-[0_12px_28px_rgba(76,29,149,0.08)]">
            <div className="flex items-center gap-2 border-b border-violet-100 bg-violet-50/70 px-5 py-3">
              <ClipboardCheck className="h-4 w-4 text-violet-700" />
              <p className="text-xs font-semibold uppercase tracking-[0.12em] text-violet-900">{text.result}</p>
            </div>
            <div className="p-5 sm:p-6"><p className="whitespace-pre-wrap text-sm leading-7 text-slate-700">{plan}</p></div>
          </motion.article>
        )}
      </AnimatePresence>
    </div>
  )
}
