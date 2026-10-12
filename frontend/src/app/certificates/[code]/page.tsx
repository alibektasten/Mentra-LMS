'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import QRCode from 'qrcode'
import {
  ArrowLeft,
  Award,
  BadgeCheck,
  Building2,
  CalendarDays,
  Download,
  GraduationCap,
  Loader2,
  ShieldCheck,
  UserRound,
} from 'lucide-react'
import { verifyCertificate } from '@/lib/api/certificates'

function formatDate(value: string) {
  return new Intl.DateTimeFormat('en', {
    day: 'numeric', month: 'long', year: 'numeric',
  }).format(new Date(value))
}

export default function CertificateVerificationPage() {
  const { code } = useParams<{ code: string }>()
  const certificateCode = Array.isArray(code) ? code[0] : code
  const certificateRef = useRef<HTMLElement>(null)
  const [qrCode, setQrCode] = useState('')
  const [isDownloading, setIsDownloading] = useState(false)
  const { data: certificate, isLoading, isError } = useQuery({
    queryKey: ['certificate-verification', certificateCode],
    queryFn: () => verifyCertificate(certificateCode),
    enabled: Boolean(certificateCode),
    retry: false,
  })

  useEffect(() => {
    if (!certificateCode) return
    QRCode.toDataURL(`${window.location.origin}/certificates/${certificateCode}`, {
      width: 220,
      margin: 1,
      color: { dark: '#07553d', light: '#ffffff' },
    }).then(setQrCode).catch(() => setQrCode(''))
  }, [certificateCode])

  async function downloadPDF() {
    if (!certificateRef.current || !certificate) return
    setIsDownloading(true)
    const node = certificateRef.current
    const originalStyle = {
      width: node.style.width,
      minWidth: node.style.minWidth,
      maxWidth: node.style.maxWidth,
    }
    try {
      const [{ default: html2canvas }, { jsPDF }] = await Promise.all([
        import('html2canvas'), import('jspdf'),
      ])
      // Render at a fixed desktop width so the exported certificate always fills landscape A4.
      node.style.width = '1120px'
      node.style.minWidth = '1120px'
      node.style.maxWidth = 'none'
      const canvas = await html2canvas(node, {
        backgroundColor: '#ffffff', scale: 2, useCORS: true, logging: false,
      })
      const image = canvas.toDataURL('image/png')
      const document = new jsPDF({ orientation: 'landscape', unit: 'mm', format: 'a4' })
      const pageWidth = document.internal.pageSize.getWidth()
      const pageHeight = document.internal.pageSize.getHeight()
      const scale = Math.min(pageWidth / canvas.width, pageHeight / canvas.height)
      const width = canvas.width * scale
      const height = canvas.height * scale
      document.addImage(image, 'PNG', (pageWidth - width) / 2, (pageHeight - height) / 2, width, height)
      document.save(`Mentra-certificate-${certificate.verification_code}.pdf`)
    } finally {
      node.style.width = originalStyle.width
      node.style.minWidth = originalStyle.minWidth
      node.style.maxWidth = originalStyle.maxWidth
      setIsDownloading(false)
    }
  }

  if (isLoading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#f3f7f4] px-6 text-[#12352a]">
        <Loader2 className="h-8 w-8 animate-spin text-emerald-600" aria-label="Loading certificate" />
      </main>
    )
  }

  if (isError || !certificate) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#f3f7f4] px-6 py-12">
        <section className="w-full max-w-lg rounded-3xl border border-rose-200 bg-white p-9 text-center shadow-xl shadow-emerald-950/5">
          <span className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-rose-50 text-rose-600"><ShieldCheck className="h-7 w-7" /></span>
          <h1 className="mt-5 text-2xl font-bold text-slate-900">Certificate not found</h1>
          <p className="mt-2 text-sm leading-6 text-slate-600">This verification link is invalid or the certificate is no longer available.</p>
          <Link href="/login" className="mt-7 inline-flex items-center gap-2 rounded-xl bg-emerald-700 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-emerald-800"><ArrowLeft className="h-4 w-4" />Back to Mentra LMS</Link>
        </section>
      </main>
    )
  }

  return (
    <main className="min-h-screen overflow-hidden bg-[#f2f7f3] px-4 py-8 text-slate-900 sm:px-8 sm:py-12">
      <div className="mx-auto max-w-6xl">
        <header className="mb-7 flex flex-wrap items-center justify-between gap-3 px-1 sm:mb-9">
          <Link href="/login" className="inline-flex items-center gap-2 text-sm font-semibold text-slate-600 transition hover:text-emerald-700"><ArrowLeft className="h-4 w-4" />Mentra LMS</Link>
          <div className="flex items-center gap-3">
            <span className="inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-800"><BadgeCheck className="h-4 w-4" />Verified credential</span>
            <button onClick={downloadPDF} disabled={isDownloading} className="inline-flex items-center gap-2 rounded-full bg-emerald-700 px-3.5 py-1.5 text-xs font-bold text-white transition hover:bg-emerald-800 disabled:cursor-wait disabled:opacity-70"><Download className="h-4 w-4" />{isDownloading ? 'Preparing PDF…' : 'Download PDF'}</button>
          </div>
        </header>

        <article ref={certificateRef} className="relative overflow-hidden rounded-[30px] border border-emerald-950/10 bg-white p-3 shadow-[0_28px_80px_-34px_rgba(15,80,54,0.35)] sm:p-5">
          <div className="absolute -right-24 -top-24 h-64 w-64 rounded-full bg-emerald-100/70 blur-3xl" />
          <div className="absolute -bottom-28 -left-24 h-64 w-64 rounded-full bg-amber-100/60 blur-3xl" />
          <div className="relative border border-emerald-800/25 px-6 py-9 sm:px-12 sm:py-12 lg:px-20 lg:py-16">
            <div className="absolute left-3 top-3 h-7 w-7 border-l-2 border-t-2 border-emerald-700 sm:left-5 sm:top-5" />
            <div className="absolute right-3 top-3 h-7 w-7 border-r-2 border-t-2 border-emerald-700 sm:right-5 sm:top-5" />
            <div className="absolute bottom-3 left-3 h-7 w-7 border-b-2 border-l-2 border-emerald-700 sm:bottom-5 sm:left-5" />
            <div className="absolute bottom-3 right-3 h-7 w-7 border-b-2 border-r-2 border-emerald-700 sm:bottom-5 sm:right-5" />

            <div className="flex flex-col items-center text-center">
              <div className="flex items-center gap-3 text-emerald-800">
                <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-emerald-700 text-lg font-black text-white shadow-lg shadow-emerald-700/20">M</span>
                <span className="text-left text-xs font-bold uppercase leading-4 tracking-[0.24em]">Mentra<br />LMS</span>
              </div>
              <div className="mt-8 flex h-14 w-14 items-center justify-center rounded-full border border-amber-300 bg-amber-50 text-amber-700 shadow-sm"><Award className="h-7 w-7" /></div>
              <p className="mt-5 text-xs font-bold uppercase tracking-[0.32em] text-emerald-700">Certificate of completion</p>
              <h1 className="mt-3 font-serif text-3xl font-semibold tracking-tight text-slate-900 sm:text-5xl">{certificate.student_name}</h1>
              <p className="mt-5 max-w-2xl text-sm leading-7 text-slate-600 sm:text-base">has successfully completed the learning requirements for</p>
              <h2 className="mt-2 max-w-3xl text-2xl font-bold leading-tight text-emerald-800 sm:text-4xl">{certificate.course_title}</h2>
              <p className="mt-4 text-sm font-medium text-slate-600">Issued by {certificate.organization_name}</p>
            </div>

            <div className="mt-10 grid gap-5 border-y border-slate-200 py-6 sm:grid-cols-3 sm:gap-7">
              <CertificateDetail icon={Building2} label="University" value={certificate.organization_name} />
              <CertificateDetail icon={UserRound} label="Instructor" value={certificate.teacher_name} />
              <CertificateDetail icon={CalendarDays} label="Issue date" value={formatDate(certificate.issued_at)} />
            </div>

            <div className="mt-8 flex flex-col items-center justify-between gap-8 sm:flex-row sm:items-end">
              <div className="text-center sm:text-left">
                <div className="mx-auto h-px w-52 bg-slate-400 sm:mx-0" />
                <p className="mt-2 font-serif text-xl italic text-slate-700">Mentra Academic Office</p>
                <p className="mt-1 text-xs font-semibold uppercase tracking-wider text-slate-500">Digital signature</p>
              </div>
              <div className="flex items-center gap-3 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-left text-emerald-900">
                {qrCode ? <img src={qrCode} alt="QR code to verify this certificate" className="h-12 w-12 rounded-md bg-white p-0.5" /> : <ShieldCheck className="h-8 w-8 shrink-0 text-emerald-700" />}
                <div><p className="text-xs font-bold uppercase tracking-wide">Authenticity confirmed</p><p className="mt-0.5 font-mono text-[11px] text-emerald-800">{certificate.verification_code}</p><p className="mt-0.5 text-[10px] text-emerald-700">Scan to verify online</p></div>
              </div>
            </div>
          </div>
        </article>

        <p className="mt-6 flex items-center justify-center gap-2 text-center text-xs text-slate-500"><GraduationCap className="h-4 w-4 text-emerald-700" />This credential was verified securely by Mentra LMS.</p>
      </div>
    </main>
  )
}

function CertificateDetail({ icon: Icon, label, value }: { icon: typeof Building2; label: string; value: string }) {
  return (
    <div className="flex items-start gap-3 text-left">
      <span className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-emerald-700"><Icon className="h-4 w-4" /></span>
      <div><p className="text-[11px] font-bold uppercase tracking-wider text-slate-500">{label}</p><p className="mt-1 text-sm font-semibold leading-5 text-slate-800">{value}</p></div>
    </div>
  )
}
