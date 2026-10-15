'use client'

import { useMemo, useState } from 'react'
import { Download, FileText, Loader2, Pencil, Send, Undo2, Users } from 'lucide-react'
import { toast } from 'sonner'
import { useGradebook, useSetGradebookPublication, useUpdateCourseGrade } from '@/lib/queries/grades.queries'
import type { GradebookStudentDTO } from '@/types/grade'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'

interface GradeBookProps {
  courseId: string
}

function percent(value: number | null | undefined) {
  return value === null || value === undefined ? '—' : `${value.toFixed(1)}%`
}

function csvCell(value: string | number | null | undefined) {
  const content = value === null || value === undefined ? '' : String(value)
  // Avoid formula execution when a CSV is opened in a spreadsheet editor.
  const safe = /^[=+\-@]/.test(content) ? `'${content}` : content
  return `"${safe.replaceAll('"', '""')}"`
}

function downloadCSV(rows: GradebookStudentDTO[], courseId: string) {
  const header = ['Student', 'Email', 'Assignments %', 'Midterm %', 'Endterm %', 'Final %', 'Calculated grade', 'Override', 'Published']
  const entries = rows.map(({ student_name, student_email, grade }) => [
    student_name,
    student_email,
    grade.assignments_percent?.toFixed(1),
    grade.midterm_percent?.toFixed(1),
    grade.endterm_percent?.toFixed(1),
    grade.final_exam_percent?.toFixed(1),
    grade.calculated_final_grade?.toFixed(1),
    grade.override_grade?.toFixed(1),
    grade.is_published ? 'Yes' : 'No',
  ])
  const content = [header, ...entries].map((row) => row.map(csvCell).join(',')).join('\r\n')
  const blob = new Blob([`\uFEFF${content}`], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `mentra-gradebook-${courseId}.csv`
  link.click()
  URL.revokeObjectURL(url)
}

export function GradeBook({ courseId }: GradeBookProps) {
  const { data, isLoading, isError } = useGradebook(courseId)
  const publishResults = useSetGradebookPublication(courseId)
  const [editor, setEditor] = useState<GradebookStudentDTO | null>(null)
  const [selectedStudentIDs, setSelectedStudentIDs] = useState<string[]>([])
  const rows = data?.students ?? []
  const publishedCount = data?.published_count ?? 0
  const selectedRows = selectedStudentIDs.length > 0 ? rows.filter((row) => selectedStudentIDs.includes(row.student_id)) : rows
  const allPublished = selectedRows.length > 0 && selectedRows.every((row) => row.grade.is_published)
  const hasExplicitSelection = selectedStudentIDs.length > 0

  const publish = async (published: boolean) => {
    try {
      const result = await publishResults.mutateAsync({ published, studentIDs: hasExplicitSelection ? selectedStudentIDs : undefined })
      toast.success(published ? `Results published for ${result.affected_count} student${result.affected_count === 1 ? '' : 's'}` : `Results withdrawn for ${result.affected_count} student${result.affected_count === 1 ? '' : 's'}`)
      setSelectedStudentIDs([])
    } catch {
      toast.error('Could not update result publication')
    }
  }

  const toggleStudent = (studentID: string) => {
    setSelectedStudentIDs((current) => current.includes(studentID) ? current.filter((id) => id !== studentID) : [...current, studentID])
  }

  const toggleAll = () => {
    setSelectedStudentIDs((current) => current.length === rows.length ? [] : rows.map((row) => row.student_id))
  }

  if (isLoading) return <Skeleton className="h-56 rounded-lg" />
  if (isError) return <p className="py-4 text-sm text-destructive">Could not load the gradebook.</p>

  return (
    <div className="space-y-4">
      <div className="flex flex-col justify-between gap-3 sm:flex-row sm:items-start">
        <div>
          <p className="text-sm font-semibold text-[#1a1a1a]">Academic register</p>
          <p className="mt-0.5 text-xs text-[#77766f]">Final grade = assignments 25% + midterm 12.5% + endterm 12.5% + final 50%.</p>
        </div>
        <div className="flex flex-wrap gap-2 print:hidden">
          <Button type="button" variant="outline" size="sm" disabled={!rows.length} onClick={() => downloadCSV(rows, courseId)}>
            <Download className="mr-1.5 h-3.5 w-3.5" />CSV
          </Button>
          <Button type="button" variant="outline" size="sm" disabled={!rows.length} onClick={() => window.print()}>
            <FileText className="mr-1.5 h-3.5 w-3.5" />Save PDF
          </Button>
          <Button type="button" size="sm" variant={allPublished ? 'outline' : 'default'} disabled={!selectedRows.length || publishResults.isPending} onClick={() => publish(!allPublished)} className={allPublished ? '' : 'bg-[#059669] text-white hover:bg-[#047857]'}>
            {publishResults.isPending ? <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" /> : allPublished ? <Undo2 className="mr-1.5 h-3.5 w-3.5" /> : <Send className="mr-1.5 h-3.5 w-3.5" />}
            {allPublished ? `Withdraw ${hasExplicitSelection ? 'selected' : 'results'}` : `Publish ${hasExplicitSelection ? 'selected' : 'results'}`}
          </Button>
        </div>
      </div>

      <div className="flex items-center gap-2 rounded-lg border border-[#e4e2de] bg-[#fafaf8] px-3 py-2 text-xs text-[#66655f]">
        <Users className="h-4 w-4 text-[#059669]" />
        <span>{rows.length} enrolled student{rows.length === 1 ? '' : 's'} · {publishedCount} result{publishedCount === 1 ? '' : 's'} published{hasExplicitSelection ? ` · ${selectedStudentIDs.length} selected` : ''}</span>
      </div>

      {rows.length === 0 ? (
        <p className="rounded-lg border border-dashed border-[#e4e2de] py-10 text-center text-sm text-[#888780]">Enroll students in a course group to populate the academic register.</p>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-[#e4e2de]">
          <table className="w-full min-w-[960px] text-left text-xs">
            <thead className="bg-[#fafaf8] text-[#6b6b6b]">
              <tr className="border-b border-[#e4e2de]">
                <th className="w-10 px-3 py-3 text-center"><input type="checkbox" aria-label="Select all students" checked={rows.length > 0 && selectedStudentIDs.length === rows.length} onChange={toggleAll} /></th>
                <th className="px-4 py-3 font-semibold">Student</th>
                <th className="px-3 py-3 text-right font-semibold">Assignments</th>
                <th className="px-3 py-3 text-right font-semibold">Midterm</th>
                <th className="px-3 py-3 text-right font-semibold">Endterm</th>
                <th className="px-3 py-3 text-right font-semibold">Final</th>
                <th className="px-3 py-3 text-right font-semibold">Calculated</th>
                <th className="px-3 py-3 text-right font-semibold">Result</th>
                <th className="px-3 py-3 text-center font-semibold">State</th>
                <th className="px-3 py-3 text-right font-semibold print:hidden">Edit</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#f0eeeb] bg-white">
              {rows.map((row) => <GradeRow key={row.student_id} row={row} selected={selectedStudentIDs.includes(row.student_id)} onSelect={() => toggleStudent(row.student_id)} onEdit={() => setEditor(row)} />)}
            </tbody>
          </table>
        </div>
      )}

      <GradeEditor key={editor?.student_id ?? 'closed'} courseId={courseId} student={editor} onClose={() => setEditor(null)} />
    </div>
  )
}

function GradeRow({ row, selected, onSelect, onEdit }: { row: GradebookStudentDTO; selected: boolean; onSelect: () => void; onEdit: () => void }) {
  const { grade } = row
  const result = grade.final_grade
  const hasOverride = grade.override_grade != null
  return (
    <tr>
      <td className="px-3 py-3 text-center"><input type="checkbox" aria-label={`Select ${row.student_name}`} checked={selected} onChange={onSelect} /></td>
      <td className="px-4 py-3">
        <p className="font-semibold text-[#292925]">{row.student_name}</p>
        <p className="mt-0.5 text-[#888780]">{row.student_email}</p>
      </td>
      <td className="px-3 py-3 text-right tabular-nums">{percent(grade.assignments_percent)}</td>
      <td className="px-3 py-3 text-right tabular-nums">{percent(grade.midterm_percent)}</td>
      <td className="px-3 py-3 text-right tabular-nums">{percent(grade.endterm_percent)}</td>
      <td className="px-3 py-3 text-right tabular-nums">{percent(grade.final_exam_percent)}</td>
      <td className="px-3 py-3 text-right font-medium tabular-nums">{grade.is_final_grade_ready ? percent(grade.calculated_final_grade) : 'In progress'}</td>
      <td className="px-3 py-3 text-right font-bold text-[#1a1a1a] tabular-nums">{result !== undefined ? percent(result) : '—'}{hasOverride && <span className="ml-1 text-[10px] font-medium text-amber-700">adj.</span>}</td>
      <td className="px-3 py-3 text-center"><span className={grade.is_published ? 'rounded-full bg-emerald-50 px-2 py-1 text-[10px] font-semibold text-emerald-700' : 'rounded-full bg-amber-50 px-2 py-1 text-[10px] font-semibold text-amber-700'}>{grade.is_published ? 'Published' : 'Draft'}</span></td>
      <td className="px-3 py-3 text-right print:hidden"><Button type="button" variant="ghost" size="sm" className="h-7 px-2 text-xs" onClick={onEdit}><Pencil className="mr-1 h-3.5 w-3.5" />Adjust</Button></td>
    </tr>
  )
}

function GradeEditor({ courseId, student, onClose }: { courseId: string; student: GradebookStudentDTO | null; onClose: () => void }) {
  const mutation = useUpdateCourseGrade(courseId)
  const [gradeValue, setGradeValue] = useState('')
  const [note, setNote] = useState('')
  const open = Boolean(student)
  const grade = student?.grade
  const initial = useMemo(() => ({ value: grade?.override_grade?.toString() ?? '', note: grade?.override_note ?? '' }), [grade?.override_grade, grade?.override_note])

  const close = () => {
    setGradeValue('')
    setNote('')
    onClose()
  }

  const save = async () => {
    if (!student) return
    const value = gradeValue.trim() === '' ? undefined : Number(gradeValue)
    if (value !== undefined && (!Number.isFinite(value) || value < 0 || value > 100)) {
      toast.error('The adjusted result must be between 0 and 100')
      return
    }
    if (value !== undefined && !note.trim()) {
      toast.error('Add a reason for the manual adjustment')
      return
    }
    try {
      await mutation.mutateAsync({ studentID: student.student_id, input: { override_grade: value, override_note: note } })
      toast.success('Grade adjustment saved')
      close()
    } catch {
      toast.error('Could not save the adjustment')
    }
  }

  const clear = async () => {
    if (!student) return
    try {
      await mutation.mutateAsync({ studentID: student.student_id, input: { clear_override: true } })
      toast.success('Grade adjustment removed')
      close()
    } catch {
      toast.error('Could not remove the adjustment')
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => { if (!next) close() }}>
      <DialogContent className="max-w-md">
        <DialogHeader><DialogTitle>Adjust final result</DialogTitle></DialogHeader>
        <p className="text-sm text-[#66655f]">{student?.student_name} · calculated result: {grade?.is_final_grade_ready ? percent(grade.calculated_final_grade) : 'in progress'}</p>
        <div className="space-y-4 py-2">
          <div className="space-y-1.5"><Label htmlFor="override-grade">Adjusted result (%)</Label><Input id="override-grade" type="number" min="0" max="100" step="0.1" defaultValue={initial.value} onChange={(event) => setGradeValue(event.target.value)} placeholder="Leave empty to use calculated result" /></div>
          <div className="space-y-1.5"><Label htmlFor="override-note">Reason / note</Label><Textarea id="override-note" maxLength={2000} defaultValue={initial.note} onChange={(event) => setNote(event.target.value)} placeholder="Visible in the academic record" /></div>
        </div>
        <DialogFooter className="gap-2 sm:gap-0">
          {grade?.override_grade != null && <Button type="button" variant="outline" onClick={clear} disabled={mutation.isPending}>Remove adjustment</Button>}
          <Button type="button" variant="outline" onClick={close}>Cancel</Button>
          <Button type="button" className="bg-[#059669] text-white hover:bg-[#047857]" onClick={save} disabled={mutation.isPending}>{mutation.isPending && <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />}Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
