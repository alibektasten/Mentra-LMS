'use client'

import { useParams, useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { useCreateCourse } from '@/lib/queries/courses.queries'
import { courseSchema, CourseFormData } from '@/lib/validators/course.schema'
import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { useI18n } from '@/lib/i18n'

export default function TeacherNewCoursePage() {
  const { org } = useParams<{ org: string }>()
  const router = useRouter()
  const createCourse = useCreateCourse()
  const { t } = useI18n()

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<CourseFormData>({
    resolver: zodResolver(courseSchema),
    defaultValues: { status: 'draft' },
  })

  const onSubmit = async (data: CourseFormData) => {
    try {
      const course = await createCourse.mutateAsync(data)
      toast.success(t('courseCreated'))
      router.push(`/${org}/teacher/courses/${course.id}`)
    } catch {
      toast.error(t('failedToCreateCourse'))
    }
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <PageHeader title={t('newCourse')} />
      <Card>
        <CardContent className="pt-6">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="title">{t('title')}</Label>
              <Input id="title" {...register('title')} />
              {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="description">{t('description')}</Label>
              <Textarea id="description" rows={4} {...register('description')} />
              {errors.description && <p className="text-xs text-destructive">{errors.description.message}</p>}
            </div>
            <p className="text-xs text-[#6b6b6b] rounded-md bg-[#f7f6f3] px-3 py-2.5">
              New courses start as drafts. An organization administrator can publish them when they are ready.
            </p>
            <div className="flex gap-3 pt-2">
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? t('creating') : t('createCourse')}
              </Button>
              <Button type="button" variant="outline" onClick={() => router.back()}>{t('cancel')}</Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
