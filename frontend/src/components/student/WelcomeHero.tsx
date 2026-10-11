'use client'

import { useAuthStore } from '@/lib/stores/auth.store'
import { Flame } from 'lucide-react'
import { useI18n } from '@/lib/i18n'

interface WelcomeHeroProps {
  streak: number
}

export function WelcomeHero({ streak }: WelcomeHeroProps) {
  const { user } = useAuthStore()
  const { locale, t } = useI18n()

  const now = new Date()
  const hour = now.getHours()
  const greeting = hour < 12 ? t('goodMorning') : hour < 17 ? t('goodAfternoon') : t('goodEvening')

  const dateStr = now.toLocaleDateString(locale === 'kk' ? 'kk-KZ' : locale === 'ru' ? 'ru-RU' : 'en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  })

  return (
    <div className="flex items-center justify-between py-1">
      <div>
        <p className="text-xs text-[#9b9b9b]">{dateStr}</p>
        <h1 className="mt-0.5 text-2xl font-bold tracking-tight text-[#1a1a1a]">
          {greeting}, {user?.first_name ?? t('student')}
        </h1>
      </div>

      {streak > 0 && (
        <div className="flex items-center gap-1.5 rounded-lg border border-[#e8e8e6] bg-[#fbfbfa] px-3 py-1.5 text-sm">
          <Flame className="h-3.5 w-3.5 text-orange-400" />
          <span className="font-semibold text-[#1a1a1a]">{streak}</span>
          <span className="text-[#9b9b9b] text-xs">{t('dayStreak')}</span>
        </div>
      )}
    </div>
  )
}
