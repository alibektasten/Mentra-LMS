'use client'

import { motion } from 'framer-motion'
import { Flame, Sparkles, Target } from 'lucide-react'
import { useFlashcardDeckStore } from '@/lib/stores/flashcard-decks.store'

const DAILY_GOAL = 12

function dayKey(date: Date) {
  return date.toISOString().slice(0, 10)
}

export function FlashcardMomentum({ ownerId }: { ownerId?: string }) {
  const decks = useFlashcardDeckStore((state) => state.decks)
  const learnerDecks = decks.filter((deck) => deck.ownerId === ownerId)

  if (!ownerId || learnerDecks.length === 0) return null

  const activityFor = (key: string) => learnerDecks.reduce(
    (total, deck) => total + (deck.dailyActivity?.[key] ?? 0),
    0
  )
  const today = dayKey(new Date())
  const reviewedToday = activityFor(today)
  const goalProgress = Math.min(100, (reviewedToday / DAILY_GOAL) * 100)

  let streak = 0
  const cursor = new Date()
  while (activityFor(dayKey(cursor)) > 0) {
    streak += 1
    cursor.setUTCDate(cursor.getUTCDate() - 1)
  }

  const cardsNeedingAnotherPass = learnerDecks.reduce(
    (total, deck) => total + Object.values(deck.progress).filter((review) => review.intervalDays === 0).length,
    0
  )

  return (
    <section className="overflow-hidden rounded-2xl border border-[#d9e8df] bg-white shadow-[0_16px_38px_rgba(16,55,37,0.06)]">
      <div className="flex flex-col gap-5 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-700">
              <Target className="h-4 w-4" />
            </span>
            <div>
              <p className="text-sm font-semibold text-[#1b2a22]">Today&apos;s momentum</p>
              <p className="text-xs text-[#748078]">A little recall practice compounds quickly.</p>
            </div>
          </div>
          <div className="mt-4">
            <div className="mb-2 flex items-center justify-between text-xs">
              <span className="font-medium text-[#45574c]">{reviewedToday} of {DAILY_GOAL} cards reviewed</span>
              <span className="font-semibold text-emerald-700">{Math.min(reviewedToday, DAILY_GOAL)}/{DAILY_GOAL}</span>
            </div>
            <div className="h-2 overflow-hidden rounded-full bg-[#e9f1ec]">
              <motion.div
                className="h-full rounded-full bg-[linear-gradient(90deg,#0f9f6e,#34c992)]"
                initial={{ width: 0 }}
                animate={{ width: `${goalProgress}%` }}
                transition={{ type: 'spring', stiffness: 110, damping: 20 }}
              />
            </div>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-2 sm:w-[230px]">
          <div className="rounded-xl border border-orange-100 bg-orange-50/70 p-3">
            <div className="flex items-center gap-1.5 text-orange-700"><Flame className="h-3.5 w-3.5" /><span className="text-[11px] font-semibold">STREAK</span></div>
            <p className="mt-1 text-lg font-bold tracking-[-0.04em] text-[#28362d]">{streak} day{streak === 1 ? '' : 's'}</p>
          </div>
          <div className="rounded-xl border border-violet-100 bg-violet-50/70 p-3">
            <div className="flex items-center gap-1.5 text-violet-700"><Sparkles className="h-3.5 w-3.5" /><span className="text-[11px] font-semibold">REVISIT</span></div>
            <p className="mt-1 text-lg font-bold tracking-[-0.04em] text-[#28362d]">{cardsNeedingAnotherPass} card{cardsNeedingAnotherPass === 1 ? '' : 's'}</p>
          </div>
        </div>
      </div>
    </section>
  )
}
