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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'

export default function NewCoursePage() {
  const { org } = useParams<{ org: string }>()
  const router = useRouter()
  const createCourse = useCreateCourse()

  const { register, handleSubmit, setValue, watch, formState: { errors, isSubmitting } } = useForm<CourseFormData>({
    resolver: zodResolver(courseSchema),
    defaultValues: { status: 'draft' },
  })

  const onSubmit = async (data: CourseFormData) => {
    try {
      const course = await createCourse.mutateAsync(data)
      toast.success('Course created!')
      router.push(`/${org}/admin/courses/${course.id}`)
    } catch {
      toast.error('Failed to create course')
    }
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <PageHeader title="New Course" />
      <Card>
        <CardContent className="pt-6">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="title">Title</Label>
              <Input id="title" placeholder="Introduction to..." {...register('title')} />
              {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="description">Description</Label>
              <Textarea id="description" rows={4} placeholder="What will students learn..." {...register('description')} />
              {errors.description && <p className="text-xs text-destructive">{errors.description.message}</p>}
            </div>
            <div className="space-y-1.5">
              <Label>Status</Label>
              <Select value={watch('status')} onValueChange={(v) => setValue('status', v as 'draft' | 'published' | 'archived')}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="draft">Draft</SelectItem>
                  <SelectItem value="published">Published</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex gap-3 pt-2">
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? 'Creating...' : 'Create Course'}
              </Button>
              <Button type="button" variant="outline" onClick={() => router.back()}>
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
