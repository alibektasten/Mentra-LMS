import { z } from 'zod'

export const courseSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200),
  description: z.string().min(1, 'Description is required'),
  status: z.enum(['draft', 'published', 'archived']).default('draft'),
})

export type CourseFormData = z.infer<typeof courseSchema>
