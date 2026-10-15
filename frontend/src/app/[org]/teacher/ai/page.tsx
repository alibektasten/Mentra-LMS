'use client'

import { PageHeader } from '@/components/shared/PageHeader'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { QuizGenerator } from '@/components/ai/QuizGenerator'
import { LessonSummarizer } from '@/components/ai/LessonSummarizer'
import { FlashcardGenerator } from '@/components/ai/FlashcardGenerator'
import { TeachingPlanGenerator } from '@/components/ai/TeachingPlanGenerator'
import { useI18n } from '@/lib/i18n'

export default function TeacherAIPage() {
  const { locale } = useI18n()
  const copy = {
    en: { title: 'AI Tools', description: 'AI-powered teaching assistants', plan: 'Lesson Plan', quiz: 'Quiz Generator', cards: 'Flashcard Generator', summary: 'Lesson Summarizer', planCard: 'Plan a lesson', quizCard: 'Generate quiz', cardsCard: 'Generate flashcards', summaryCard: 'Summarize lesson' },
    ru: { title: 'AI-инструменты', description: 'AI-помощники для подготовки занятий', plan: 'План занятия', quiz: 'Генератор тестов', cards: 'Генератор карточек', summary: 'Конспект урока', planCard: 'Спланировать занятие', quizCard: 'Создать тест', cardsCard: 'Создать карточки', summaryCard: 'Создать конспект' },
    kk: { title: 'AI құралдары', description: 'Сабақ дайындауға арналған AI көмекшілері', plan: 'Сабақ жоспары', quiz: 'Тест генераторы', cards: 'Карточка генераторы', summary: 'Сабақ конспектісі', planCard: 'Сабақты жоспарлау', quizCard: 'Тест жасау', cardsCard: 'Карточкалар жасау', summaryCard: 'Конспект жасау' },
  }[locale]
  return (
    <div className="space-y-6">
      <PageHeader title={copy.title} description={copy.description} />
      <Tabs defaultValue="plan">
        <TabsList>
          <TabsTrigger value="plan">{copy.plan}</TabsTrigger>
          <TabsTrigger value="quiz">{copy.quiz}</TabsTrigger>
          <TabsTrigger value="flashcards">{copy.cards}</TabsTrigger>
          <TabsTrigger value="summary">{copy.summary}</TabsTrigger>
        </TabsList>
        <TabsContent value="plan" className="mt-4">
          <Card><CardHeader><CardTitle className="text-sm">{copy.planCard}</CardTitle></CardHeader><CardContent><TeachingPlanGenerator /></CardContent></Card>
        </TabsContent>
        <TabsContent value="quiz" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">{copy.quizCard}</CardTitle></CardHeader>
            <CardContent><QuizGenerator /></CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="flashcards" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">{copy.cardsCard}</CardTitle></CardHeader>
            <CardContent><FlashcardGenerator /></CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="summary" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">{copy.summaryCard}</CardTitle></CardHeader>
            <CardContent><LessonSummarizer /></CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
