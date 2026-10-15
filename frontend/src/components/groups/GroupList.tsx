'use client'

import { useState } from 'react'
import { Plus, Users, ChevronDown, ChevronRight, Calendar, ClipboardCheck } from 'lucide-react'
import { toast } from 'sonner'
import { GroupDTO } from '@/types/group'
import { useGroups, useCreateGroup, useUpdateGroup, useDeleteGroup } from '@/lib/queries/groups.queries'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { GroupCard } from './GroupCard'
import { GroupForm } from './GroupForm'
import { GroupMembers } from './GroupMembers'
import { GroupScheduleList } from './GroupScheduleList'
import { AttendanceRegister } from './AttendanceRegister'
import { useI18n } from '@/lib/i18n'

interface GroupListProps {
  courseId: string
}

export function GroupList({ courseId }: GroupListProps) {
  const { t } = useI18n()
  const { data: groupsData, isLoading } = useGroups(courseId)
  const createGroup = useCreateGroup(courseId)
  const updateGroup = useUpdateGroup(courseId)
  const deleteGroup = useDeleteGroup(courseId)

  const [formOpen, setFormOpen] = useState(false)
  const [editingGroup, setEditingGroup] = useState<GroupDTO | undefined>()
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'members' | 'schedule' | 'attendance'>('members')

  const groups = Array.isArray(groupsData) ? groupsData : []
  const selectedGroup = groups.find((g) => g.id === selectedGroupId)

  const handleCreate = async (input: { name: string; teacher_id?: string }) => {
    try {
      await createGroup.mutateAsync(input)
      toast.success(t('groupCreated'))
      setFormOpen(false)
    } catch {
      toast.error(t('failedToCreateGroup'))
    }
  }

  const handleUpdate = async (input: { name: string; teacher_id?: string }) => {
    if (!editingGroup) return
    try {
      await updateGroup.mutateAsync({ id: editingGroup.id, input })
      toast.success(t('groupUpdated'))
      setEditingGroup(undefined)
    } catch {
      toast.error(t('failedToUpdateGroup'))
    }
  }

  const handleDelete = async (groupId: string) => {
    try {
      await deleteGroup.mutateAsync(groupId)
      toast.success(t('groupDeleted'))
      if (selectedGroupId === groupId) setSelectedGroupId(null)
    } catch {
      toast.error(t('failedToDeleteGroup'))
    }
  }

  if (isLoading) {
    return <div className="space-y-3">{[1, 2].map((i) => <Skeleton key={i} className="h-16 rounded-xl" />)}</div>
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-[#1a1a1a]">{t('courseGroups')}</h3>
        <Button size="sm" variant="outline" onClick={() => setFormOpen(true)}>
          <Plus className="h-4 w-4 mr-1.5" />
          {t('newGroup')}
        </Button>
      </div>

      {groups.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center rounded-xl border border-dashed border-[#e4e2de]">
          <Users className="h-7 w-7 text-[#d4d2ce] mb-2" />
          <p className="text-sm text-[#9b9b9b]">{t('noGroupsToCreate')}</p>
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {groups.map((g) => (
            <GroupCard
              key={g.id}
              group={g}
              courseId={courseId}
              onEdit={(group) => setEditingGroup(group)}
              onDelete={handleDelete}
              onSelect={(group) => setSelectedGroupId(group.id === selectedGroupId ? null : group.id)}
            />
          ))}
        </div>
      )}

      {selectedGroup && (
        <div className="rounded-xl border border-[#e4e2de] bg-white shadow-sm overflow-hidden">
          <div className="flex items-center justify-between px-5 py-3 bg-[#f7f6f3] border-b border-[#e4e2de]">
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4 text-[#6b6b6b]" />
              <span className="text-sm font-semibold text-[#1a1a1a]">{selectedGroup.name}</span>
            </div>
            <button
              onClick={() => setSelectedGroupId(null)}
              className="text-xs text-[#9b9b9b] hover:text-[#1a1a1a] transition-colors"
            >
              {t('close')}
            </button>
          </div>

          <div className="flex border-b border-[#e4e2de]">
            {(['members', 'schedule', 'attendance'] as const).map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`flex items-center gap-1.5 px-5 py-2.5 text-xs font-medium transition-colors border-b-2 ${
                  activeTab === tab
                    ? 'border-[#059669] text-[#059669]'
                    : 'border-transparent text-[#9b9b9b] hover:text-[#1a1a1a]'
                }`}
              >
                {tab === 'members' ? <Users className="h-3.5 w-3.5" /> : tab === 'schedule' ? <Calendar className="h-3.5 w-3.5" /> : <ClipboardCheck className="h-3.5 w-3.5" />}
                {tab === 'members' ? t('members') : tab === 'schedule' ? t('schedule') : 'Attendance'}
              </button>
            ))}
          </div>

          <div className="p-5">
            {activeTab === 'members' ? (
              <GroupMembers groupId={selectedGroup.id} courseId={courseId} />
            ) : activeTab === 'schedule' ? (
              <GroupScheduleList groupId={selectedGroup.id} courseId={courseId} />
            ) : <AttendanceRegister groupId={selectedGroup.id} courseId={courseId} />}
          </div>
        </div>
      )}

      <GroupForm
        open={formOpen}
        onOpenChange={setFormOpen}
        onSubmit={handleCreate}
        isPending={createGroup.isPending}
      />
      <GroupForm
        open={!!editingGroup}
        onOpenChange={(o) => { if (!o) setEditingGroup(undefined) }}
        onSubmit={handleUpdate}
        defaultValues={editingGroup}
        isPending={updateGroup.isPending}
      />
    </div>
  )
}
