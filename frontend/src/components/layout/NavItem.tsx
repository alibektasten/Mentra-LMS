'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils/cn'

interface NavItemProps {
  href: string
  label: string
  icon: LucideIcon
}

export function NavItem({ href, label, icon: Icon }: NavItemProps) {
  const pathname = usePathname()
  const isActive = pathname === href || pathname.startsWith(href + '/')

  return (
    <Link
      href={href}
      data-active={isActive}
      className={cn(
        'group flex items-center gap-2.5 rounded-xl px-3 py-2 text-sm transition-all duration-200',
        isActive
          ? 'bg-[#e9f7f0] text-[#075e46] font-semibold shadow-[inset_0_0_0_1px_rgba(16,185,129,0.09)]'
          : 'text-[#6b6b67] hover:bg-[#f2f1ed] hover:text-[#282824] font-medium'
      )}
    >
      <Icon className={cn('h-4 w-4 shrink-0 transition-colors', isActive ? 'text-[#07855f]' : 'text-[#a09f99] group-hover:text-[#575750]')} />
      {label}
    </Link>
  )
}
