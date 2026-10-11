'use client'

import { useState } from 'react'
import { Sparkles, Loader2, ChevronDown, ChevronUp, CheckCircle2, AlertCircle, Lightbulb } from 'lucide-react'
import { toast } from 'sonner'
import { useAssignmentFeedback } from '@/lib/queries/ai.queries'
import { AssignmentFeedbackDTO } from '@/lib/api/ai'
import { Button } from '@/components/ui/button'

interface AssignmentFeedbackPanelProps {
  submissionId: string
  hasTextContent: boolean
}

function FeedbackSection({ icon, label, items, color }: {
  icon: React.ReactNode
  label: string
  items: string[]
  color: string
}) {
  if (!items?.length) return null
  return (
    <div>
      <div className={`flex items-center gap-1.5 text-xs font-semibold mb-1.5 ${color}`}>
        {icon}
        {label}
      </div>
      <ul className="space-y-1">
        {items.map((item, i) => (
          <li key={i} className="text-sm text-ink-muted pl-3 border-l-2 border-muted leading-relaxed">
            {item}
          </li>
        ))}
      </ul>
    </div>
  )
}

export function AssignmentFeedbackPanel({ submissionId, hasTextContent }: AssignmentFeedbackPanelProps) {
  const [expanded, setExpanded] = useState(false)
  const [feedback, setFeedback] = useState<AssignmentFeedbackDTO | null>(null)
  const getFeedback = useAssignmentFeedback()

  const handleGetFeedback = async () => {
    try {
      const result = await getFeedback.mutateAsync(submissionId)
      setFeedback(result)
      setExpanded(true)
      toast.success('AI feedback generated')
    } catch {
      toast.error('Failed to generate feedback. Please try again later.')
    }
  }

  if (!hasTextContent) return null

  return (
    <div className="rounded-lg border border-violet-200 bg-violet-50/50 overflow-hidden">
      <div className="px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-violet-600" />
          <span className="text-sm font-medium text-violet-800">AI Feedback</span>
        </div>
        <div className="flex items-center gap-2">
          {feedback && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="text-violet-600 hover:text-violet-800"
            >
              {expanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </button>
          )}
          <Button
            size="sm"
            variant="ghost"
            onClick={handleGetFeedback}
            disabled={getFeedback.isPending}
            className="h-7 text-xs text-violet-700 hover:text-violet-900 hover:bg-violet-100 gap-1.5"
          >
            {getFeedback.isPending ? (
              <><Loader2 className="h-3 w-3 animate-spin" /> Analyzing…</>
            ) : feedback ? (
              'Refresh'
            ) : (
              'Get feedback'
            )}
          </Button>
        </div>
      </div>

      {feedback && expanded && (
        <div className="px-4 pb-4 space-y-4 border-t border-violet-200 pt-3">
          <p className="text-sm text-ink leading-relaxed italic">{feedback.overall}</p>

          <FeedbackSection
            icon={<CheckCircle2 className="h-3.5 w-3.5" />}
            label="Strengths"
            items={feedback.strengths}
            color="text-green-700"
          />
          <FeedbackSection
            icon={<AlertCircle className="h-3.5 w-3.5" />}
            label="Gaps"
            items={feedback.gaps}
            color="text-amber-700"
          />
          <FeedbackSection
            icon={<Lightbulb className="h-3.5 w-3.5" />}
            label="Suggestions"
            items={feedback.improvements}
            color="text-blue-700"
          />
        </div>
      )}
    </div>
  )
}
