'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { Plus, Users, Pencil, Trash2, Link2, Link2Off, ChevronDown, Calendar } from 'lucide-react'
import {
  useOrgGroups,
  useCreateOrgGroup,
  useUpdateOrgGroup,
  useDeleteOrgGroup,
  useAssignGroupToCourse,
  useUnassignGroupFromCourse,
} from '@/lib/queries/groups.queries'
import { useCourses } from '@/lib/queries/courses.queries'
import { GroupDTO, CreateGroupInput } from '@/types/group'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { GroupForm } from '@/components/groups/GroupForm'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { GroupMembers } from '@/components/groups/GroupMembers'
import { GroupScheduleList } from '@/components/groups/GroupScheduleList'
import { useI18n } from '@/lib/i18n'

function AssignCourseDialog({
  group,
  open,
  onOpenChange,
}: {
  group: GroupDTO
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useI18n()
  const [courseId, setCourseId] = useState(group.course_id ?? '')
  const { data: coursesData } = useCourses({ page: 1, page_size: 100 })
  const assign = useAssignGroupToCourse()

  const courses = coursesData?.data ?? []

  const handleAssign = async () => {
    if (!courseId) return
    try {
      await assign.mutateAsync({ groupId: group.id, input: { course_id: courseId } })
      toast.success(t('groupAssigned'))
      onOpenChange(false)
    } catch {
      toast.error(t('failedToAssignGroup'))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Assign "{group.name}" to Course</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <Select value={courseId} onValueChange={setCourseId}>
            <SelectTrigger className="border-[#e4e2de]">
              <SelectValue placeholder={t('selectCourse')} />
            </SelectTrigger>
            <SelectContent>
              {courses.map((c) => (
                <SelectItem key={c.id} value={c.id}>
                  {c.title}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t('cancel')}</Button>
          <Button
            disabled={!courseId || assign.isPending}
            onClick={handleAssign}
            className="bg-[#059669] hover:bg-[#047857] text-white"
          >
            {assign.isPending ? t('assigning') : t('assign')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export default function AdminGroupsPage() {
  const { t } = useI18n()
  const { data: groups, isLoading } = useOrgGroups()
  const createGroup = useCreateOrgGroup()
  const updateGroup = useUpdateOrgGroup()
  const deleteGroup = useDeleteOrgGroup()
  const unassign = useUnassignGroupFromCourse()

  const { data: coursesData } = useCourses({ page: 1, page_size: 100 })
  const courseMap = new Map((coursesData?.data ?? []).map((c) => [c.id, c]))

  const [createOpen, setCreateOpen] = useState(false)
  const [editingGroup, setEditingGroup] = useState<GroupDTO | undefined>()
  const [assigningGroup, setAssigningGroup] = useState<GroupDTO | undefined>()
  const [expandedGroupId, setExpandedGroupId] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'members' | 'schedule'>('members')

  const handleCreate = async (input: CreateGroupInput) => {
    try {
      await createGroup.mutateAsync(input)
      toast.success(t('groupCreated'))
      setCreateOpen(false)
    } catch {
      toast.error(t('failedToCreateGroup'))
    }
  }

  const handleUpdate = async (input: CreateGroupInput) => {
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
    } catch {
      toast.error(t('failedToDeleteGroup'))
    }
  }

  const handleUnassign = async (groupId: string) => {
    try {
      await unassign.mutateAsync(groupId)
      toast.success(t('groupUnassigned'))
    } catch {
      toast.error(t('failedToUnassignGroup'))
    }
  }

  return (
    <div className="max-w-4xl py-2">
      <div className="mb-7 flex items-start justify-between">
        <div>
          <div className="flex items-center gap-2.5 mb-1.5">
            <Users className="h-5 w-5 text-[#6b6b6b]" />
            <h1 className="text-[1.4rem] font-bold tracking-tight text-[#1a1a1a]">{t('groupManagement')}</h1>
          </div>
          <p className="text-sm text-[#6b6b6b]">{t('groupManagementDescription')}</p>
        </div>
        <Button
          size="sm"
          onClick={() => setCreateOpen(true)}
          className="bg-[#059669] hover:bg-[#047857] text-white gap-1.5"
        >
          <Plus className="h-3.5 w-3.5" />
          {t('newGroup')}
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-16 rounded-xl" />)}
        </div>
      ) : !groups || groups.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center rounded-xl border border-dashed border-[#e4e2de]">
          <Users className="h-7 w-7 text-[#d4d2ce] mb-2" />
          <p className="text-sm text-[#9b9b9b]">{t('noGroupsToCreate')}</p>
        </div>
      ) : (
        <div className="space-y-2">
          {groups.map((group) => {
            const course = group.course_id ? courseMap.get(group.course_id) : undefined
            const isExpanded = expandedGroupId === group.id
            return (
              <div key={group.id} className="rounded-xl border border-[#e4e2de] bg-white shadow-sm overflow-hidden">
                {/* Header row */}
                <div
                  className="px-4 py-3 flex items-center gap-4 cursor-pointer hover:bg-[#fafaf9] transition-colors"
                  onClick={() => setExpandedGroupId(isExpanded ? null : group.id)}
                >
                  <div className="h-9 w-9 rounded-lg bg-[#f0eeeb] flex items-center justify-center shrink-0">
                    <Users className="h-4 w-4 text-[#6b6b6b]" />
                  </div>

                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-semibold text-[#1a1a1a] truncate">{group.name}</p>
                    <p className="text-xs mt-0.5">
                      {course ? (
                        <span className="text-emerald-600">{t('assignedCourse')}: {course.title}</span>
                      ) : (
                        <span className="text-amber-600">{t('notAssigned')}</span>
                      )}
                    </p>
                  </div>

                  <div className="flex items-center gap-1 shrink-0" onClick={(e) => e.stopPropagation()}>
                    {group.course_id ? (
                      <ConfirmDialog
                        trigger={
                          <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-[#9b9b9b] hover:text-amber-600" title={t('unassignFromCourse')}>
                            <Link2Off className="h-3.5 w-3.5" />
                          </Button>
                        }
                        title={t('unassignFromCourse')}
                        description={`"${group.name}" will be removed from its course. Members remain in the group.`}
                        confirmLabel="Unassign"
                        onConfirm={() => handleUnassign(group.id)}
                      />
                    ) : (
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-7 w-7 p-0 text-[#9b9b9b] hover:text-[#059669]"
                        title={t('assignToCourse')}
                        onClick={() => setAssigningGroup(group)}
                      >
                        <Link2 className="h-3.5 w-3.5" />
                      </Button>
                    )}
                    <Button
                      size="sm"
                      variant="ghost"
                      className="h-7 w-7 p-0 text-[#9b9b9b] hover:text-[#1a1a1a]"
                      onClick={() => setEditingGroup(group)}
                    >
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <ConfirmDialog
                      trigger={
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-[#9b9b9b] hover:text-destructive">
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      }
                      title={t('deleteGroup')}
                      description={`"${group.name}" and all its members and schedules will be deleted.`}
                      confirmLabel="Delete"
                      onConfirm={() => handleDelete(group.id)}
                      destructive
                    />
                    <ChevronDown className={`h-4 w-4 text-[#9b9b9b] ml-1 transition-transform ${isExpanded ? 'rotate-180' : ''}`} />
                  </div>
                </div>

                {/* Expanded panel */}
                {isExpanded && (
                  <div className="border-t border-[#e4e2de]">
                    <div className="flex border-b border-[#e4e2de]">
                      {(['members', 'schedule'] as const).map((tab) => (
                        <button
                          key={tab}
                          onClick={() => setActiveTab(tab)}
                          className={`flex items-center gap-1.5 px-5 py-2.5 text-xs font-medium transition-colors border-b-2 ${
                            activeTab === tab
                              ? 'border-[#059669] text-[#059669]'
                              : 'border-transparent text-[#9b9b9b] hover:text-[#1a1a1a]'
                          }`}
                        >
                          {tab === 'members' ? <Users className="h-3.5 w-3.5" /> : <Calendar className="h-3.5 w-3.5" />}
                          {tab === 'members' ? t('members') : t('schedule')}
                        </button>
                      ))}
                    </div>
                    <div className="p-5">
                      {activeTab === 'members' ? (
                        <GroupMembers groupId={group.id} />
                      ) : (
                        <GroupScheduleList groupId={group.id} />
                      )}
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      <GroupForm
        open={createOpen}
        onOpenChange={setCreateOpen}
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
      {assigningGroup && (
        <AssignCourseDialog
          group={assigningGroup}
          open={!!assigningGroup}
          onOpenChange={(o) => { if (!o) setAssigningGroup(undefined) }}
        />
      )}
    </div>
  )
}
