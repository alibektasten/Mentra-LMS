'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ChevronLeft, GraduationCap, Clock, AlertCircle, CheckCircle2, Upload } from 'lucide-react'
import { toast } from 'sonner'
import { useExam, useMyAttempts, useStartExamAttempt, useSubmitExamAttempt, useSaveExamDraft } from '@/lib/queries/exams.queries'
import { ExamAttemptDTO, ExamMCQAnswerInput, StartAttemptResponse } from '@/types/exam'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils/cn'

type Phase = 'landing' | 'active' | 'submitted'

function ExamCountdown({ expiresAt, onExpire }: { expiresAt: string; onExpire: () => void }) {
  const [secondsLeft, setSecondsLeft] = useState(() => {
    const diff = Math.floor((new Date(expiresAt).getTime() - Date.now()) / 1000)
    return Math.max(0, diff)
  })

  useEffect(() => {
    if (secondsLeft <= 0) { onExpire(); return }
    const timer = setInterval(() => {
      setSecondsLeft((s) => {
        if (s <= 1) { clearInterval(timer); onExpire(); return 0 }
        return s - 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [expiresAt, onExpire, secondsLeft])

  const mins = Math.floor(secondsLeft / 60)
  const secs = secondsLeft % 60
  const urgent = secondsLeft <= 60

  return (
    <div className={cn(
      'inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border text-sm font-mono font-semibold',
      urgent
        ? 'bg-red-50 text-red-600 border-red-200'
        : 'bg-amber-50 text-amber-700 border-amber-200',
    )}>
      <Clock className="h-4 w-4" />
      {String(mins).padStart(2, '0')}:{String(secs).padStart(2, '0')}
    </div>
  )
}

export default function StudentExamPage() {
  const { org, courseId, examId } = useParams<{ org: string; courseId: string; examId: string }>()
  const { data: exam, isLoading: examLoading } = useExam(examId)
  const { data: attempts, isLoading: attemptsLoading } = useMyAttempts(examId)
  const startAttempt = useStartExamAttempt(examId)
  const submitAttempt = useSubmitExamAttempt(examId)
  const saveDraft = useSaveExamDraft()

  const [phase, setPhase] = useState<Phase>('landing')
  const [activeAttempt, setActiveAttempt] = useState<StartAttemptResponse | null>(null)
  const [mcqAnswers, setMcqAnswers] = useState<Record<string, string>>({})
  const [file, setFile] = useState<File | null>(null)
  const [lastResult, setLastResult] = useState<ExamAttemptDTO | null>(null)
  const autoSubmitRef = useRef(false)
  const hydratedDraftRef = useRef(false)

  // Resume active attempt on load
  useEffect(() => {
    if (!attempts || !exam) return
    const active = attempts.find((a) => a.status === 'in_progress')
    if (active && phase === 'landing') {
      // Re-call start to get exam data back
      startAttempt.mutateAsync().then((res) => {
        setActiveAttempt(res)
        setMcqAnswers(Object.fromEntries(res.saved_answers.map((answer) => [answer.question_id, answer.answer_id])))
        hydratedDraftRef.current = true
        setPhase('active')
      }).catch(() => {})
    }
    // Show last submitted result
    const submitted = [...attempts].reverse().find((a) => a.status === 'submitted')
    if (submitted && phase === 'landing') {
      setLastResult(submitted)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [attempts, exam])

  // Save MCQ answers after a short pause. Files are sent only on final submit,
  // while every selected answer can safely survive a refresh or temporary loss
  // of connectivity.
  useEffect(() => {
    if (phase !== 'active' || !activeAttempt || !hydratedDraftRef.current) return
    const timer = window.setTimeout(() => {
      const answers: ExamMCQAnswerInput[] = Object.entries(mcqAnswers).map(([question_id, answer_id]) => ({ question_id, answer_id }))
      saveDraft.mutate({ attemptID: activeAttempt.attempt_id, mcqAnswers: answers })
    }, 750)
    return () => window.clearTimeout(timer)
  }, [activeAttempt, mcqAnswers, phase, saveDraft])

  const doSubmit = useCallback(async (expired = false) => {
    if (!activeAttempt || autoSubmitRef.current) return
    autoSubmitRef.current = true

    const formData = new FormData()
    formData.append('attempt_id', activeAttempt.attempt_id)

    const answersArr: ExamMCQAnswerInput[] = Object.entries(mcqAnswers).map(([qid, aid]) => ({
      question_id: qid,
      answer_id: aid,
    }))
    formData.append('mcq_answers', JSON.stringify(answersArr))

    if (file) formData.append('file', file)

    try {
      const result = await submitAttempt.mutateAsync({ attemptID: activeAttempt.attempt_id, formData })
      setLastResult(result)
      setPhase('submitted')
      if (expired) toast.info('Time expired — exam auto-submitted')
      else toast.success('Exam submitted')
    } catch {
      autoSubmitRef.current = false
      toast.error('Failed to submit exam')
    }
  }, [activeAttempt, mcqAnswers, file, submitAttempt])

  const handleStart = async () => {
    try {
      const res = await startAttempt.mutateAsync()
      setActiveAttempt(res)
      setMcqAnswers(Object.fromEntries(res.saved_answers.map((answer) => [answer.question_id, answer.answer_id])))
      setFile(null)
      autoSubmitRef.current = false
      hydratedDraftRef.current = true
      setPhase('active')
    } catch (err) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      toast.error(msg ?? 'Failed to start exam')
    }
  }

  const isLoading = examLoading || attemptsLoading
  const attemptList = attempts ?? []
  const totalAttempts = attemptList.length

  if (isLoading) {
    return (
      <div className="max-w-2xl space-y-4 py-6">
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-8 w-56" />
        <Skeleton className="h-40 rounded-xl" />
      </div>
    )
  }

  return (
    <div className="max-w-2xl py-2">
      <Link
        href={`/${org}/student/courses/${courseId}`}
        className="inline-flex items-center gap-1 text-xs text-[#9b9b9b] hover:text-[#1a1a1a] transition-colors mb-5"
      >
        <ChevronLeft className="h-3.5 w-3.5" />
        Back to Course
      </Link>

      {/* ——— LANDING ——— */}
      {phase === 'landing' && exam && (
        <div className="space-y-4">
          <div className="rounded-xl border border-[#e4e2de] bg-white p-6 shadow-sm">
            <div className="flex items-start gap-3 mb-4">
              <div className="h-9 w-9 rounded-lg bg-[#f0eeeb] flex items-center justify-center shrink-0">
                <GraduationCap className="h-5 w-5 text-[#6b6b6b]" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-[#1a1a1a]">{exam.title}</h1>
                {exam.description && (
                  <p className="mt-1 text-sm text-[#6b6b6b] leading-relaxed">{exam.description}</p>
                )}
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3 text-sm mb-5">
              <div className="rounded-lg bg-[#f7f6f3] px-3 py-2">
                <p className="text-xs text-[#9b9b9b] mb-0.5">Duration</p>
                <p className="font-semibold text-[#1a1a1a]">{exam.duration_minutes} minutes</p>
              </div>
              <div className="rounded-lg bg-[#f7f6f3] px-3 py-2">
                <p className="text-xs text-[#9b9b9b] mb-0.5">Total Points</p>
                <p className="font-semibold text-[#1a1a1a]">{exam.total_points} pts</p>
              </div>
              <div className="rounded-lg bg-[#f7f6f3] px-3 py-2">
                <p className="text-xs text-[#9b9b9b] mb-0.5">Attempts</p>
                <p className="font-semibold text-[#1a1a1a]">{totalAttempts} / {exam.max_attempts}</p>
              </div>
              {exam.due_date && (
                <div className="rounded-lg bg-[#f7f6f3] px-3 py-2">
                  <p className="text-xs text-[#9b9b9b] mb-0.5">Due Date</p>
                  <p className="font-semibold text-[#1a1a1a]">{new Date(exam.due_date).toLocaleDateString()}</p>
                </div>
              )}
            </div>

            <div className="flex gap-2 mb-5">
              {exam.mcq_enabled && (
                <span className="text-xs px-2 py-1 rounded-full bg-amber-50 text-amber-700 border border-amber-200">
                  MCQ · {exam.mcq_points} pts
                </span>
              )}
              {exam.file_enabled && (
                <span className="text-xs px-2 py-1 rounded-full bg-sky-50 text-sky-700 border border-sky-200">
                  File Upload · {exam.file_points} pts
                </span>
              )}
            </div>

            <div className="border-t border-[#f0eeeb] pt-4">
              <p className="text-xs text-[#9b9b9b] mb-3 flex items-center gap-1">
                <AlertCircle className="h-3.5 w-3.5" />
                Starting the exam uses one of your {exam.max_attempts} attempt(s). The timer starts immediately.
              </p>
              <Button
                onClick={handleStart}
                disabled={startAttempt.isPending || totalAttempts >= exam.max_attempts}
                className="bg-[#059669] hover:bg-[#047857] text-white"
              >
                {startAttempt.isPending ? 'Starting...' : 'Start Exam'}
              </Button>
              {totalAttempts >= exam.max_attempts && (
                <p className="mt-2 text-xs text-red-600">No attempts remaining.</p>
              )}
            </div>
          </div>

          {/* Past attempts */}
          {attemptList.length > 0 && (
            <div className="rounded-xl border border-[#e4e2de] bg-white p-5 shadow-sm">
              <p className="text-sm font-semibold text-[#1a1a1a] mb-3">Past Attempts</p>
              <div className="space-y-2">
                {attemptList.map((a, i) => (
                  <div key={a.id} className="flex items-center justify-between text-sm">
                    <span className="text-[#6b6b6b]">Attempt {i + 1}</span>
                    <div className="flex items-center gap-3">
                      <span className={cn(
                        'text-[10px] font-semibold uppercase tracking-wide px-1.5 py-0.5 rounded border',
                        a.status === 'submitted' ? 'bg-sky-50 text-sky-700 border-sky-200' :
                        a.status === 'expired' ? 'bg-red-50 text-red-600 border-red-200' :
                        'bg-amber-50 text-amber-700 border-amber-200'
                      )}>
                        {a.status}
                      </span>
                      {a.total_score != null && (
                        <span className="text-xs font-semibold text-[#1a1a1a]">{a.total_score} pts</span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* ——— ACTIVE ——— */}
      {phase === 'active' && activeAttempt && exam && (
        <div className="space-y-4">
          {/* Timer bar */}
          <div className="rounded-xl border border-[#e4e2de] bg-white px-5 py-3 flex items-center justify-between shadow-sm">
            <p className="text-sm font-semibold text-[#1a1a1a]">{exam.title}</p>
            <ExamCountdown
              expiresAt={activeAttempt.expires_at}
              onExpire={() => doSubmit(true)}
            />
          </div>

          {/* MCQ Section */}
          {exam.mcq_enabled && activeAttempt.exam.questions.length > 0 && (
            <div className="rounded-xl border border-[#e4e2de] bg-white p-5 shadow-sm space-y-5">
              <p className="text-sm font-semibold text-[#1a1a1a]">Multiple Choice ({exam.mcq_points} pts)</p>
              {activeAttempt.exam.questions.map((q, qi) => (
                <div key={q.id} className="space-y-2">
                  <p className="text-sm text-[#1a1a1a]">
                    <span className="text-[#9b9b9b] mr-1">{qi + 1}.</span> {q.question}
                  </p>
                  <div className="space-y-1.5 pl-4">
                    {q.answers.map((a) => (
                      <label key={a.id} className="flex items-center gap-2.5 cursor-pointer group">
                        <input
                          type="radio"
                          name={`q-${q.id}`}
                          value={a.id}
                          checked={mcqAnswers[q.id] === a.id}
                          onChange={() => setMcqAnswers((prev) => ({ ...prev, [q.id]: a.id }))}
                          className="accent-[#059669]"
                        />
                        <span className="text-sm text-[#1a1a1a] group-hover:text-[#059669]">{a.answer}</span>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* File Upload Section */}
          {exam.file_enabled && (
            <div className="rounded-xl border border-[#e4e2de] bg-white p-5 shadow-sm space-y-3">
              <p className="text-sm font-semibold text-[#1a1a1a]">File Upload ({exam.file_points} pts)</p>
              <label className="flex flex-col items-center justify-center gap-2 py-8 rounded-lg border-2 border-dashed border-[#e4e2de] cursor-pointer hover:border-[#059669] transition-colors">
                <Upload className="h-6 w-6 text-[#9b9b9b]" />
                <p className="text-sm text-[#6b6b6b]">
                  {file ? file.name : 'Click to select a file'}
                </p>
                <input type="file" className="hidden" onChange={(e) => setFile(e.target.files?.[0] ?? null)} />
              </label>
            </div>
          )}

          {/* Submit */}
          <Button
            size="lg"
            onClick={() => doSubmit(false)}
            disabled={submitAttempt.isPending}
            className="w-full bg-[#059669] hover:bg-[#047857] text-white"
          >
            {submitAttempt.isPending ? 'Submitting...' : 'Submit Exam'}
          </Button>
        </div>
      )}

      {/* ——— SUBMITTED ——— */}
      {phase === 'submitted' && lastResult && exam && (
        <div className="space-y-4">
          <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-5 flex items-start gap-3">
            <CheckCircle2 className="h-5 w-5 text-emerald-600 shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-semibold text-emerald-800">Exam submitted</p>
              <p className="text-xs text-emerald-700 mt-0.5">Your responses have been recorded.</p>
            </div>
          </div>

          <div className="rounded-xl border border-[#e4e2de] bg-white p-5 shadow-sm space-y-4">
            <p className="text-sm font-semibold text-[#1a1a1a]">Results</p>

            {exam.mcq_enabled && (
              <div className="flex items-center justify-between text-sm border-b border-[#f0eeeb] pb-3">
                <span className="text-[#6b6b6b]">MCQ Score</span>
                <span className="font-semibold text-[#1a1a1a]">
                  {lastResult.mcq_score != null
                    ? `${lastResult.mcq_score} / ${lastResult.mcq_max_score} pts`
                    : '—'}
                </span>
              </div>
            )}

            {exam.file_enabled && (
              <div className="text-sm border-b border-[#f0eeeb] pb-3 space-y-1">
                <div className="flex items-center justify-between">
                  <span className="text-[#6b6b6b]">File Upload Score</span>
                  <span className="font-semibold text-[#1a1a1a]">
                    {lastResult.file_score != null
                      ? `${lastResult.file_score} / ${lastResult.file_points} pts`
                      : 'Awaiting teacher review'}
                  </span>
                </div>
                {lastResult.file_feedback && (
                  <p className="text-xs text-[#6b6b6b] mt-1">Feedback: {lastResult.file_feedback}</p>
                )}
              </div>
            )}

            <div className="flex items-center justify-between text-sm">
              <span className="font-semibold text-[#1a1a1a]">Total Score</span>
              <span className="text-lg font-bold text-[#059669]">
                {lastResult.total_score != null ? `${lastResult.total_score} pts` : 'Pending'}
              </span>
            </div>
          </div>

          <Button variant="outline" onClick={() => setPhase('landing')} className="w-full">
            Back to Exam Info
          </Button>
        </div>
      )}
    </div>
  )
}
