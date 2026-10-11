'use client'

import { useState, useEffect } from 'react'
import { GroupDTO, CreateGroupInput } from '@/types/group'
import { useMembers } from '@/lib/queries/members.queries'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { useI18n } from '@/lib/i18n'

interface GroupFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (input: CreateGroupInput) => Promise<void>
  defaultValues?: GroupDTO
  isPending?: boolean
}

export function GroupForm({ open, onOpenChange, onSubmit, defaultValues, isPending }: GroupFormProps) {
  const { t } = useI18n()
  const [name, setName] = useState('')
  const [teacherSearch, setTeacherSearch] = useState('')
  const [selectedTeacherId, setSelectedTeacherId] = useState('')
  const [selectedTeacherName, setSelectedTeacherName] = useState('')
  const [showDropdown, setShowDropdown] = useState(false)

  const { data: membersData } = useMembers({ role: 'teacher', page: 1, page_size: 100 })
  const teachers = membersData?.data ?? []

  const filtered = teacherSearch.trim()
    ? teachers.filter((t) =>
        t.name.toLowerCase().includes(teacherSearch.toLowerCase()) ||
        t.email.toLowerCase().includes(teacherSearch.toLowerCase())
      )
    : teachers

  useEffect(() => {
    if (open) {
      if (defaultValues) {
        setName(defaultValues.name)
        if (defaultValues.teacher_id) {
          const t = teachers.find((m) => m.user_id === defaultValues.teacher_id)
          setSelectedTeacherId(defaultValues.teacher_id)
          setSelectedTeacherName(t ? `${t.name} (${t.email})` : defaultValues.teacher_id)
          setTeacherSearch(t ? `${t.name} (${t.email})` : defaultValues.teacher_id)
        } else {
          setSelectedTeacherId('')
          setSelectedTeacherName('')
          setTeacherSearch('')
        }
      } else {
        setName('')
        setSelectedTeacherId('')
        setSelectedTeacherName('')
        setTeacherSearch('')
      }
      setShowDropdown(false)
    }
  }, [open, defaultValues])

  const handleSelectTeacher = (userId: string, display: string) => {
    setSelectedTeacherId(userId)
    setSelectedTeacherName(display)
    setTeacherSearch(display)
    setShowDropdown(false)
  }

  const handleClearTeacher = () => {
    setSelectedTeacherId('')
    setSelectedTeacherName('')
    setTeacherSearch('')
  }

  const handleSubmit = async () => {
    if (!name.trim()) return
    await onSubmit({
      name: name.trim(),
      teacher_id: selectedTeacherId || undefined,
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{defaultValues ? t('editGroup') : t('createGroup')}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-1.5">
            <Label>{t('groupName')}</Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Group A"
              className="border-[#e4e2de]"
            />
          </div>
          <div className="space-y-1.5">
            <Label>
              {t('teacher')} <span className="text-[#9b9b9b] font-normal">({t('optional')})</span>
            </Label>
            <div className="relative">
              <Input
                value={teacherSearch}
                onChange={(e) => {
                  setTeacherSearch(e.target.value)
                  if (selectedTeacherName && e.target.value !== selectedTeacherName) {
                    setSelectedTeacherId('')
                    setSelectedTeacherName('')
                  }
                  setShowDropdown(true)
                }}
                onFocus={() => setShowDropdown(true)}
                onBlur={() => setTimeout(() => setShowDropdown(false), 150)}
                placeholder="Search by name or email…"
                className="border-[#e4e2de]"
              />
              {selectedTeacherId && (
                <button
                  type="button"
                  onClick={handleClearTeacher}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-[#9b9b9b] hover:text-[#1a1a1a]"
                >
                  ✕
                </button>
              )}
              {showDropdown && filtered.length > 0 && (
                <div className="absolute z-50 mt-1 w-full rounded-lg border border-[#e4e2de] bg-white shadow-md max-h-48 overflow-y-auto">
                  {filtered.map((t) => (
                    <button
                      key={t.user_id}
                      type="button"
                      onMouseDown={() => handleSelectTeacher(t.user_id, `${t.name} (${t.email})`)}
                      className="w-full flex flex-col items-start px-3 py-2 hover:bg-[#f0eeeb] transition-colors text-left"
                    >
                      <span className="text-sm font-medium text-[#1a1a1a]">{t.name}</span>
                      <span className="text-xs text-[#9b9b9b]">{t.email}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t('cancel')}</Button>
          <Button
            onClick={handleSubmit}
            disabled={isPending || !name.trim()}
            className="bg-[#059669] hover:bg-[#047857] text-white"
          >
            {isPending ? t('saving') : defaultValues ? t('saveChanges') : t('createGroup')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
