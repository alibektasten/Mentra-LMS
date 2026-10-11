'use client'

import { Award, BookOpen, CheckCircle2, Copy, ExternalLink, Sparkles } from 'lucide-react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { useI18n } from '@/lib/i18n'
import { useCertificates } from '@/lib/queries/certificates.queries'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { API_URL } from '@/lib/api/client'

export default function StudentCertificatesPage() {
  const { org } = useParams<{ org: string }>()
  const { t } = useI18n()
  const { data: certificates = [], isLoading } = useCertificates()

  return (
    <div className="max-w-3xl">
      {/* Page header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight text-[#1a1a1a]">{t('certificates')}</h1>
        <p className="mt-1 text-sm text-[#9b9b9b]">{t('certificatesDescription')}</p>
      </div>

      {isLoading ? (
        <div className="space-y-3"><Skeleton className="h-28 rounded-xl" /><Skeleton className="h-28 rounded-xl" /></div>
      ) : certificates.length > 0 ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {certificates.map((certificate) => {
            const verificationURL = `${API_URL}/certificates/${certificate.verification_code}`
            return (
              <article key={certificate.id} className="overflow-hidden rounded-xl border border-emerald-200 bg-gradient-to-br from-emerald-50 to-white p-5 shadow-sm dark:border-emerald-900 dark:from-emerald-950/30 dark:to-card">
                <div className="flex items-start justify-between gap-3">
                  <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-600 text-white"><Award className="h-5 w-5" /></span>
                  <span className="inline-flex items-center gap-1 text-xs font-medium text-emerald-700 dark:text-emerald-300"><CheckCircle2 className="h-3.5 w-3.5" />Verified</span>
                </div>
                <p className="mt-5 text-xs font-semibold uppercase tracking-wider text-emerald-700 dark:text-emerald-300">Certificate of completion</p>
                <h2 className="mt-1 text-lg font-bold text-foreground">{certificate.course_title}</h2>
                <p className="mt-2 text-xs text-muted-foreground">Issued {new Date(certificate.issued_at).toLocaleDateString()}</p>
                <p className="mt-3 font-mono text-[11px] text-muted-foreground">{certificate.verification_code}</p>
                <div className="mt-4 flex gap-2">
                  <Button size="sm" variant="outline" className="h-8 text-xs" onClick={() => { navigator.clipboard.writeText(verificationURL); toast.success('Verification link copied') }}><Copy className="mr-1 h-3 w-3" />Copy link</Button>
                  <a href={verificationURL} target="_blank" rel="noreferrer" className="inline-flex h-8 items-center rounded-md px-2 text-xs font-medium text-emerald-700 hover:bg-emerald-100 dark:text-emerald-300 dark:hover:bg-emerald-950"><ExternalLink className="mr-1 h-3 w-3" />Verify</a>
                </div>
              </article>
            )
          })}
        </div>
      ) : (
      <div className="border border-[#e8e8e6] rounded-lg overflow-hidden">
        <div className="flex flex-col items-center justify-center py-20 px-8 text-center">
          <Award className="h-10 w-10 text-[#c9c9c9] mb-4" />
          <h2 className="text-sm font-semibold text-[#1a1a1a]">{t('noCertificatesYet')}</h2>
          <p className="mt-1.5 text-sm text-[#9b9b9b] max-w-xs leading-relaxed">
            Complete every lesson in a published course to receive a verifiable certificate automatically.
          </p>

          {/* Steps */}
          <div className="mt-8 flex flex-col gap-2 w-full max-w-xs">
            {[
              { icon: BookOpen, label: 'Complete all lessons in a course' },
              { icon: Sparkles, label: 'Pass quizzes with a qualifying score' },
              { icon: Award, label: 'Receive your certificate' },
            ].map(({ icon: Icon, label }, i) => (
              <div key={i} className="flex items-center gap-3 text-left px-3 py-2 rounded-md hover:bg-[#f7f7f5] transition-colors">
                <span className="text-[11px] font-semibold text-[#9b9b9b] w-4 shrink-0 text-center">
                  {i + 1}
                </span>
                <Icon className="h-3.5 w-3.5 text-[#9b9b9b] shrink-0" />
                <p className="text-xs text-[#6b6b6b]">{label}</p>
              </div>
            ))}
          </div>

          <Link
            href={`/${org}/student/courses`}
            className="mt-8 inline-flex items-center gap-2 rounded-md bg-[#1a1a1a] px-4 py-2 text-xs font-semibold text-white hover:bg-[#2a2a2a] transition-colors"
          >
            <BookOpen className="h-3.5 w-3.5" />
            {t('browseCourses')}
          </Link>
        </div>
      </div>
      )}
    </div>
  )
}
