'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { MoreHorizontal, Trash2 } from 'lucide-react'
import { useMembers, useRemoveMember, useUpdateMemberRole } from '@/lib/queries/members.queries'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { MemberDTO, MemberRole } from '@/types/member'
import { useI18n } from '@/lib/i18n'

const ROLE_COLORS: Record<MemberRole, string> = {
  admin: 'bg-purple-50 text-purple-700 border-purple-200',
  teacher: 'bg-blue-50 text-blue-700 border-blue-200',
  student: 'bg-emerald-50 text-emerald-700 border-emerald-200',
}

function RoleFilter({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const { t } = useI18n()
  return (
    <div className="flex gap-1.5">
      {(['', 'admin', 'teacher', 'student'] as const).map((r) => (
        <button
          key={r}
          onClick={() => onChange(r)}
          className={`text-xs px-3 py-1 rounded-full border transition-colors ${
            value === r
              ? 'border-[#059669] bg-[#059669] text-white'
              : 'border-[#e4e2de] bg-white text-[#6b6b6b] hover:border-[#059669]'
          }`}
        >
          {r === '' ? t('all') : r === 'student' ? t('students') : r === 'teacher' ? t('teacher') : 'Admin'}
        </button>
      ))}
    </div>
  )
}

function MemberRow({ member }: { member: MemberDTO }) {
  const { t } = useI18n()
  const remove = useRemoveMember()
  const updateRole = useUpdateMemberRole()

  const handleRoleChange = async (newRole: string) => {
    try {
      await updateRole.mutateAsync({ id: member.id, input: { role: newRole as MemberRole } })
      toast.success(t('roleUpdated'))
    } catch {
      toast.error(t('failedToUpdateRole'))
    }
  }

  const handleRemove = async () => {
    if (!confirm(`Remove ${member.name} from the organization?`)) return
    try {
      await remove.mutateAsync(member.id)
      toast.success(t('memberRemoved'))
    } catch {
      toast.error(t('failedToRemoveMember'))
    }
  }

  return (
    <div className="flex items-center justify-between px-4 py-3 hover:bg-[#fafaf9] rounded-lg transition-colors">
      <div className="flex items-center gap-3 min-w-0">
        <div className="h-8 w-8 rounded-full bg-[#f0eeeb] flex items-center justify-center shrink-0">
          <span className="text-xs font-semibold text-[#6b6b6b]">
            {member.name.charAt(0).toUpperCase()}
          </span>
        </div>
        <div className="min-w-0">
          <p className="text-sm font-medium text-[#1a1a1a] truncate">{member.name}</p>
          <p className="text-xs text-[#9b9b9b] truncate">{member.email}</p>
        </div>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <span className={`text-[10px] font-semibold uppercase tracking-wider px-2 py-1 rounded-full border ${ROLE_COLORS[member.role]}`}>
          {member.role}
        </span>
        <span className="text-xs text-[#9b9b9b] hidden sm:block">
          {new Date(member.joined_at).toLocaleDateString()}
        </span>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-7 w-7">
              <MoreHorizontal className="h-3.5 w-3.5" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-44">
            <div className="px-2 py-1.5">
              <p className="text-xs text-[#9b9b9b] mb-1">{t('changeRole')}</p>
              <Select value={member.role} onValueChange={handleRoleChange}>
                <SelectTrigger className="h-7 text-xs border-[#e4e2de]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="student">Student</SelectItem>
                  <SelectItem value="teacher">Teacher</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="text-red-600 focus:text-red-600 focus:bg-red-50 text-xs cursor-pointer"
              onClick={handleRemove}
            >
              <Trash2 className="h-3.5 w-3.5 mr-2" />
              {t('removeMember')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}

export function MemberTable() {
  const { t } = useI18n()
  const [roleFilter, setRoleFilter] = useState('')
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, isLoading } = useMembers({ page, page_size: pageSize, role: roleFilter || undefined })
  const members = data?.data ?? []
  const total = data?.meta?.total ?? 0
  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <RoleFilter value={roleFilter} onChange={(v) => { setRoleFilter(v); setPage(1) }} />
        <p className="text-xs text-[#9b9b9b]">{total} member{total !== 1 ? 's' : ''}</p>
      </div>

      <div className="rounded-xl border border-[#e4e2de] bg-white divide-y divide-[#f0eeeb] shadow-sm">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-2 bg-[#fafaf9] rounded-t-xl">
          <span className="text-xs font-semibold text-[#9b9b9b] uppercase tracking-wide">{t('member')}</span>
          <div className="flex items-center gap-11">
            <span className="text-xs font-semibold text-[#9b9b9b] uppercase tracking-wide">{t('role')}</span>
            <span className="text-xs font-semibold text-[#9b9b9b] uppercase tracking-wide hidden sm:block">{t('joined')}</span>
            <span className="w-7" />
          </div>
        </div>

        {isLoading ? (
          Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="px-4 py-3">
              <Skeleton className="h-8 w-full" />
            </div>
          ))
        ) : members.length === 0 ? (
          <div className="px-4 py-10 text-center text-sm text-[#9b9b9b]">{t('noMembersFound')}</div>
        ) : (
          members.map((m) => <MemberRow key={m.id} member={m} />)
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
            className="border-[#e4e2de] text-xs"
          >
            {t('previous')}
          </Button>
          <span className="text-xs text-[#9b9b9b]">{t('page')} {page} {t('of')} {totalPages}</span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
            className="border-[#e4e2de] text-xs"
          >
            {t('next')}
          </Button>
        </div>
      )}
    </div>
  )
}
