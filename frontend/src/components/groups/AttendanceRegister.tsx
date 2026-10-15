'use client'

import { useEffect, useMemo, useState } from 'react'
import { Check, Loader2, Save, Users } from 'lucide-react'
import { toast } from 'sonner'
import { useAttendance, useSaveAttendance } from '@/lib/queries/groups.queries'
import type { AttendanceStatus } from '@/types/group'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

const today = () => new Date().toISOString().slice(0, 10)
const labels: Record<AttendanceStatus, string> = { present: 'Present', late: 'Late', absent: 'Absent' }

export function AttendanceRegister({ courseId, groupId }: { courseId: string; groupId: string }) {
  const [date, setDate] = useState(today)
  const { data, isLoading, isError } = useAttendance(courseId, groupId, date)
  const saveAttendance = useSaveAttendance(courseId, groupId, date)
  const initialStatuses = useMemo(() => Object.fromEntries((data?.students ?? []).map((student) => [student.student_id, (student.status || 'present') as AttendanceStatus])), [data])
  const [statuses, setStatuses] = useState<Record<string, AttendanceStatus>>({})

  useEffect(() => setStatuses(initialStatuses), [initialStatuses])

  const markAllPresent = () => {
    setStatuses(Object.fromEntries((data?.students ?? []).map((student) => [student.student_id, 'present'])))
  }

  const save = async () => {
    const students = data?.students ?? []
    try {
      await saveAttendance.mutateAsync({ date, records: students.map((student) => ({ student_id: student.student_id, status: statuses[student.student_id] ?? 'present' })) })
      toast.success('Attendance saved')
    } catch {
      toast.error('Could not save attendance')
    }
  }

  if (isLoading) return <div className="space-y-2">{[1, 2, 3].map((item) => <Skeleton key={item} className="h-12" />)}</div>
  if (isError) return <p className="py-3 text-sm text-destructive">Could not load attendance.</p>
  const students = data?.students ?? []

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-[#e4e2de] bg-[#fafaf8] p-3">
        <div><p className="text-sm font-semibold text-[#1a1a1a]">Attendance register</p><p className="mt-0.5 text-xs text-[#85847e]">Mark attendance for every enrolled student.</p></div>
        <input aria-label="Attendance date" type="date" value={date} onChange={(event) => setDate(event.target.value)} className="h-9 rounded-md border border-[#deddd8] bg-white px-2 text-sm" />
      </div>
      {students.length === 0 ? <p className="py-5 text-center text-sm text-[#888780]">No students are enrolled in this group.</p> : <>
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span className="flex items-center gap-1.5 text-xs text-[#6b6b6b]"><Users className="h-3.5 w-3.5" />{students.length} student{students.length === 1 ? '' : 's'}</span>
          <Button type="button" variant="outline" size="sm" onClick={markAllPresent}><Check className="mr-1.5 h-3.5 w-3.5" />Mark all present</Button>
        </div>
        <div className="divide-y divide-[#f0eeeb] overflow-hidden rounded-xl border border-[#e4e2de]">
          {students.map((student) => <div key={student.student_id} className="flex flex-col gap-2 bg-white px-4 py-3 sm:flex-row sm:items-center sm:justify-between"><div><p className="text-sm font-medium text-[#1a1a1a]">{student.name}</p><p className="text-xs text-[#9b9b9b]">{student.email}</p></div><div className="flex overflow-hidden rounded-lg border border-[#e4e2de]">{(Object.keys(labels) as AttendanceStatus[]).map((status) => <button key={status} type="button" onClick={() => setStatuses((current) => ({ ...current, [student.student_id]: status }))} className={`px-3 py-1.5 text-xs font-medium transition-colors ${statuses[student.student_id] === status ? status === 'present' ? 'bg-emerald-600 text-white' : status === 'late' ? 'bg-amber-500 text-white' : 'bg-rose-600 text-white' : 'bg-white text-[#6b6b6b] hover:bg-[#f7f6f3]'}`}>{labels[status]}</button>)}</div></div>)}
        </div>
        <div className="flex justify-end"><Button type="button" className="bg-[#059669] text-white hover:bg-[#047857]" onClick={save} disabled={saveAttendance.isPending}>{saveAttendance.isPending ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <Save className="mr-1.5 h-4 w-4" />}Save attendance</Button></div>
      </>}
    </div>
  )
}
