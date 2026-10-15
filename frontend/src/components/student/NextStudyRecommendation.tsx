'use client'

import Link from 'next/link'
import type { ReactNode } from 'react'
import { ArrowRight, BookOpenCheck, CircleAlert, CheckCircle2 } from 'lucide-react'
import { useParams } from 'next/navigation'
import { useI18n } from '@/lib/i18n'
import type { NextStudyRecommendation as Recommendation } from '@/lib/personalization/next-study-recommendation'
import { Skeleton } from '@/components/ui/skeleton'

export function NextStudyRecommendation({ recommendation, state = 'ready' }: { recommendation?: Recommendation; state?: 'loading' | 'unavailable' | 'ready' }) {
  const { org } = useParams<{ org: string }>()
  const { locale } = useI18n()
  const copy = getCopy(locale)

  if (state === 'loading') {
    return <div className="rounded-lg border border-[#e8e8e6] bg-[#fbfbfa] px-4 py-3"><Skeleton className="h-3 w-28" /><Skeleton className="mt-3 h-3 w-full" /><Skeleton className="mt-2 h-3 w-3/4" /></div>
  }
  if (state === 'unavailable' || !recommendation) {
    return <Panel icon={<CircleAlert className="h-4 w-4" />} title={copy.title} body={copy.unavailable} />
  }

  if (recommendation.status === 'no-graded-work') {
    return <Panel icon={<BookOpenCheck className="h-4 w-4" />} title={copy.title} body={copy.noGrades} />
  }
  if (recommendation.status === 'all-current-work-mastered') {
    return <Panel icon={<CheckCircle2 className="h-4 w-4" />} title={copy.title} body={copy.mastered} tone="success" />
  }
  if (recommendation.status === 'no-next-material') {
    return <Panel
      icon={<CircleAlert className="h-4 w-4" />}
      title={copy.title}
      body={copy.noMaterial(recommendation.assessment.title, formatScore(recommendation.percentage, locale), recommendation.course.title)}
    />
  }

  const href = `/${org}/student/courses/${recommendation.course.id}/lessons/${recommendation.lesson.lessonID}`
  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50/60 px-4 py-3">
      <div className="flex gap-3">
        <CircleAlert className="mt-0.5 h-4 w-4 shrink-0 text-amber-700" />
        <div className="min-w-0 flex-1">
          <p className="text-xs font-semibold uppercase tracking-widest text-amber-800">{copy.title}</p>
          <p className="mt-1 text-sm font-medium text-[#3b3b3b]">{copy.evidence(recommendation.assessment.title, formatScore(recommendation.percentage, locale), recommendation.course.title)}</p>
          <p className="mt-1 text-xs leading-relaxed text-[#6b6250]">{copy.scope}</p>
          <Link href={href} className="mt-3 flex items-center justify-between gap-3 rounded-md bg-white px-3 py-2 text-sm font-medium text-[#3b3b3b] ring-1 ring-amber-200 transition-colors hover:bg-amber-100">
            <span className="min-w-0"><span className="block truncate">{recommendation.lesson.lessonTitle}</span><span className="block truncate text-[11px] font-normal text-[#8b7b58]">{recommendation.lesson.moduleTitle}</span></span>
            <ArrowRight className="h-4 w-4 shrink-0 text-amber-700" />
          </Link>
        </div>
      </div>
    </div>
  )
}

function Panel({ icon, title, body, tone = 'muted' }: { icon: ReactNode; title: string; body: string; tone?: 'muted' | 'success' }) {
  const colors = tone === 'success' ? 'border-emerald-200 bg-emerald-50/60 text-emerald-700' : 'border-[#e8e8e6] bg-[#fbfbfa] text-[#73736f]'
  return <div className={`flex gap-3 rounded-lg border px-4 py-3 ${colors}`}><span className="mt-0.5 shrink-0">{icon}</span><div><p className="text-xs font-semibold uppercase tracking-widest">{title}</p><p className="mt-1 text-sm leading-relaxed text-[#5d5d58]">{body}</p></div></div>
}

function formatScore(score: number, locale: string) {
  return new Intl.NumberFormat(locale === 'kk' ? 'kk-KZ' : locale === 'ru' ? 'ru-RU' : 'en-US', { maximumFractionDigits: 0 }).format(score) + '%'
}

function getCopy(locale: string) {
  if (locale === 'ru') return {
    title: 'Следующий шаг',
    noGrades: 'Пока нет проверенных работ. Рекомендация появится после первой доступной оценки.',
    mastered: 'По всем доступным проверенным работам результат не ниже 70%. Продолжайте обучение в удобном темпе.',
    unavailable: 'Не удалось загрузить оценки или доступные материалы, поэтому рекомендация пока не показана.',
    noMaterial: (assessment: string, score: string, course: string) => `По работе «${assessment}» в курсе «${course}» получено ${score}, но доступных незавершённых уроков больше нет.`,
    evidence: (assessment: string, score: string, course: string) => `По работе «${assessment}» в курсе «${course}» получено ${score}.`,
    scope: 'Сигналом служит результат ниже 70%. Это рекомендация по загруженным материалам курса: связь работы с отдельной темой пока не задана.',
  }
  if (locale === 'kk') return {
    title: 'Келесі қадам',
    noGrades: 'Әзірге тексерілген жұмыс жоқ. Ұсыныс алғашқы қолжетімді баға шыққаннан кейін пайда болады.',
    mastered: 'Барлық қолжетімді тексерілген жұмыстар бойынша нәтиже 70%-дан төмен емес. Оқуды өз қарқыныңызбен жалғастырыңыз.',
    unavailable: 'Бағалар немесе қолжетімді материалдар жүктелмеді, сондықтан ұсыныс әзірге көрсетілмейді.',
    noMaterial: (assessment: string, score: string, course: string) => `«${course}» курсындағы «${assessment}» жұмысы бойынша ${score} алынды, бірақ аяқталмаған қолжетімді сабақ қалған жоқ.`,
    evidence: (assessment: string, score: string, course: string) => `«${course}» курсындағы «${assessment}» жұмысы бойынша ${score} алынды.`,
    scope: '70%-дан төмен нәтиже сигнал ретінде алынады. Бұл — жүктелген курс материалдары бойынша ұсыныс: жұмыс пен жеке тақырып арасындағы байланыс әзірге көрсетілмеген.',
  }
  return {
    title: 'Next step',
    noGrades: 'There is no graded work yet. A recommendation will appear after your first available grade.',
    mastered: 'Every available graded item is at or above 70%. Continue learning at your own pace.',
    unavailable: 'Grades or available materials could not be loaded, so no recommendation is shown yet.',
    noMaterial: (assessment: string, score: string, course: string) => `You scored ${score} on “${assessment}” in “${course}”, but there is no unfinished available lesson left.`,
    evidence: (assessment: string, score: string, course: string) => `You scored ${score} on “${assessment}” in “${course}”.`,
    scope: 'A result below 70% is used as the signal. This recommendation uses the course material currently loaded on the dashboard; an assessment-to-topic link is not available yet.',
  }
}
