'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Sparkles, Loader2, RefreshCw } from 'lucide-react'
import * as aiApi from '@/lib/api/ai'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

interface AIStudyPanelProps {
  courseId?: string
  courseTitle?: string
}

export function AIStudyPanel({ courseId, courseTitle }: AIStudyPanelProps) {
  const [enabled, setEnabled] = useState(false)

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['ai', 'insights', courseId],
    queryFn: () => aiApi.getAIInsights(),
    enabled: !!courseId && enabled,
  })

  const handleGenerate = () => {
    if (!enabled) {
      setEnabled(true)
    } else {
      refetch()
    }
  }

  return (
    <div className="flex flex-col">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest">Study AI</p>
        <span className="text-[10px] text-[#9b9b9b]">AI</span>
      </div>

      <div className="border border-[#e8e8e6] rounded-lg overflow-hidden">
        {/* Content area */}
        <div className="px-4 py-4 min-h-[100px]">
          {!enabled && !data && (
            <p className="text-sm text-[#9b9b9b] leading-relaxed">
              {courseTitle
                ? `Get personalized study tips for "${courseTitle}".`
                : 'Select a course to get AI-powered study recommendations.'}
            </p>
          )}

          {isLoading && (
            <div className="space-y-2">
              <Skeleton className="h-3 w-full" />
              <Skeleton className="h-3 w-5/6" />
              <Skeleton className="h-3 w-4/6" />
              <Skeleton className="h-3 w-full mt-3" />
              <Skeleton className="h-3 w-3/4" />
            </div>
          )}

          {data && !isLoading && (
            <div className="space-y-2">
              <p className="text-sm text-[#3b3b3b] leading-relaxed whitespace-pre-wrap">
                {data.insights}
              </p>
              <p className="text-[11px] text-[#9b9b9b]">
                {data.completed_lessons} / {data.total_lessons} lessons · avg {data.average_score.toFixed(0)}%
              </p>
            </div>
          )}
        </div>

        {/* Action row */}
        <div className="border-t border-[#e8e8e6] px-4 py-2.5 bg-[#fbfbfa]">
          <Button
            onClick={handleGenerate}
            disabled={!courseId || isLoading}
            size="sm"
            variant="ghost"
            className="h-7 w-full justify-start gap-2 text-xs text-[#3b3b3b] hover:bg-[#f0efed] hover:text-[#1a1a1a] disabled:opacity-40"
          >
            {isLoading ? (
              <><Loader2 className="h-3.5 w-3.5 animate-spin text-[#9b9b9b]" /> Analyzing...</>
            ) : data ? (
              <><RefreshCw className="h-3.5 w-3.5 text-[#9b9b9b]" /> Refresh insights</>
            ) : (
              <><Sparkles className="h-3.5 w-3.5 text-[#9b9b9b]" /> Generate insights</>
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}
