'use client'

import { PageHeader } from '@/components/shared/PageHeader'
import { MemberTable } from './MemberTable'
import { InviteMemberDialog } from './InviteMemberDialog'
import { ImportCSVDialog } from './ImportCSVDialog'
import { useI18n } from '@/lib/i18n'

export default function AdminMembersPage() {
  const { t } = useI18n()
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <PageHeader title={t('members')} description={t('manageMembers')} />
        <div className="flex items-center gap-2">
          <ImportCSVDialog />
          <InviteMemberDialog />
        </div>
      </div>
      <MemberTable />
    </div>
  )
}
