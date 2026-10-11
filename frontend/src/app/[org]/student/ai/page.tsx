'use client'

import { FileText, Layers } from 'lucide-react'
import { LessonSummarizer } from '@/components/ai/LessonSummarizer'
import { FlashcardGenerator } from '@/components/ai/FlashcardGenerator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useI18n } from '@/lib/i18n'

const tabCls = 'h-8 rounded-none border-b-2 border-transparent px-4 gap-2 text-xs text-[#9b9b9b] font-medium data-[state=active]:border-[#1a1a1a] data-[state=active]:text-[#1a1a1a] data-[state=active]:bg-transparent data-[state=active]:shadow-none'

export default function StudentAIPage() {
  const { t } = useI18n()
  return (
    <div className="max-w-3xl">
      <div className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight text-[#1a1a1a]">{t('studyAi')}</h1>
        <p className="mt-1 text-sm text-[#9b9b9b]">
          Generate flashcards and concise lesson summaries to support your study.
        </p>
      </div>

      <Tabs defaultValue="flashcards">
        <TabsList className="h-8 bg-transparent border-b border-[#e8e8e6] p-0 w-full justify-start rounded-none gap-0 mb-6">
          <TabsTrigger value="flashcards" className={tabCls}>
            <Layers className="h-3.5 w-3.5" />
            {t('flashcardGenerator')}
          </TabsTrigger>
          <TabsTrigger value="summary" className={tabCls}>
            <FileText className="h-3.5 w-3.5" />
            {t('lessonSummarizer')}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="flashcards">
          <div className="border border-[#e8e8e6] rounded-lg overflow-hidden">
            <div className="border-b border-[#e8e8e6] px-5 py-3.5 bg-[#fbfbfa]">
              <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest">{t('flashcardGenerator')}</p>
              <p className="text-xs text-[#9b9b9b] mt-0.5">Generate term/definition cards to study key concepts.</p>
            </div>
            <div className="p-5"><FlashcardGenerator /></div>
          </div>
        </TabsContent>

        <TabsContent value="summary">
          <div className="border border-[#e8e8e6] rounded-lg overflow-hidden">
            <div className="border-b border-[#e8e8e6] px-5 py-3.5 bg-[#fbfbfa]">
              <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest">{t('lessonSummarizer')}</p>
              <p className="text-xs text-[#9b9b9b] mt-0.5">Get a concise AI summary to review key concepts quickly.</p>
            </div>
            <div className="p-5"><LessonSummarizer /></div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
