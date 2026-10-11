'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import { Play, Sparkles, TrendingUp, Award, ArrowRight } from 'lucide-react'
import { ProgressDTO } from '@/types/progress'
import { useI18n } from '@/lib/i18n'

interface QuickActionsProps {
  lastProgress?: ProgressDTO
}

export function QuickActions({ lastProgress }: QuickActionsProps) {
  const { org } = useParams<{ org: string }>()
  const { t } = useI18n()

  const actions = [
    {
      icon: Play,
      label: lastProgress ? t('resumeLearning') : t('startLearning'),
      href: `/${org}/student/courses`,
    },
    {
      icon: Sparkles,
      label: t('quizGenerator'),
      href: `/${org}/student/ai`,
    },
    {
      icon: TrendingUp,
      label: t('viewProgress'),
      href: `/${org}/student/progress`,
    },
    {
      icon: Award,
      label: t('certificates'),
      href: `/${org}/student/certificates`,
    },
  ]

  return (
    <div className="flex flex-col">
      <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest mb-3">{t('quickActions')}</p>
      <div className="divide-y divide-[#e8e8e6] border border-[#e8e8e6] rounded-lg overflow-hidden">
        {actions.map(({ icon: Icon, label, href }) => (
          <Link
            key={label}
            href={href}
            className="group flex items-center gap-3 px-3.5 py-2.5 hover:bg-[#f7f7f5] transition-colors"
          >
            <Icon className="h-3.5 w-3.5 shrink-0 text-[#9b9b9b]" />
            <span className="flex-1 text-sm text-[#3b3b3b] group-hover:text-[#1a1a1a] transition-colors">
              {label}
            </span>
            <ArrowRight className="h-3.5 w-3.5 text-[#c9c9c9] opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
          </Link>
        ))}
      </div>
    </div>
  )
}
