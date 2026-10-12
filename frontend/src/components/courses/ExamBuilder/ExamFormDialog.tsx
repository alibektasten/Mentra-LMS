'use client'

import { useState, useEffect } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { useCreateExam, useUpdateExam } from '@/lib/queries/exams.queries'
import { ExamListItemDTO, CreateExamQuestionInput } from '@/types/exam'

interface LocalAnswer {
  answer: string
  is_correct: boolean
}

interface LocalQuestion {
  question: string
  position: number
  answers: LocalAnswer[]
}

function emptyQuestion(position: number): LocalQuestion {
  return {
    question: '',
    position,
    answers: [
      { answer: '', is_correct: false },
      { answer: '', is_correct: false },
    ],
  }
}

interface ExamFormDialogProps {
  courseId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  editExam?: ExamListItemDTO | null
}

export function ExamFormDialog({ courseId, open, onOpenChange, editExam }: ExamFormDialogProps) {
  const createExam = useCreateExam(courseId)
  const updateExam = useUpdateExam(courseId)

  const [title, setTitle] = useState('')
  const [assessmentType, setAssessmentType] = useState<'midterm' | 'endterm' | 'final' | 'other'>('midterm')
  const [description, setDescription] = useState('')
  const [durationMinutes, setDurationMinutes] = useState(60)
  const [maxAttempts, setMaxAttempts] = useState(1)
  const [dueDate, setDueDate] = useState('')
  const [mcqEnabled, setMcqEnabled] = useState(true)
  const [mcqPoints, setMcqPoints] = useState(100)
  const [fileEnabled, setFileEnabled] = useState(false)
  const [filePoints, setFilePoints] = useState(100)
  const [questions, setQuestions] = useState<LocalQuestion[]>([emptyQuestion(0)])

  useEffect(() => {
    if (open) {
      if (editExam) {
        setTitle(editExam.title)
        setAssessmentType(editExam.assessment_type ?? 'other')
        setDescription(editExam.description || '')
        setDurationMinutes(editExam.duration_minutes)
        setMaxAttempts(editExam.max_attempts)
        setDueDate(editExam.due_date ? editExam.due_date.slice(0, 16) : '')
        setMcqEnabled(editExam.mcq_enabled)
        setMcqPoints(editExam.mcq_points)
        setFileEnabled(editExam.file_enabled)
        setFilePoints(editExam.file_points)
        setQuestions([emptyQuestion(0)])
      } else {
        setTitle('')
        setAssessmentType('midterm')
        setDescription('')
        setDurationMinutes(60)
        setMaxAttempts(1)
        setDueDate('')
        setMcqEnabled(true)
        setMcqPoints(100)
        setFileEnabled(false)
        setFilePoints(100)
        setQuestions([emptyQuestion(0)])
      }
    }
  }, [open, editExam])

  const addQuestion = () => setQuestions((p) => [...p, emptyQuestion(p.length)])

  const removeQuestion = (qi: number) =>
    setQuestions((p) => p.filter((_, i) => i !== qi).map((q, i) => ({ ...q, position: i })))

  const updateQuestion = (qi: number, text: string) =>
    setQuestions((p) => p.map((q, i) => (i === qi ? { ...q, question: text } : q)))

  const addAnswer = (qi: number) =>
    setQuestions((p) =>
      p.map((q, i) => (i === qi ? { ...q, answers: [...q.answers, { answer: '', is_correct: false }] } : q))
    )

  const removeAnswer = (qi: number, ai: number) =>
    setQuestions((p) =>
      p.map((q, i) => (i === qi ? { ...q, answers: q.answers.filter((_, j) => j !== ai) } : q))
    )

  const updateAnswer = (qi: number, ai: number, text: string) =>
    setQuestions((p) =>
      p.map((q, i) =>
        i === qi ? { ...q, answers: q.answers.map((a, j) => (j === ai ? { ...a, answer: text } : a)) } : q
      )
    )

  const setCorrect = (qi: number, ai: number) =>
    setQuestions((p) =>
      p.map((q, i) =>
        i === qi ? { ...q, answers: q.answers.map((a, j) => ({ ...a, is_correct: j === ai })) } : q
      )
    )

  const handleSubmit = async () => {
    if (!title.trim()) { toast.error('Title is required'); return }
    if (!mcqEnabled && !fileEnabled) { toast.error('At least one section must be enabled'); return }
    if (mcqEnabled) {
      for (const q of questions) {
        if (!q.question.trim()) { toast.error('All questions must have text'); return }
        if (!q.answers.some((a) => a.is_correct)) { toast.error('Each question must have a correct answer'); return }
      }
    }

    const questionPayload: CreateExamQuestionInput[] = mcqEnabled
      ? questions.map((q) => ({
          question: q.question,
          position: q.position,
          answers: q.answers.map((a) => ({ answer: a.answer, is_correct: a.is_correct })),
        }))
      : []

    const payload = {
		assessment_type: assessmentType,
      title: title.trim(),
      description: description.trim(),
      duration_minutes: durationMinutes,
      max_attempts: maxAttempts,
      due_date: dueDate ? new Date(dueDate).toISOString() : null,
      mcq_enabled: mcqEnabled,
      mcq_points: mcqEnabled ? mcqPoints : 0,
      file_enabled: fileEnabled,
      file_points: fileEnabled ? filePoints : 0,
      questions: questionPayload,
    }

    try {
      if (editExam) {
        await updateExam.mutateAsync({ id: editExam.id, data: payload })
        toast.success('Exam updated')
      } else {
        await createExam.mutateAsync(payload)
        toast.success('Exam created')
      }
      onOpenChange(false)
    } catch {
      toast.error(editExam ? 'Failed to update exam' : 'Failed to create exam')
    }
  }

  const isPending = createExam.isPending || updateExam.isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{editExam ? 'Edit Exam' : 'New Exam'}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4 py-2">
          {/* Title */}
          <div className="space-y-1.5">
            <Label>Assessment type <span className="text-destructive">*</span></Label>
            <Select value={assessmentType} onValueChange={(value) => setAssessmentType(value as typeof assessmentType)}>
              <SelectTrigger className="border-[#e4e2de]"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="midterm">Midterm - 12.5% of final grade</SelectItem>
                <SelectItem value="endterm">Endterm - 12.5% of final grade</SelectItem>
                <SelectItem value="final">Final exam - 50% of final grade</SelectItem>
                <SelectItem value="other">Other assessment</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-[#9b9b9b]">Each course can have one midterm, one endterm, and one final exam.</p>
          </div>
          <div className="space-y-1.5">
            <Label>Title <span className="text-destructive">*</span></Label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Exam title"
              className="border-[#e4e2de]"
            />
          </div>

          {/* Description */}
          <div className="space-y-1.5">
            <Label>Description</Label>
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              placeholder="Optional description"
              className="border-[#e4e2de] resize-none"
            />
          </div>

          {/* Duration / Max Attempts */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>Duration (minutes) <span className="text-destructive">*</span></Label>
              <Input
                type="number"
                min={1}
                value={durationMinutes}
                onChange={(e) => setDurationMinutes(Number(e.target.value))}
                className="border-[#e4e2de]"
              />
            </div>
            <div className="space-y-1.5">
              <Label>Max Attempts</Label>
              <Input
                type="number"
                min={1}
                value={maxAttempts}
                onChange={(e) => setMaxAttempts(Number(e.target.value))}
                className="border-[#e4e2de]"
              />
            </div>
          </div>

          {/* Due Date */}
          <div className="space-y-1.5">
            <Label>Due Date (optional)</Label>
            <Input
              type="datetime-local"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              className="border-[#e4e2de]"
            />
          </div>

          {/* MCQ Section Toggle */}
          <div className="rounded-lg border border-[#e4e2de] p-4 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-semibold text-[#1a1a1a]">MCQ Section</p>
                <p className="text-xs text-[#9b9b9b]">Auto-graded multiple choice questions</p>
              </div>
              <button
                type="button"
                onClick={() => setMcqEnabled((v) => !v)}
                className={`relative h-5 w-9 rounded-full transition-colors ${mcqEnabled ? 'bg-[#059669]' : 'bg-[#d4d2ce]'}`}
              >
                <span className={`absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-transform ${mcqEnabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
              </button>
            </div>
            {mcqEnabled && (
              <div className="space-y-1.5">
                <Label className="text-xs">MCQ Points</Label>
                <Input
                  type="number"
                  min={0}
                  value={mcqPoints}
                  onChange={(e) => setMcqPoints(Number(e.target.value))}
                  className="border-[#e4e2de] w-32"
                />
              </div>
            )}
          </div>

          {/* File Upload Section Toggle */}
          <div className="rounded-lg border border-[#e4e2de] p-4 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-semibold text-[#1a1a1a]">File Upload Section</p>
                <p className="text-xs text-[#9b9b9b]">Manually graded file submission</p>
              </div>
              <button
                type="button"
                onClick={() => setFileEnabled((v) => !v)}
                className={`relative h-5 w-9 rounded-full transition-colors ${fileEnabled ? 'bg-[#059669]' : 'bg-[#d4d2ce]'}`}
              >
                <span className={`absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-transform ${fileEnabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
              </button>
            </div>
            {fileEnabled && (
              <div className="space-y-1.5">
                <Label className="text-xs">File Points</Label>
                <Input
                  type="number"
                  min={0}
                  value={filePoints}
                  onChange={(e) => setFilePoints(Number(e.target.value))}
                  className="border-[#e4e2de] w-32"
                />
              </div>
            )}
          </div>

          {/* Question builder (only when MCQ enabled) */}
          {mcqEnabled && (
            <div className="space-y-3">
              <Label className="text-sm font-semibold">Questions</Label>
              {questions.map((q, qi) => (
                <div key={qi} className="rounded-md border border-[#e4e2de] p-3 space-y-2">
                  <div className="flex items-start gap-2">
                    <span className="text-xs text-[#9b9b9b] pt-2 w-5 shrink-0">{qi + 1}.</span>
                    <Input
                      value={q.question}
                      onChange={(e) => updateQuestion(qi, e.target.value)}
                      placeholder="Question text"
                      className="flex-1 border-[#e4e2de]"
                    />
                    <Button
                      type="button"
                      size="sm"
                      variant="ghost"
                      className="h-8 w-8 p-0 text-destructive hover:text-destructive shrink-0"
                      onClick={() => removeQuestion(qi)}
                      disabled={questions.length <= 1}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                  <div className="pl-7 space-y-1.5">
                    {q.answers.map((a, ai) => (
                      <div key={ai} className="flex items-center gap-2">
                        <input
                          type="radio"
                          name={`correct-${qi}`}
                          checked={a.is_correct}
                          onChange={() => setCorrect(qi, ai)}
                          className="accent-[#059669] shrink-0"
                          title="Mark as correct"
                        />
                        <Input
                          value={a.answer}
                          onChange={(e) => updateAnswer(qi, ai, e.target.value)}
                          placeholder={`Answer ${ai + 1}`}
                          className="flex-1 h-8 text-sm border-[#e4e2de]"
                        />
                        <Button
                          type="button"
                          size="sm"
                          variant="ghost"
                          className="h-8 w-8 p-0 text-destructive hover:text-destructive shrink-0"
                          onClick={() => removeAnswer(qi, ai)}
                          disabled={q.answers.length <= 2}
                        >
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      </div>
                    ))}
                    <Button
                      type="button"
                      size="sm"
                      variant="ghost"
                      className="h-7 text-xs gap-1 text-[#9b9b9b]"
                      onClick={() => addAnswer(qi)}
                    >
                      <Plus className="h-3 w-3" />
                      Add answer
                    </Button>
                  </div>
                </div>
              ))}
              <Button
                type="button"
                size="sm"
                variant="outline"
                className="gap-1.5"
                onClick={addQuestion}
              >
                <Plus className="h-3.5 w-3.5" />
                Add Question
              </Button>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button
            onClick={handleSubmit}
            disabled={isPending}
            className="bg-[#059669] hover:bg-[#047857] text-white"
          >
            {isPending ? 'Saving...' : editExam ? 'Update Exam' : 'Create Exam'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
