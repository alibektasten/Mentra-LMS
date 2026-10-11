'use client'

import { BookOpen, Users, TrendingUp, Sparkles } from 'lucide-react'
import { useCourses } from '@/lib/queries/courses.queries'
import { useMembers } from '@/lib/queries/members.queries'
import { PageHeader } from '@/components/shared/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useI18n } from '@/lib/i18n'

function StatCard({ title, value, icon: Icon, description }: {
  title: string; value: string | number; icon: React.ElementType; description?: string
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-ink-muted">{title}</CardTitle>
        <Icon className="h-4 w-4 text-ink-muted" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold text-ink">{value}</div>
        {description && <p className="text-xs text-ink-muted mt-1">{description}</p>}
      </CardContent>
    </Card>
  )
}

export default function AdminDashboard() {
  const { t } = useI18n()
  const { data: courses, isLoading } = useCourses({ page: 1, page_size: 100 })
  const { data: members } = useMembers({ page: 1, page_size: 1, role: 'student' })

  const totalCourses = courses?.meta?.total ?? 0
  const publishedCourses = courses?.data?.filter((c) => c.status === 'published').length ?? 0
  const totalStudents = members?.meta?.total ?? 0

  return (
    <div className="space-y-6">
      <PageHeader title={t('dashboard')} description={t('organizationOverview')} />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-28" />)
        ) : (
          <>
            <StatCard title={t('totalCourses')} value={totalCourses} icon={BookOpen} />
            <StatCard title={t('published')} value={publishedCourses} icon={TrendingUp} description={t('activeCoursesDescription')} />
            <StatCard title={t('students')} value={totalStudents} icon={Users} description={t('enrolledMembers')} />
            <StatCard title={t('aiInsights')} value={t('ready')} icon={Sparkles} description={t('generateReports')} />
          </>
        )}
      </div>
      <Card>
        <CardHeader>
          <CardTitle>{t('courses')}</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
            </div>
          ) : (
            <div className="space-y-2">
              {(courses?.data ?? []).slice(0, 5).map((course) => (
                <div key={course.id} className="flex items-center justify-between rounded-md p-2 hover:bg-muted/50">
                  <span className="text-sm font-medium">{course.title}</span>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${course.status === 'published' ? 'bg-green-100 text-green-700' : 'bg-muted text-ink-muted'}`}>
                    {course.status}
                  </span>
                </div>
              ))}
              {(courses?.data ?? []).length === 0 && (
                <p className="text-sm text-ink-muted">{t('noCoursesYet')}</p>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
