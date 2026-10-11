'use client'

import { PageHeader } from '@/components/shared/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ProgressChart } from '@/components/analytics/ProgressChart'
import { Skeleton } from '@/components/ui/skeleton'
import { useTeacherOverview } from '@/lib/queries/teacher.queries'
import { useI18n } from '@/lib/i18n'

export default function TeacherAnalyticsPage() {
  const { data, isLoading } = useTeacherOverview()
  const { t } = useI18n()
  if (isLoading) return <div className="space-y-6"><Skeleton className="h-8 w-48" /><Skeleton className="h-40 w-full" /></div>
  const chartData = [{ label: t('students'), value: data?.total_students ?? 0 }, { label: t('groups'), value: data?.total_groups ?? 0 }, { label: t('published'), value: data?.published_courses ?? 0 }, { label: t('drafts'), value: data?.draft_courses ?? 0 }]
  return (
    <div className="space-y-6">
      <PageHeader title={t('analytics')} description={t('assignedPerformance')} />
      <div className="grid gap-4 sm:grid-cols-3">{[[t('courses'), data?.total_courses ?? 0], [t('students'), data?.total_students ?? 0], [t('groups'), data?.total_groups ?? 0]].map(([label, value]) => <Card key={String(label)}><CardContent className="pt-5"><p className="text-xs text-[#6b6b6b]">{label}</p><p className="mt-1 text-2xl font-semibold">{value}</p></CardContent></Card>)}</div>
      <Card>
        <CardHeader><CardTitle className="text-sm">{t('teachingOverview')}</CardTitle></CardHeader>
        <CardContent><ProgressChart data={chartData} /></CardContent>
      </Card>
    </div>
  )
}
