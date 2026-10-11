'use client'

import { useState } from 'react'
import { ChevronDown, Users } from 'lucide-react'
import { useGroups } from '@/lib/queries/groups.queries'
import { GroupMembers } from '@/components/groups/GroupMembers'
import { EmptyState } from '@/components/shared/EmptyState'
import { Skeleton } from '@/components/ui/skeleton'

interface EnrollmentManagerProps {
  courseId: string
}

export function EnrollmentManager({ courseId }: EnrollmentManagerProps) {
  const { data: groups = [], isLoading } = useGroups(courseId)
  const [openGroupID, setOpenGroupID] = useState<string | null>(groups[0]?.id ?? null)

  if (isLoading) {
    return <div className="space-y-2">{Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
  }

  // A group is the source of truth for course access. It also carries the
  // timetable and responsible teacher, so a separate, disconnected list of
  // "enrollments" would be misleading to administrators.
  if (groups.length === 0) {
    return (
      <EmptyState
        icon={Users}
        title="Нет групп для зачисления"
        description="Сначала назначьте или создайте группу на вкладке «Группы», затем добавьте в неё студентов."
      />
    )
  }

  return (
    <div className="space-y-2">
      <p className="text-sm text-muted-foreground">
        Зачисление организовано по группам: студент получает доступ к курсу сразу после добавления в группу.
      </p>
      {groups.map((group) => {
        const isOpen = openGroupID === group.id
        return (
          <section key={group.id} className="overflow-hidden rounded-xl border border-border bg-card">
            <button
              type="button"
              onClick={() => setOpenGroupID(isOpen ? null : group.id)}
              className="flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-muted/50"
              aria-expanded={isOpen}
            >
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">
                <Users className="h-4 w-4" />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block text-sm font-semibold text-foreground">{group.name}</span>
                <span className="block text-xs text-muted-foreground">Управление зачисленными студентами</span>
              </span>
              <ChevronDown className={`h-4 w-4 text-muted-foreground transition-transform ${isOpen ? 'rotate-180' : ''}`} />
            </button>
            {isOpen && (
              <div className="border-t border-border p-4">
                <GroupMembers groupId={group.id} courseId={courseId} />
              </div>
            )}
          </section>
        )
      })}
    </div>
  )
}
