'use client'

import { PageHeader } from '@/components/shared/PageHeader'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { QuizGenerator } from '@/components/ai/QuizGenerator'
import { LessonSummarizer } from '@/components/ai/LessonSummarizer'
import { FlashcardGenerator } from '@/components/ai/FlashcardGenerator'
import { useI18n } from '@/lib/i18n'

export default function AdminAIPage() {
  const { t } = useI18n()
  return (
    <div className="space-y-6">
      <PageHeader
        title={t('ai')}
        description={t('aiToolsDescription')}
      />
      <Tabs defaultValue="quiz">
        <TabsList>
          <TabsTrigger value="quiz">{t('quizGenerator')}</TabsTrigger>
          <TabsTrigger value="flashcards">{t('flashcardGenerator')}</TabsTrigger>
          <TabsTrigger value="summary">{t('lessonSummarizer')}</TabsTrigger>
        </TabsList>
        <TabsContent value="quiz" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">Generate Quiz from Lesson</CardTitle>
            </CardHeader>
            <CardContent>
              <QuizGenerator />
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="flashcards" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">Generate Flashcards from Lesson</CardTitle>
            </CardHeader>
            <CardContent>
              <FlashcardGenerator />
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="summary" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">Summarize Lesson Content</CardTitle>
            </CardHeader>
            <CardContent>
              <LessonSummarizer />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
