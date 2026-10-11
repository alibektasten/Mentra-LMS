'use client'

import { useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { CheckCircle2, Eye, EyeOff, Sparkles, XCircle } from 'lucide-react'
import { QuizDTO } from '@/types/quiz'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface GeneratedQuizDisplayProps {
  quiz: QuizDTO
}

export function GeneratedQuizDisplay({ quiz }: GeneratedQuizDisplayProps) {
  const [showAnswers, setShowAnswers] = useState(false)

  return (
    <div className="space-y-4">
      <div className="flex flex-col justify-between gap-3 rounded-2xl border border-violet-100 bg-violet-50/60 px-4 py-3 sm:flex-row sm:items-center">
        <div className="flex items-center gap-2">
          <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-violet-600 text-white"><Sparkles className="h-3.5 w-3.5" /></span>
          <div>
            <p className="text-sm font-semibold tracking-[-0.02em] text-violet-950">Quiz ready to review</p>
            <p className="text-xs text-violet-700/70">{quiz.questions.length} questions with AI-generated explanations</p>
          </div>
        </div>
        <button
          onClick={() => setShowAnswers(!showAnswers)}
          className="inline-flex items-center gap-1.5 self-start rounded-lg bg-white px-3 py-2 text-xs font-semibold text-violet-700 shadow-sm transition-colors hover:bg-violet-100 sm:self-auto"
        >
          {showAnswers ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
          {showAnswers ? 'Hide answers' : 'Reveal answers'}
        </button>
      </div>
      {quiz.questions.map((q, qi) => (
        <motion.div key={q.id} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: qi * 0.05, duration: 0.28 }}>
          <Card className="overflow-hidden border-[#e7e5ef]">
          <CardContent className="p-4 sm:p-5">
            <div className="mb-4 flex items-start gap-3">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-lg bg-violet-100 text-[11px] font-bold text-violet-700">{qi + 1}</span>
              <p className="pt-0.5 text-sm font-semibold leading-relaxed text-[#302b3c]">{q.question}</p>
            </div>
            <div className="space-y-1.5">
              {q.answers.map((a) => (
                <motion.div
                  key={a.id}
                  layout
                  className={`flex items-center gap-2 rounded-xl border px-3 py-2.5 text-sm transition-colors ${
                    showAnswers && a.is_correct
                      ? 'border-emerald-200 bg-emerald-50 text-emerald-900'
                      : 'border-transparent bg-[#f8f8f6] text-[#575750]'
                  }`}
                >
                  <AnimatePresence initial={false}>
                    {showAnswers && (a.is_correct
                      ? <motion.span initial={{ scale: 0 }} animate={{ scale: 1 }} exit={{ scale: 0 }}><CheckCircle2 className="h-4 w-4 shrink-0 text-emerald-600" /></motion.span>
                      : <motion.span initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}><XCircle className="h-4 w-4 shrink-0 text-[#b8b7b1]" /></motion.span>
                    )}
                  </AnimatePresence>
                  <span>{a.answer}</span>
                  {showAnswers && a.is_correct && (
                    <Badge variant="success" className="ml-auto text-xs">Correct</Badge>
                  )}
                </motion.div>
              ))}
            </div>
            <AnimatePresence initial={false}>
              {showAnswers && q.explanation && (
                <motion.div initial={{ opacity: 0, height: 0 }} animate={{ opacity: 1, height: 'auto' }} exit={{ opacity: 0, height: 0 }} className="overflow-hidden">
                  <p className="mt-4 border-t border-[#ecebe7] pt-3 text-xs leading-relaxed text-[#72716b]"><span className="font-semibold text-violet-700">Why it works: </span>{q.explanation}</p>
                </motion.div>
              )}
            </AnimatePresence>
          </CardContent>
          </Card>
        </motion.div>
      ))}
    </div>
  )
}
