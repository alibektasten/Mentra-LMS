'use client'

import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { AnimatePresence, motion } from 'framer-motion'
import { Layers3, Loader2, WandSparkles } from 'lucide-react'
import { toast } from 'sonner'
import { FlashcardDTO } from '@/lib/api/ai'
import { createFlashcardDeck, listFlashcardDecks } from '@/lib/api/flashcard-decks'
import { useGenerateFlashcards } from '@/lib/queries/ai.queries'
import { useAuthStore } from '@/lib/stores/auth.store'
import { useFlashcardDeckStore } from '@/lib/stores/flashcard-decks.store'
import { useCourses } from '@/lib/queries/courses.queries'
import { useModules } from '@/lib/queries/modules.queries'
import { useLessons } from '@/lib/queries/lessons.queries'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { FlashcardDeck } from './FlashcardDeck'
import { FlashcardMomentum } from './FlashcardMomentum'
import { FlashcardReviewQueue } from './FlashcardReviewQueue'
import { SavedFlashcardDecks } from './SavedFlashcardDecks'
import { getApiErrorMessage } from '@/lib/api/error-message'
import { useI18n } from '@/lib/i18n'

export function FlashcardGenerator() {
  const { t } = useI18n()
  const [selectedCourse, setSelectedCourse] = useState('')
  const [selectedModule, setSelectedModule] = useState('')
  const [selectedLesson, setSelectedLesson] = useState('')
  const [numCards, setNumCards] = useState(8)
  const [result, setResult] = useState<FlashcardDTO[] | null>(null)
  const [deckVersion, setDeckVersion] = useState(0)
  const [activeDeckId, setActiveDeckId] = useState<string | undefined>()
  const [startIndex, setStartIndex] = useState(0)
  const pendingUploads = useRef(new Set<string>())
  const reviewDeckRef = useRef<HTMLDivElement>(null)

  const { data: courses } = useCourses({ page: 1, page_size: 100 })
  const { data: modules = [] } = useModules(selectedCourse)
  const { data: lessons = [] } = useLessons(selectedModule)
  const generateFlashcards = useGenerateFlashcards()
  const { user } = useAuthStore()
  const createDeck = useFlashcardDeckStore((state) => state.createDeck)
  const saveSyncedDeck = useFlashcardDeckStore((state) => state.saveSyncedDeck)
  const syncRemoteDecks = useFlashcardDeckStore((state) => state.syncRemoteDecks)
  const removeDeck = useFlashcardDeckStore((state) => state.removeDeck)
  const localDecks = useFlashcardDeckStore((state) => state.decks)
  const { data: remoteDecks } = useQuery({
    queryKey: ['flashcard-decks', user?.id],
    queryFn: listFlashcardDecks,
    enabled: Boolean(user?.id),
  })

  useEffect(() => {
    if (!user?.id || !remoteDecks) return
    syncRemoteDecks(user.id, remoteDecks)
  }, [remoteDecks, syncRemoteDecks, user?.id])

  useEffect(() => {
    if (!user?.id || !remoteDecks) return
    localDecks
      .filter((deck) => deck.ownerId === user.id && !deck.synced && !pendingUploads.current.has(deck.id))
      .forEach((deck) => {
        pendingUploads.current.add(deck.id)
        void createFlashcardDeck({ lesson_id: deck.lessonId, title: deck.title, cards: deck.cards })
          .then((remoteDeck) => {
            saveSyncedDeck(user.id, remoteDeck)
            removeDeck(deck.id)
          })
          .catch(() => pendingUploads.current.delete(deck.id))
      })
  }, [localDecks, remoteDecks, removeDeck, saveSyncedDeck, user?.id])

  useEffect(() => {
    if (!result?.length) return
    const frame = window.requestAnimationFrame(() => {
      reviewDeckRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
      reviewDeckRef.current?.focus({ preventScroll: true })
    })
    return () => window.cancelAnimationFrame(frame)
  }, [deckVersion, result])

  const handleGenerate = async () => {
    if (!selectedLesson) {
      toast.error(t('selectLessonFirst'))
      return
    }
    try {
      const cards = await generateFlashcards.mutateAsync({ lessonId: selectedLesson, numCards })
      const lessonTitle = lessons.find((lesson) => lesson.id === selectedLesson)?.title ?? 'AI study deck'
      let deck
      let synced = false

      if (user?.id) {
        try {
          const remoteDeck = await createFlashcardDeck({ lesson_id: selectedLesson, title: lessonTitle, cards })
          deck = saveSyncedDeck(user.id, remoteDeck)
          synced = true
        } catch {
          deck = createDeck({ ownerId: user.id, title: lessonTitle, lessonId: selectedLesson, cards })
        }
      } else {
        deck = createDeck({ ownerId: 'anonymous', title: lessonTitle, lessonId: selectedLesson, cards })
      }

      setResult(deck.cards)
      setActiveDeckId(deck.id)
      setStartIndex(0)
      setDeckVersion((version) => version + 1)
      toast.success(synced ? t('flashcardsGenerated') : t('flashcardsSavedLocal'))
    } catch (error) {
      toast.error(getApiErrorMessage(error, 'Не удалось создать карточки. Попробуйте ещё раз.'))
    }
  }

  return (
    <div className="space-y-6">
      <div className="ai-generator-surface ai-generator-surface--emerald rounded-2xl border border-[#e1ebe5] bg-[linear-gradient(135deg,_rgba(236,253,245,0.85),_rgba(255,255,255,0.9)_52%,_rgba(240,253,250,0.85))] p-4 sm:p-5">
        <div className="mb-5 flex items-start gap-3">
          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-emerald-600 text-white shadow-[0_8px_18px_rgba(5,120,93,0.18)]">
            <Layers3 className="h-4 w-4" />
          </span>
          <div>
            <p className="text-sm font-semibold tracking-[-0.02em] text-[#1b2a22]">{t('buildFocusedDeck')}</p>
            <p className="mt-0.5 text-xs leading-relaxed text-[#728078]">Choose a lesson and let AI turn its key ideas into cards you can actively recall.</p>
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div className="space-y-1.5">
            <Label className="text-[#59665e]">{t('course')}</Label>
          <Select value={selectedCourse} onValueChange={(v) => { setSelectedCourse(v); setSelectedModule(''); setSelectedLesson('') }}>
              <SelectTrigger className="h-10 rounded-xl border-[#dbe5df] bg-white/90 shadow-none"><SelectValue placeholder={t('selectCourse')} /></SelectTrigger>
            <SelectContent>
              {(courses?.data ?? []).map((c) => (
                <SelectItem key={c.id} value={c.id}>{c.title}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-[#59665e]">{t('module')}</Label>
          <Select value={selectedModule} onValueChange={(v) => { setSelectedModule(v); setSelectedLesson('') }} disabled={!selectedCourse}>
              <SelectTrigger className="h-10 rounded-xl border-[#dbe5df] bg-white/90 shadow-none"><SelectValue placeholder={t('selectModule')} /></SelectTrigger>
            <SelectContent>
              {modules.map((m) => (
                <SelectItem key={m.id} value={m.id}>{m.title}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-[#59665e]">{t('lesson')}</Label>
          <Select value={selectedLesson} onValueChange={setSelectedLesson} disabled={!selectedModule}>
              <SelectTrigger className="h-10 rounded-xl border-[#dbe5df] bg-white/90 shadow-none"><SelectValue placeholder={t('selectLesson')} /></SelectTrigger>
            <SelectContent>
              {lessons.map((l) => (
                <SelectItem key={l.id} value={l.id}>{l.title}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-[#59665e]">{t('cards')}</Label>
          <Input
            type="number"
            min={3}
            max={20}
            value={numCards}
            onChange={(e) => setNumCards(Number(e.target.value))}
              className="h-10 rounded-xl border-[#dbe5df] bg-white/90 shadow-none"
            />
          </div>
        </div>

        <div className="mt-5 flex flex-col items-start justify-between gap-3 border-t border-[#dce9e1] pt-4 sm:flex-row sm:items-center">
          <p className="text-xs text-[#758179]">AI will generate {numCards || 0} recall prompts from the selected lesson.</p>
          <Button onClick={handleGenerate} disabled={!selectedLesson || generateFlashcards.isPending} className="w-full rounded-xl sm:w-auto">
            {generateFlashcards.isPending ? (
              <><Loader2 className="h-4 w-4 animate-spin" /> {t('buildingDeck')}</>
            ) : (
              <><WandSparkles className="h-4 w-4" /> {t('generateFlashcards')}</>
            )}
          </Button>
        </div>
      </div>
      <FlashcardReviewQueue
        ownerId={user?.id}
        onStartReview={(deck, cardIndex) => {
          setResult(deck.cards)
          setActiveDeckId(deck.id)
          setStartIndex(cardIndex)
          setDeckVersion((version) => version + 1)
        }}
      />
      <FlashcardMomentum ownerId={user?.id} />
      <SavedFlashcardDecks
        ownerId={user?.id}
        onOpenDeck={(deck) => {
          setResult(deck.cards)
          setActiveDeckId(deck.id)
          setStartIndex(0)
          setDeckVersion((version) => version + 1)
        }}
      />
      <div ref={reviewDeckRef} tabIndex={-1} className="scroll-mt-6 outline-none">
        <AnimatePresence mode="wait">
          {result && result.length > 0 && (
            <motion.div
              key={deckVersion}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
            >
              <FlashcardDeck cards={result} deckId={activeDeckId} initialIndex={startIndex} />
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}
